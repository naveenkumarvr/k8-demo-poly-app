# Kind Deployment

Step-by-step commands to deploy PolyShop on a local kind cluster. Run all commands from the repo root.

| Component | Namespace |
|-----------|-----------|
| PostgreSQL (Helm) | `postgres` |
| product-service | `demo` |

## 1. Create the cluster

```bash
kind create cluster --name k8s --config k8s/kind_config.yaml
kubectl get nodes
```

## 2. Create namespaces

```bash
kubectl create namespace postgres
kubectl create namespace demo
```

## 3. Install PostgreSQL

Schema and seed data are defined in `primary.initdb.scripts` in [postgres/values.yaml](postgres/values.yaml). They run only once, when the data volume is empty.

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

# Optional: preview the rendered init-scripts ConfigMap
helm template postgres bitnami/postgresql -n postgres -f k8s/postgres/values.yaml | grep -n -A3 "docker-entrypoint-initdb.d"

helm upgrade --install postgres bitnami/postgresql -n postgres -f k8s/postgres/values.yaml
kubectl wait -n postgres --for=condition=Ready pod/postgres-postgresql-0 --timeout=180s
```

Verify the schema and seed data (expect owner `productuser` and 16 rows):

```bash
kubectl exec -n postgres postgres-postgresql-0 -- env PGPASSWORD=productpass \
  psql -U productuser -d products -c "\dt" -c "select count(*) from products;"
```

## 4. Deploy product-service

`DB_HOST` in [product-svc/configmap.yaml](product-svc/configmap.yaml) must be `postgres-postgresql.postgres.svc.cluster.local`.

```bash
kubectl apply -n demo -f k8s/product-svc/secrets.yaml
kubectl apply -n demo -f k8s/product-svc/configmap.yaml
kubectl apply -n demo -f k8s/product-svc/deploy.yaml
kubectl apply -n demo -f k8s/product-svc/service.yaml

kubectl rollout status -n demo deploy/poly-shop-product-service
kubectl get all -n demo
```

After changing the ConfigMap or Secret, restart the pods to pick up the new values:

```bash
kubectl rollout restart -n demo deploy/poly-shop-product-service
```

## 5. Test product-service

```bash
kubectl port-forward -n demo svc/poly-shop-product-service 8090:8090
curl http://localhost:8090/ready
curl http://localhost:8090/products
```

## Troubleshooting

```bash
# App logs (use --previous for a crashed container)
kubectl logs -n demo deploy/poly-shop-product-service --previous
kubectl describe pod -n demo -l app=poly-shop-product-service

# Postgres logs, including init script output
kubectl logs -n postgres postgres-postgresql-0 | grep -i init

# Test DB connectivity from inside the demo namespace
kubectl run pgtest -n demo --rm -it --image=postgres:16 -- \
  psql "postgres://productuser:productpass@postgres-postgresql.postgres.svc.cluster.local:5432/products" -c "select 1"

# Inspect applied Helm values
helm get values postgres -n postgres
```

### Re-run init scripts (deletes all Postgres data)

Init scripts do not run again on an existing volume. To re-seed after changing `values.yaml`:

```bash
helm uninstall postgres -n postgres
kubectl delete pvc -n postgres data-postgres-postgresql-0
helm upgrade --install postgres bitnami/postgresql -n postgres -f k8s/postgres/values.yaml
```

## Cleanup

```bash
kind delete cluster --name k8s
```