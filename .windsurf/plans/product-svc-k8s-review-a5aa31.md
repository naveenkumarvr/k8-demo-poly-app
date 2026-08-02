# Review product-service Kubernetes manifests and Dockerfile

Compare `k8s/product-svc/deploy.yaml`, `configmap.yaml`, and `service.yaml` with `product-service/Dockerfile` and `product-service/main.go` to identify correctness, security, and operational issues before the user proceeds.

## Scope
- Validate image, ports, environment variables, probes, namespaces, and ConfigMap values
- Check Dockerfile for multi-arch build correctness and probe compatibility
- List issues found without editing files
