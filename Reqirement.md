Role & Context:
You are a Principal Cloud-Native & Kubernetes Architect conducting a strict, production-readiness codebase audit on this repository (`k8-demo-poly-app`).
We are preparing this multi-tier application for local orchestration on a multi-node Kind (Kubernetes in Docker) cluster, followed by deployment to an AWS EKS production cluster with ArgoCD GitOps, Istio Service Mesh, Karpenter autoscaling, Kyverno security policies, Prometheus observability, and k6/Locust load testing.

Objective:
Deeply analyze all directories, source code, Dockerfiles, configurations, and docker-compose files in this workspace. Assess whether this repository satisfies all necessary requirements for an enterprise-grade, multi-tier Kubernetes demo platform.

Evaluate the codebase against these 6 Core Enterprise Dimensions:

1. MULTI-TIER ARCHITECTURE & DEPENDENCIES
   - Is there a clear separation of tiers? (e.g., Frontend UI, API Gateway/Backend Microservices, Relational DB, NoSQL/Cache, Async Worker/Message Queue).
   - Are microservices polyglot or distinct? Identify all languages and frameworks used.
   - Is there at least one asynchronous or event-driven background processing pipeline (e.g., Kafka, RabbitMQ, Redis Pub/Sub, Worker process)?
   - Are database connection pools, read/write routes, and cache hit/miss logic real (functional) rather than stubbed mock endpoints?

2. CONTAINERIZATION & DOCKER HARDENING (CKS / CIS Standards)
   - Are multi-stage Dockerfiles present for all custom services?
   - Do Dockerfiles run as a explicit NON-ROOT user (e.g., `USER 10001` or unprivileged service account)?
   - Are base images minimal (e.g., Alpine, Distroless, or Debian-slim)?
   - Is `.dockerignore` configured properly for each service to exclude build artifacts, `.git`, `node_modules`, and local secrets?
   - Is a `docker-compose.yaml` present that allows testing the entire application locally?

3. KUBERNETES LIFECYCLE & RESILIENCE (Zero-Downtime Ready)
   - Does each HTTP service expose explicit Health Check endpoints:
     a) Liveness Probe (e.g., `/healthz` or `/health/liveness`)
     b) Readiness Probe (e.g., checking active DB/Cache connection at `/ready` or `/health/readiness`)
   - Do application processes explicitly capture `SIGTERM` signals and execute Graceful Shutdown routines (draining HTTP connections, flushing DB connections, finishing active queue jobs)?

4. OBSERVABILITY & TELEMETRY (Prometheus / Grafana / Istio Ready)
   - Does each microservice expose Prometheus metrics (e.g., at `/metrics`) or generate RED metrics (Request rate, Error rate, Duration/Latency)?
   - Are logs formatted as structured JSON sent exclusively to `stdout` / `stderr`?
   - Do logging statements include essential contextual fields (timestamp, log level, service name, request/trace ID)?
   - Do HTTP microservices pass along/propagate HTTP headers (e.g., `x-request-id`, `x-b3-traceid`, W3C tracecontext) for distributed tracing?

5. CONFIGURATION & SECRETS EXTERNALIZATION (12-Factor App)
   - Are ALL configuration variables, service URLs, DB strings, and ports externalized via Environment Variables?
   - Are there ANY hardcoded credentials, API keys, passwords, or hostnames in the source code or Dockerfiles?
   - Can services read configuration mounted from files (to support ConfigMaps, Kubernetes Secrets, and HashiCorp Vault injection)?

6. LOAD & INFRASTRUCTURE STRESS COMPATIBILITY
   - Do the API endpoints execute actual computational or database workload operations (so that under synthetic load via k6/Locust, CPU/Memory usage and DB connections measurably rise to trigger Horizontal Pod Autoscalers)?

---

REQUIRED OUTPUT FORMAT:

Provide a structured, executive-level audit report organized as follows:

1. EXECUTIVE SUMMARY:
   - Brief overview of detected services, language stack, databases, and general architecture.

2. DETAILED AUDIT CHECKLIST:
   For each item below, mark as [PASS], [FAIL], [WARN] (Sub-optimal), or [MISSING]:
   - [ ] Multi-Tier Architecture & Async Processing
   - [ ] Multi-Stage & Non-Root Dockerfiles
   - [ ] Liveness & Readiness Probes
   - [ ] Graceful Shutdown (`SIGTERM` Handling)
   - [ ] Prometheus `/metrics` Endpoints
   - [ ] Structured JSON Logging to `stdout`
   - [ ] Distributed Trace Header Propagation
   - [ ] Externalized Configuration & Zero Hardcoded Secrets
   - [ ] Load Test Workload Realism

3. GAPS & CRITICAL DEFICIENCIES:
   - List every missing feature, security issue, or Kubernetes incompatibility found.

4. REMEDIATION CODE SNIPPETS & ACTION PLAN:
   - Provide exact, production-ready code snippets, updated `Dockerfile`s, health check handlers, or configuration blocks needed to bring the codebase up to 100% enterprise standards.