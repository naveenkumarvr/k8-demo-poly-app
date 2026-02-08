# Product-Service

A production-grade Go microservice for Poly-Shop's inventory management, featuring OpenTelemetry instrumentation, comprehensive testing, and Kubernetes-ready health probes.

## Overview

The Product-Service is a core component of the Poly-Shop microservices suite. It provides a RESTful API for product inventory management and includes CPU-intensive endpoints for Horizontal Pod Autoscaler (HPA) testing.

**Key Features:**
- 🚀 **High Performance**: Built with Gin framework
- 📊 **Full Observability**: OpenTelemetry instrumentation with W3C Trace Context propagation
- 🧪 **Well Tested**: 85%+ code coverage with unit and integration tests
- 🔒 **Secure**: Runs as non-root user in Google Distroless container
- ☸️ **Kubernetes Ready**: Health probes and HPA-compatible stress endpoint

## Architecture

```
product-service/
├── main.go                 # Application entry point
├── handlers/               # HTTP request handlers
│   ├── products.go         # Product inventory endpoint
│   ├── stress.go           # CPU stress testing endpoint
│   └── health.go           # Health check endpoints
├── telemetry/              # OpenTelemetry configuration
│   └── tracer.go           # OTLP/gRPC exporter setup
├── middleware/             # Gin middleware
│   └── tracing.go          # Trace context propagation
└── scripts/                # Testing and utilities
    └── k6-test.js          # Load testing script
```

## API Contract

### Products Endpoint

**GET /products**

Returns a list of available products in the inventory.

**Response:** `200 OK`
```json
[
  {
    "id": 1,
    "name": "Wireless Headphones",
    "description": "Premium noise-cancelling wireless headphones with 30-hour battery life",
    "price": 199.99,
    "image_url": "https://images.example.com/headphones.jpg"
  },
  ...
]
```

**Custom Span:** Creates a `fetch_products_from_database` span with 50-100ms simulated database latency.

---

### Stress Testing Endpoint

**GET /stress?n={number}**

Performs CPU-intensive recursive Fibonacci calculation for HPA testing.

**Query Parameters:**
- `n` (optional): Fibonacci number to calculate (default: 42, max: 50)

**Response:** `200 OK`
```json
{
  "input": 42,
  "result": 267914296,
  "computation_time": "2.543s",
  "message": "CPU stress test completed successfully"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid parameter or n > 50

**Performance Guide:**
- `n=35`: ~0.5 seconds
- `n=40`: ~2-3 seconds
- `n=42`: ~4-5 seconds (default for HPA testing)
- `n=45`: ~15-30 seconds
- `n=50`: ~2-5 minutes

---

### Health Check Endpoints

**GET /healthz**

General health check endpoint.

**Response:** `200 OK`
```json
{
  "status": "ok",
  "service": "product-service"
}
```

---

**GET /ready**

Kubernetes readiness probe. Indicates service is ready to accept traffic.

**Response:** `200 OK`
```json
{
  "status": "ready",
  "service": "product-service"
}
```

---

**GET /live**

Kubernetes liveness probe. Indicates service is alive and doesn't need restart.

**Response:** `200 OK`
```json
{
  "status": "alive",
  "service": "product-service"
}
```

## OpenTelemetry Instrumentation

### Trace Context Propagation

The service implements **W3C Trace Context** standard for distributed tracing:

1. **Incoming Requests**: Extracts `traceparent` header from HTTP requests
2. **Span Creation**: Automatically creates spans for all HTTP requests
3. **Outgoing Requests**: Injects trace context into downstream service calls
4. **Custom Spans**: Manual instrumentation for database operations

### Span Attributes

**HTTP Request Spans:**
- `http.method`: HTTP method (GET, POST, etc.)
- `http.route`: Route pattern
- `http.status_code`: Response status code
- `http.client_ip`: Client IP address

**Product Fetch Spans:**
- `product.count`: Number of products returned
- `database.operation`: SQL operation (SELECT)
- `database.table`: Table name
- `fetch.duration_ms`: Fetch duration in milliseconds

**Stress Test Spans:**
- `fibonacci.input`: Input value for calculation
- `computation.duration_ms`: Computation time
- `fibonacci.result`: Calculation result

### OTLP Exporter Configuration

The service exports traces to an OpenTelemetry Collector via OTLP/gRPC:

```
Service → OTLP/gRPC (port 4317) → OTel Collector → Backend (Jaeger, Tempo, etc.)
```

**Configuration:** Set `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable.

