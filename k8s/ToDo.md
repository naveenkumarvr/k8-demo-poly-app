# TODO

## PostgreSQL on Kubernetes - To-Do and Best Practices

### Goal
Deploy a single-node PostgreSQL instance on Kind for the `product-service` using Helm, and initialize it with the `schema.sql` and `seed.sql` files in a way that is safe, idempotent, and restart-resistant.

---

## 1. Deploy PostgreSQL with Helm

### Why
Helm lets us manage the Postgres lifecycle declaratively. The Bitnami chart is the most popular and gives us persistence, auth, and one-time init script support out of the box.

### What to do
1. Add the Bitnami repo:
   ```bash
   helm repo add bitnami https://charts.bitnami.com/bitnami
   helm repo update
   ```
2. Create `postgres-values.yaml` and install:
   ```bash
   helm upgrade --install postgres bitnami/postgresql \
     -n demo \
     -f postgres-values.yaml
   ```

### Recommended minimal values

```yaml
# postgres-values.yaml
auth:
  database: products
  username: productuser
  password: productpass
  postgresPassword: adminpass

primary:
  persistence:
    enabled: true
    size: 2Gi
    storageClass: standard

  initdb:
    scripts:
      01-schema.sql: |
        -- paste contents of product-service/database/schema.sql here
      02-seed.sql: |
        -- paste contents of product-service/database/seed.sql here
```

### Why these values
- `auth.database` and `auth.username` mirror the `docker-compose.yml` credentials that `product-service` expects.
- `persistence.enabled` ensures data survives pod restarts.
- `storageClass: standard` works with the Kind local dynamic provisioner.
- `primary.initdb.scripts` runs only when the data directory is empty, so schema and seed are applied once on first PVC creation.

---

## 2. Initialize the Database

### Option A: Helm initdb scripts (quick, demo only)
Use the `primary.initdb.scripts` block shown above. It is the simplest way because Bitnami entrypoint executes these files once on initial data directory creation.

### Why it works
- Scripts in `/docker-entrypoint-initdb.d` are executed by the Bitnami container entrypoint **only when the data directory is empty**.
- If the pod restarts or the deployment scales, the data directory is reused and the scripts are not re-run.

### Caveat
The `seed.sql` in this repo uses plain `INSERT` statements. If the PVC is deleted and the database is recreated, the same products will be inserted again, causing duplicates. To avoid this, make the seed idempotent:

```sql
INSERT INTO products (name, description, price, stock, category, image_url)
VALUES (...)
ON CONFLICT (name) DO NOTHING;
```

Or add a `UNIQUE` constraint on `name`.

### Why
`ON CONFLICT DO NOTHING` is a PostgreSQL idiom that makes the seed insert a no-op if the product already exists, so re-running it never duplicates data.

---

## 3. Avoid Init Containers for Schema/Seed

### Why not an init container in `product-service`
- An init container runs **every time the `product-service` pod starts**, not just once.
- If you scale `product-service` to multiple replicas, each pod's init container will try to initialize the database at the same time, causing race conditions.
- It couples application startup to database schema lifecycle, making rollouts slower and less reliable.
- Seed data is not naturally idempotent, so repeated runs will create duplicate products.

### When init containers are okay
- For waiting until a dependency is ready, e.g., `pg_isready` or `wait-for-it`.
- For one-time, non-data side effects that are safe to repeat.

---

## 4. Industry Best Practice: Use a Migration Tool

### Why
Migrations are version-controlled, idempotent, and auditable. They track which scripts have already been applied in a `schema_migrations` or `flyway_schema_history` table, so re-running the job only applies new changes.

### Tools
- **Flyway** - very common with Spring Boot (Java) apps
- **Liquibase** - XML/YAML/JSON change-log based
- **golang-migrate** - CLI and Go library for Go services
- **Atlas** - newer, schema-as-code approach

### How it works
1. Store migrations as ordered files, e.g.:
   ```
   migrations/
   ├── V1__create_products_table.sql
   ├── V2__create_products_indexes.sql
   └── V3__seed_products.sql
   ```
2. Package a small Docker image with the migration tool and the SQL files.
3. Run the migrations as a Kubernetes `Job` with a Helm `pre-install,pre-upgrade` hook.

### Example Kubernetes Job

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: product-service-migrations
  namespace: demo
  annotations:
    "helm.sh/hook": pre-install,pre-upgrade
    "helm.sh/hook-delete-policy": before-hook-creation
spec:
  template:
    spec:
      restartPolicy: OnFailure
      containers:
        - name: migrate
          image: migrate/migrate:latest
          command:
            - "./migrate"
            - "-path=/migrations"
            - "-database=$(DATABASE_URL)"
            - "up"
          env:
            - name: DATABASE_URL
              value: "postgres://productuser:productpass@postgres-postgresql.demo.svc.cluster.local:5432/products?sslmode=disable"
          volumeMounts:
            - name: migrations
              mountPath: /migrations
      volumes:
        - name: migrations
          configMap:
            name: product-service-migrations
```

### Why this is the best pattern
- **Idempotent**: migration tools track applied scripts and only apply new ones.
- **Safe on restart**: hook job runs once per Helm install/upgrade, not on every pod start.
- **Version-controlled**: each schema change is a small, reviewable file in Git.
- **Rollback support**: `down` migrations can undo changes if a deployment fails.
- **No duplication**: seed inserts are guarded by the migration tool's history table.

---

## 5. Update `product-service` K8s Deployment

### What to do
Set the `DATABASE_URL` environment variable to point to the Helm-generated PostgreSQL service:

```yaml
- name: DATABASE_URL
  value: "postgres://productuser:productpass@postgres-postgresql.demo.svc.cluster.local:5432/products?sslmode=disable"
```

### Service DNS pattern
- Release name `postgres` + chart name `postgresql`:
  `postgres-postgresql.<namespace>.svc.cluster.local`
- In the `demo` namespace:
  `postgres-postgresql.demo.svc.cluster.local:5432`

### Why
`product-service/database/client.go` uses `pgxpool.ParseConfig(cfg.DatabaseURL)` and expects a full connection string. The service name must match the Helm release name and namespace.

### Optional but recommended
Create a Kubernetes Secret for the password and reference it in the deployment instead of hardcoding `productpass`:

```yaml
env:
  - name: DATABASE_URL
    valueFrom:
      secretKeyRef:
        name: postgres-postgresql
        key: password
```

Then construct the full connection string in the app or a startup script. For a Kind demo, a hardcoded secret is acceptable; for anything else, use a real secret manager.

---

## 6. Test the Deployment

### Commands

```bash
# Verify the Postgres pod is running
kubectl get pods -n demo -l app.kubernetes.io/name=postgresql

# Check logs
kubectl logs -n demo -l app.kubernetes.io/name=postgresql

# Port-forward to test locally
kubectl port-forward -n demo svc/postgres-postgresql 5432:5432

# Connect with psql
psql postgres://productuser:productpass@localhost:5432/products

# Verify products table and seed data
\dt
SELECT COUNT(*) FROM products;
```

---

## Summary

| Approach | Restart Safe | Duplicate Safe | Recommended |
| --- | --- | --- | --- |
| Init container in app pod | No | No | Not for schema/seed |
| Helm `initdb.scripts` | Yes (if PVC survives) | No, unless seed is idempotent | Okay for local demo |
| Migration tool (Flyway/golang-migrate) | Yes | Yes | Industry standard |

### Suggested next step
For this demo, use **Helm `primary.initdb.scripts`** with idempotent `seed.sql`. When the project grows, migrate to **golang-migrate or Flyway** as a pre-install Kubernetes Job.
