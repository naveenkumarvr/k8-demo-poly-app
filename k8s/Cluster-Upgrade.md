# Cluster Upgrade: Kubernetes 1.31 → 1.36

Step-by-step guide to move the local kind cluster from Kubernetes **1.31** to **1.36** so that Envoy Gateway v1.9 can be installed. Run all commands in **WSL** from the repo root (`/mnt/d/code/k8-demo-poly-app`).

## Why

Envoy Gateway v1.9.1 ships Gateway API CRDs whose validation rules use the CEL `format` library (`format.dns1123Label()`). The Kubernetes 1.31 API server does not have that library, so the install fails with:

```
undeclared reference to 'format'
```

| Envoy Gateway | Supported Kubernetes |
|---------------|---------------------|
| v1.9 | 1.33 – 1.36 |
| v1.8 | 1.32 – 1.35 |

Source: [Envoy Gateway compatibility matrix](https://gateway.envoyproxy.io/news/releases/matrix/)

## Target versions

| Tool | From | To |
|------|------|----|
| kind CLI | current | **v0.33.0** |
| Kubernetes (node image) | v1.31 | **v1.36.4** (pinned in [kind_config.yaml](kind_config.yaml)) |
| kubectl | current | **v1.36.x** (must be within ±1 minor of the cluster) |
| Envoy Gateway | – | **v1.9.1** |

v1.36 is used instead of kind's default v1.37 because Envoy Gateway v1.9 is tested up to 1.36.

> **kind clusters cannot be upgraded in place.** The upgrade deletes the cluster and creates a new one. Everything in the cluster is lost, including Postgres and Redis data, and must be redeployed. Complete Step 1 first if you need any data.

---

## Step 0: Check the current state

```bash
kind version
kubectl version
kind get clusters
kubectl get pods -A
```

Also confirm the host uses cgroup v2, because Kubernetes 1.35+ dropped cgroup v1. Expect `cgroup2fs`:

```bash
stat -fc %T /sys/fs/cgroup
```

If it prints `tmpfs` (cgroup v1), update WSL with `wsl --update` from PowerShell, then run `wsl --shutdown` and reopen WSL.

## Step 1 (optional): Back up data

Only needed if you want to keep checkout transactions or current cart contents. Products are re-seeded automatically by the Postgres init scripts.

```bash
mkdir -p ~/k8s-backup

# Checkout transactions
kubectl exec -n postgres postgres-postgresql-0 -- env PGPASSWORD=checkoutpass \
  pg_dump -U checkoutuser -d checkout --data-only -t transactions > ~/k8s-backup/checkout.sql

# Redis snapshot (carts)
kubectl exec -n demo deploy/polyapp-redis -- redis-cli SAVE
kubectl cp demo/$(kubectl get pod -n demo -l app=polyapp-redis -o jsonpath='{.items[0].metadata.name}'):/data/dump.rdb ~/k8s-backup/dump.rdb

ls -lh ~/k8s-backup
```

## Step 2: Upgrade the kind CLI

```bash
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.33.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind

kind version    # expect: kind v0.33.0
```

## Step 3: Upgrade kubectl

```bash
curl -LO "https://dl.k8s.io/release/v1.36.4/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/kubectl

kubectl version --client    # expect: Client Version: v1.36.4
```

If `which kubectl` points somewhere other than `/usr/local/bin/kubectl`, for example Docker Desktop's bundled copy, update or remove that one so the new binary is used.

## Step 4: Delete the old cluster

```bash
kind delete cluster --name k8s
docker ps --filter "name=k8s-"    # expect: no containers
```

## Step 5: Create the new cluster

[kind_config.yaml](kind_config.yaml) pins every node to `kindest/node:v1.36.4` by digest, so no `--image` flag is needed:

```bash
kind create cluster --name k8s --config k8s/kind_config.yaml
```

Verify that all 3 nodes are `Ready` on v1.36.4:

```bash
kubectl version            # Server Version: v1.36.4
kubectl get nodes -o wide  # 1 control-plane + 2 workers, all Ready, VERSION v1.36.4
kubectl get pods -n kube-system
```

## Step 6: Install Envoy Gateway

```bash
helm install eg oci://docker.io/envoyproxy/gateway-helm \
  --version v1.9.1 -n envoy-gateway-system --create-namespace

kubectl wait --timeout=5m -n envoy-gateway-system deployment/envoy-gateway --for=condition=Available
kubectl get crd | grep gateway.networking    # Gateway API CRDs installed
```

## Step 7: Redeploy the application

Follow [Readme.md](Readme.md) from **step 2** onward. Quick reference, in dependency order:

```bash
# Namespaces
kubectl create namespace postgres
kubectl create namespace demo
kubectl create namespace observability

# Postgres (products + checkout databases are created by init scripts)
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
helm upgrade --install postgres bitnami/postgresql -n postgres -f k8s/postgres/values.yaml
kubectl wait -n postgres --for=condition=Ready pod/postgres-postgresql-0 --timeout=180s

# Redis
kubectl apply -n demo -f k8s/redis/

# Jaeger
kubectl apply -f k8s/jaeger_tracing/

# Backend services
kubectl apply -n demo -f k8s/product-svc/
kubectl apply -n demo -f k8s/cart-service/
kubectl apply -f k8s/checkout-svc/
kubectl apply -f k8s/notification-worker/

# UI last, because nginx resolves backend Service names at startup
kubectl apply -f k8s/shop-ui/

kubectl get pods -A
```

If any image was loaded with `kind load` instead of pushed to Docker Hub, load it into the new cluster again:

```bash
kind load docker-image naveenvr0792/poly_app-checkout_svc:v1 --name k8s
```

## Step 8 (optional): Restore data

Run after checkout-service has started at least once, so that Hibernate has created the `transactions` table:

```bash
# Checkout transactions
kubectl exec -i -n postgres postgres-postgresql-0 -- env PGPASSWORD=checkoutpass \
  psql -U checkoutuser -d checkout < ~/k8s-backup/checkout.sql

# Redis: copy the snapshot, then restart so Redis loads it
REDIS_POD=$(kubectl get pod -n demo -l app=polyapp-redis -o jsonpath='{.items[0].metadata.name}')
kubectl cp ~/k8s-backup/dump.rdb demo/$REDIS_POD:/data/dump.rdb
kubectl rollout restart -n demo deploy/polyapp-redis
```

Redis runs with AOF enabled, and on startup it loads the AOF file rather than `dump.rdb` when both exist. If the carts do not come back, delete `/data/appendonlydir` before restarting.

## Step 9: Verify

```bash
kubectl get nodes                                   # v1.36.4, all Ready
kubectl get pods -A | grep -v Running               # only the header line (and Completed jobs)
kubectl get pods -n envoy-gateway-system            # envoy-gateway Running

kubectl exec -n postgres postgres-postgresql-0 -- env PGPASSWORD=productpass \
  psql -U productuser -d products -c "select count(*) from products;"   # 16

kubectl port-forward -n demo svc/shop-ui-ms-svc 8070:8080
# open http://localhost:8070 — products list, add to cart, checkout
```

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| `kind create cluster` fails pulling the node image | Check Docker Desktop is running and WSL integration is enabled, then retry. |
| Nodes stay `NotReady` | `kubectl describe node <name>` and `docker logs k8s-control-plane`. On cgroup v1, see Step 0. |
| kubectl warns about version skew | Client is more than one minor version away from 1.36. Redo Step 3. |
| Envoy Gateway install still fails with the CEL error | `kubectl version` still shows < 1.33, so the old cluster or kubeconfig context is in use. Run `kubectl config current-context` (expect `kind-k8s`). |
| Helm says `cannot re-use a name that is still in use` | Remove the stale release: `helm uninstall eg -n envoy-gateway-system`. |
| Postgres pod not ready | `kubectl logs -n postgres postgres-postgresql-0` shows init script errors. |
| shop-ui pod crash-loops | A backend Service does not exist yet. Deploy the backends, then run `kubectl rollout restart -n demo deploy/shop-ui-ms`. |

## Rollback

To go back to Kubernetes 1.31, remove the `image:` lines from [kind_config.yaml](kind_config.yaml), recreate the cluster with `--image kindest/node:v1.31.14@sha256:6f86cf509dbb42767b6e79debc3f2c32e4ee01386f0489b3b2be24b0a55aac2b` (from kind v0.31.0), and use Envoy Gateway v1.6.x. v1.6 is end of life.