## Local Development

### Prerequisites

- Go 1.21 or higher
- Docker (for containerization)
- k6 (optional, for load testing)

### Build and Run Locally

**Method 1: Using Go**

```bash
# Navigate to the service directory
cd d:\code\poly-app\product-service

# Download dependencies
go mod download

# Run the service
go run main.go

# The service will start on port 8080
# Visit http://localhost:8080/products
```

**Method 2: Using Docker**

```bash
# Build the Docker image
cd d:\code\poly-app\product-service
docker build -t product-service:latest .

# Run the container
docker run -d \
  -p 8080:8080 \
  --name product-service \
  -e SERVICE_NAME=product-service \
  -e OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317 \
  -e ENVIRONMENT=development \
  product-service:latest

# View logs
docker logs -f product-service

# Stop and remove
docker stop product-service
docker rm product-service
```

**Windows PowerShell:**

```powershell
# Build
docker build -t product-service:latest .

# Run
docker run -d `
  -p 8080:8080 `
  --name product-service `
  -e SERVICE_NAME=product-service `
  -e OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317 `
  -e ENVIRONMENT=development `
  product-service:latest
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVICE_NAME` | Service identifier for traces | `product-service` |
| `SERVICE_VERSION` | Service version | `1.0.0` |
| `ENVIRONMENT` | Deployment environment | `development` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTel Collector endpoint | `localhost:4317` |
| `PORT` | HTTP server port | `8080` |

Copy `.env.example` to `.env` and update as needed.

## Testing

### Run All Tests

```bash
# Run all tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage summary
go tool cover -func=coverage.out | grep total

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
```

### Run Specific Tests

```bash
# Test only product handlers
go test -v ./handlers -run TestGetProducts

# Test stress endpoint
go test -v ./handlers -run TestStressTest

# Test health endpoints
go test -v ./handlers -run TestHealthz
```

### Benchmark Tests

```bash
# Run all benchmarks
go test -bench=. -benchmem ./...

# Benchmark specific functions
go test -bench=BenchmarkFibonacci -benchmem ./handlers
```

## HPA Testing

The `/stress` endpoint is designed for Horizontal Pod Autoscaler testing.

### Triggering CPU Load

**Single Request:**
```bash
# Light load (n=35)
curl "http://localhost:8080/stress?n=35"

# Medium load (n=40)
curl "http://localhost:8080/stress?n=40"

# Heavy load (n=42, recommended for HPA)
curl "http://localhost:8080/stress?n=42"
```

**Sustained Load with Apache Bench:**
```bash
# 100 requests, 10 concurrent
ab -n 100 -c 10 "http://localhost:8080/stress?n=40"
```

**Sustained Load with k6:**
```bash
# Run the provided k6 script
k6 run scripts/k6-test.js
```

### Monitoring Container Metrics

While running stress tests, monitor CPU usage:

```bash
# Docker stats
docker stats product-service

# Expected output during stress:
# CPU should spike to 80-100% during Fibonacci calculations
```

**In Kubernetes:**
```bash
# Watch pod metrics
kubectl top pod -l app=product-service

# Watch HPA status
kubectl get hpa product-service -w
```

### Expected HPA Behavior

1. Deploy with HPA configured for 50% CPU target
2. Hit `/stress?n=42` endpoint repeatedly
3. CPU usage spikes to 90-100%
4. HPA detects high CPU usage
5. New pods are scheduled
6. Load distributes across pods
7. CPU per pod drops below threshold

## Load Testing with k6

### Install k6

**Windows:**
```powershell
winget install k6
```

**MacOS:**
```bash
brew install k6
```

**Linux:**
```bash
curl -L https://github.com/grafana/k6/releases/download/v0.47.0/k6-v0.47.0-linux-amd64.tar.gz | tar xvz
sudo mv k6-v0.47.0-linux-amd64/k6 /usr/local/bin/
```

### Run Load Tests

