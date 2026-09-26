# Kind Deployment

Step-by-step commands to deploy PolyShop on a local kind cluster. Run all commands from the repo root.

| Component | Namespace |
|-----------|-----------|
| PostgreSQL (Helm) | `postgres` |
| Redis | `demo` |
| Jaeger (all-in-one) | `observability` |
| product-service | `demo` |
| checkout-service | `demo` |
| notification-worker | `demo` |

## 1. Create the cluster

```bash
kind create cluster --name k8s --config k8s/kind_config.yaml
kubectl get nodes
```

## 2. Create namespaces

```bash
kubectl create namespace postgres
kubectl create namespace demo
kubectl create namespace observability
```

## 3. Install PostgreSQL

[postgres/values.yaml](postgres/values.yaml) sets up both databases through `primary.initdb.scripts`. The scripts run as the `postgres` superuser, and only once, when the data volume is empty:

| Script | Creates |
|--------|---------|
| `00-checkout-db.sql` | `checkoutuser` role and `checkout` database (checkout-service creates its `transactions` table on startup) |
| `01-schema.sql` | `products` table, owned by `productuser` |
| `02-seed.sql` | 16 sample products |

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

# Optional: preview the rendered init-scripts ConfigMap
helm template postgres bitnami/postgresql -n postgres -f k8s/postgres/values.yaml | grep -n -A3 "docker-entrypoint-initdb.d"

helm upgrade --install postgres bitnami/postgresql -n postgres -f k8s/postgres/values.yaml
kubectl wait -n postgres --for=condition=Ready pod/postgres-postgresql-0 --timeout=180s
```

Verify the products schema and seed data (expect owner `productuser` and 16 rows):

```bash
kubectl exec -n postgres postgres-postgresql-0 -- env PGPASSWORD=productpass \
  psql -U productuser -d products -c "\dt" -c "select count(*) from products;"
```

Verify the checkout database exists and `checkoutuser` can log in:

```bash
kubectl exec -n postgres postgres-postgresql-0 -- env PGPASSWORD=checkoutpass \
  psql -U checkoutuser -d checkout -c "select current_user, current_database();"
```

## 4. Deploy Redis

Used by cart-service. Runs with AOF persistence on a 2Gi PVC (`standard` storage class) and a 256MB LRU memory cap.

```bash
kubectl apply -n demo -f k8s/redis/redis_pvc.yaml
kubectl apply -n demo -f k8s/redis/redis_deploy.yaml
kubectl apply -n demo -f k8s/redis/redis_svc.yaml

kubectl rollout status -n demo deploy/polyapp-redis
kubectl get pvc,pods,svc -n demo -l app=polyapp-redis
```

Verify Redis responds (expect `PONG`) and persistence is on (expect `appendonly yes`):

```bash
kubectl exec -n demo deploy/polyapp-redis -- redis-cli ping
kubectl exec -n demo deploy/polyapp-redis -- redis-cli config get appendonly
```

In-cluster address for cart-service (`REDIS_ADDR`): `polyapp-redis.demo.svc.cluster.local:6379`

## 5. Deploy Jaeger (tracing)

Jaeger all-in-one with in-memory storage. Traces are lost when the pod restarts.

```bash
kubectl apply -f k8s/jaeger_tracing/configmap.yaml
kubectl apply -f k8s/jaeger_tracing/deploy.yaml
kubectl apply -f k8s/jaeger_tracing/service.yaml

kubectl rollout status -n observability deploy/jaeger-tracing
kubectl get pods,svc -n observability
```

Open the UI at http://localhost:16686 (until ingress is added):

```bash
kubectl port-forward -n observability svc/jaeger-tracing-svc 16686:16686
```

| Port | Purpose |
|------|---------|
| 16686 | Web UI |
| 4317 | OTLP gRPC receiver (used by the Go services) |
| 4318 | OTLP HTTP receiver |
| 14268 | Jaeger Thrift HTTP collector |
| 9411 | Zipkin-compatible endpoint |

Set `OTEL_EXPORTER_OTLP_ENDPOINT` in each service ConfigMap to `jaeger-tracing-svc.observability.svc.cluster.local:4317`. Go services use host:port with no scheme. The Java agent in checkout-service needs the `http://` prefix and `OTEL_EXPORTER_OTLP_PROTOCOL=grpc`.