```bash
# Default test (targets localhost:8080)
k6 run scripts/k6-test.js

# Custom base URL
k6 run -e BASE_URL=http://product-service:8080 scripts/k6-test.js

# With custom thresholds
k6 run --no-thresholds scripts/k6-test.js
```

### Load Test Stages

1. **Smoke Test** (30s, 1 VU): Verify basic functionality
2. **Ramp-up** (1m, 10 VUs): Gradual load increase
3. **Load Test** (2m, 50 VUs): Sustained load
4. **Stress Test** (1m, 100 VUs): High load
5. **Spike Test** (30s, 150 VUs): Sudden traffic spike
6. **Ramp-down** (1m, 0 VUs): Graceful decrease

### Expected Metrics

- **p95 latency** < 500ms for `/products`
- **Error rate** < 1%
- **Success rate** > 95%

## Docker Image Details

### Multi-Stage Build

The Dockerfile uses a **multi-stage build** for optimal security and size:

1. **Builder Stage** (`golang:1.21-alpine`):
   - Compiles the Go application
   - Downloads dependencies
   - Creates static binary

2. **Runtime Stage** (`gcr.io/distroless/static-debian12`):
   - Contains only the compiled binary
   - No shell, package managers, or unnecessary tools
   - Runs as non-root user (UID 65532)

### Image Size

- **Final image**: ~15-20 MB (distroless base + binary)
- **Comparison**: Standard `golang:1.21` image is ~800 MB

### Security Features

✅ **Non-root user**: Runs as UID 65532  
✅ **Minimal attack surface**: No shell or package manager  
✅ **Static binary**: No runtime dependencies  
✅ **Distroless base**: Google-maintained minimal image

### Verify Security

```bash
# Verify non-root user
docker run --rm product-service:latest whoami
# Expected: nonroot (or error, since distroless has no shell)

# Inspect image layers
docker history product-service:latest

# Check image size
docker images product-service:latest
```

## Kubernetes Deployment

### Sample Deployment Manifest

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: product-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: product-service
  template:
    metadata:
      labels:
        app: product-service
    spec:
      containers:
      - name: product-service
        image: product-service:latest
        ports:
        - containerPort: 8080
        env:
        - name: SERVICE_NAME
          value: "product-service"
        - name: OTEL_EXPORTER_OTLP_ENDPOINT
          value: "otel-collector.monitoring.svc.cluster.local:4317"
        - name: ENVIRONMENT
          value: "production"
        livenessProbe:
          httpGet:
            path: /live
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            cpu: 100m
            memory: 64Mi
          limits:
            cpu: 500m
            memory: 128Mi
```

### HPA Configuration

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: product-service-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: product-service
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 50
```

## Troubleshooting

### Service Won't Start

**Issue:** Service fails to start or exits immediately.

**Solution:**
1. Check logs: `docker logs product-service`
2. Verify environment variables are set correctly
3. Ensure port 8080 is not already in use
4. Check OTel Collector is accessible

### High Memory Usage

**Issue:** Container uses excessive memory.

**Solution:**
1. The service is designed to be lightweight (~20-30 MB)
2. Check for memory leaks with `go test -memprofile`
3. Monitor with `docker stats`

### Traces Not Appearing

**Issue:** Traces not visible in observability backend.

**Solution:**
1. Verify `OTEL_EXPORTER_OTLP_ENDPOINT` is correct
2. Check OTel Collector is running and accessible
3. Verify network connectivity: `telnet otel-collector 4317`
4. Check collector logs for errors

### Tests Failing

**Issue:** Tests fail or coverage is low.

**Solution:**
1. Ensure all dependencies are downloaded: `go mod download`
2. Run tests in verbose mode: `go test -v ./...`
3. Check for race conditions: `go test -race ./...`

## Contributing

When contributing to this service, please:

1. Maintain 85%+ code coverage
2. Add tests for all new features
3. Update API documentation
4. Follow Go best practices and conventions
5. Ensure Docker image builds successfully

## License

Part of the Poly-Shop microservices suite. For educational and practice purposes.

---

**Service Version:** 1.0.0  
**Go Version:** 1.21+  
**Framework:** Gin  
**Observability:** OpenTelemetry