## 6. Deploy product-service

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

## 7. Test product-service

```bash
kubectl port-forward -n demo svc/poly-shop-product-service 8090:8090
curl http://localhost:8090/ready
curl http://localhost:8090/products
```

Then open the Jaeger UI, select `product-service` in the **Service** dropdown, and click **Find Traces**.

## 8. Deploy checkout-service

Stores transactions in the `checkout` Postgres database and calls cart-service at `CART_SERVICE_URL`. Requires the `checkout` database created by the Postgres init scripts in step 3.

Build and push the image after any code or `pom.xml` change:

```bash
docker build -t naveenvr0792/poly_app-checkout_svc:v1 checkout-service
docker push naveenvr0792/poly_app-checkout_svc:v1
# or, without a registry: kind load docker-image naveenvr0792/poly_app-checkout_svc:v1 --name k8s
```

Deploy:

```bash
kubectl apply -f k8s/checkout-svc/secrets.yaml
kubectl apply -f k8s/checkout-svc/configmap.yaml
kubectl apply -f k8s/checkout-svc/deploy.yaml
kubectl apply -f k8s/checkout-svc/service.yaml

kubectl rollout status -n demo deploy/poly-shop-checkout-ms --timeout=180s
```

The startup probe allows up to 150s for the 15s `STARTUP_DELAY_SECONDS` plus JVM boot.

## 9. Test checkout-service

```bash
kubectl port-forward -n demo svc/poly-shop-checkout-ms-svc 8085:8085
curl http://localhost:8085/actuator/health
curl -X POST http://localhost:8085/checkout -H "Content-Type: application/json" -d '{"userId":"user123"}'
```

Verify transactions are persisted:

```bash
kubectl exec -n postgres postgres-postgresql-0 -- env PGPASSWORD=checkoutpass \
  psql -U checkoutuser -d checkout -c "select transaction_id, user_id, status, created_at from transactions order by created_at desc limit 5;"
```

The rows survive `kubectl rollout restart -n demo deploy/poly-shop-checkout-ms`.

## 10. Deploy notification-worker

Background worker that subscribes to the Redis Pub/Sub channel `cart-events` and logs each `cart.updated` / `cart.cleared` event published by cart-service. It has no Service or probes, because it exposes no HTTP endpoint and runs on a distroless image. Keep it at 1 replica, since each Pub/Sub subscriber receives every message.

```bash
docker build -t naveenvr0792/poly_app-notification_worker:v1 notification-worker
docker push naveenvr0792/poly_app-notification_worker:v1

kubectl apply -f k8s/notification-worker/configmap.yaml
kubectl apply -f k8s/notification-worker/deploy.yaml
kubectl rollout status -n demo deploy/notification-worker
```

Verify: add an item to the cart in the UI, then check that the worker logged the event:

```bash
kubectl logs -n demo deploy/notification-worker -f
# {"msg":"Processed cart event asynchronously","event":"cart.updated","user_id":"user-1","total_items":1,...}
```

Or publish a test event directly:

```bash
kubectl exec -n demo deploy/polyapp-redis -- redis-cli PUBLISH cart-events '{"event":"cart.updated","user_id":"test","total_items":3}'
```

`PUBLISH` returns the number of subscribers that received the event. Expect `1`. `0` means the worker is not subscribed.

## Troubleshooting

```bash
# App logs (use --previous for a crashed container)
kubectl logs -n demo deploy/poly-shop-product-service --previous
kubectl describe pod -n demo -l app=poly-shop-product-service

# Postgres logs, including init script output
kubectl logs -n postgres postgres-postgresql-0 | grep -i init

# Redis logs and an interactive CLI
kubectl logs -n demo deploy/polyapp-redis
kubectl exec -it -n demo deploy/polyapp-redis -- redis-cli

# Jaeger logs (look for OTLP receiver start and errors)
kubectl logs -n observability deploy/jaeger-tracing

# checkout-service logs and probe events
kubectl logs -n demo deploy/poly-shop-checkout-ms
kubectl describe pod -n demo -l app=poly-shop-checkout-ms

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