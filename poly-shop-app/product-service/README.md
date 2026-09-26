# Product-Service

Production-ready microservice for managing product catalog with PostgreSQL backend, OpenTelemetry tracing, and comprehensive observability. Part of the Poly-Shop microservices suite.

## Overview

Product-Service provides a REST API for product catalog management with:
- **PostgreSQL Backend**: Persistent product storage with connection pooling and retry logic
- **Repository Pattern**: Clean abstraction layer for database operations
- **OpenTelemetry**: Full distributed tracing with W3C Trace Context propagation
- **Category Filtering**: Query products by category for efficient browsing
- **Database Health Checks**: Monitor PostgreSQL connectivity via `/healthz` endpoint
- **Sample Data**: Pre-loaded with 16 products across 4 categories
- **Production-Ready**: Distroless container, graceful shutdown, retry logic

## Architecture

```mermaid
graph LR
    Client[HTTP Client] -->|GET /products| ProductService[Product Service :8090]
    ProductService -->|OTLP/gRPC| Collector[OTel Collector]
    ProductService -->|SQL Query| PostgreSQL[(PostgreSQL)]
    Collector -->|Traces| Jaeger[Jaeger UI :16687]
    PostgreSQL -->|Schema + Seed| InitScripts[docker-entrypoint-initdb.d]
```

### Database Schema

```sql
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
    category VARCHAR(100),
    image_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for query performance
CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_name ON products(name);
CREATE INDEX idx_products_price ON products(price);
```

**Sample Categories**:
- Electronics (MacBook, iPhone, Headphones, TV, Monitor)
- Clothing (Jeans, Sneakers, Jacket, Shirt)
- Books (Tech books, Business, Fiction)
- Home & Garden (Chair, Vacuum, Mixer, Grill)

### Project Structure

```
product-service/
├── main.go                 # Application entry point with DB initialization
├── database/               # PostgreSQL integration
│   ├── client.go           # Connection pool with retry logic
│   ├── repository.go       # Product repository (CRUD operations)
│   ├── schema.sql          # Database schema with indexes
│   ├── seed.sql            # Sample product data
│   └── reset_and_seed.sql  # Script to truncate and re-seed database
├── handlers/               # HTTP request handlers
│   ├── products.go         # Product endpoints (uses repository)
│   ├── stress.go           # CPU stress testing endpoint
│   └── health.go           # Health checks with DB ping
├── telemetry/              # OpenTelemetry configuration
│   └── tracer.go           # OTLP/gRPC exporter setup
├── middleware/             # Gin middleware
│   └── tracing.go          # Trace context propagation
├── docker-compose.yml      # Local stack (postgres + service + jaeger)
└── scripts/                # Testing and utilities
    └── k6-test.js          # Load testing script
```

## API Contract

### Products Endpoint

**GET /products**

Returns all products from PostgreSQL database.

**Response:** `200 OK`
```json
[
  {
    "id": 1,
    "name": "MacBook Pro 16\"",
    "description": "Apple M3 Max chip with 16-core CPU and 40-core GPU, 64GB unified memory, 1TB SSD storage",
    "price": 3499.00,
    "stock": 25,
    "category": "Electronics",
    "image_url": "https://picsum.photos/seed/laptop1/400/300",
    "created_at": "2026-02-08T14:13:44.951622Z",
    "updated_at": "2026-02-08T14:13:44.951622Z"
  },
  ...
]
```

**GET /products?category={categoryName}**

Filter products by category.

**Query Parameters:**
- `category` (optional): Category name (e.g., "Electronics", "Clothing", "Books", "Home & Garden")

**Example:**
```bash
curl "http://localhost:8090/products?category=Electronics"
```

**OpenTelemetry Spans:** Creates `repository.GetAllProducts` or `repository.GetProductsByCategory` spans with actual database query timing.

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

Health check with database connectivity monitoring.

**Response:** `200 OK` (when healthy)
```json
{
  "status": "healthy",
  "service": "product-service",
  "pod_name": "docker-compose-product",
  "node_name": "localhost",
  "database": "healthy"
}
```

**Response:** `503 Service Unavailable` (when database is down)
```json
{
  "status": "unhealthy",
  "service": "product-service",
  "pod_name": "docker-compose-product",
  "node_name": "localhost",
  "database": "unhealthy"
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
| `PORT` | HTTP server port | `8090` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://productuser:productpass@localhost:5432/products?sslmode=disable` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTel Collector endpoint (gRPC) | `localhost:4317` |
| `POD_NAME` | Pod name for health check | `docker-compose-product` |
| `NODE_NAME` | Node name for health check | `localhost` |

Copy `.env.example` to `.env` and update as needed.

## Quick Start

### With Docker Compose (Recommended)

The easiest way to run product-service with PostgreSQL:

```bash
# Navigate to product-service directory
cd d:\code\poly-app\product-service

# Start PostgreSQL and OTel Collector first (wait for health check)
docker-compose up -d postgres otel-collector

# Wait for PostgreSQL to initialize (schema + seed data)
Start-Sleep -Seconds 10

# Start product-service
docker-compose up -d product-service

# Check service health
curl http://localhost:8090/healthz -UseBasicParsing

# Get all products
curl http://localhost:8090/products -UseBasicParsing

# Filter by category
curl "http://localhost:8090/products?category=Electronics" -UseBasicParsing
```

**Verify Database Initialization:**
```bash
# Check that products were loaded
docker-compose exec postgres psql -U productuser -d products -c "SELECT COUNT(*) FROM products;"
# Expected output: 16 products

# ⚠️ If count is 0, seed the database manually:
docker exec -i poly-shop-postgres psql -U productuser -d products < product-service/database/reset_and_seed.sql

# View sample products
docker-compose exec postgres psql -U productuser -d products -c "SELECT id, name, price, category FROM products LIMIT 5;"
```

**Access Services:**
- **Product API**: http://localhost:8090
- **Jaeger UI**: http://localhost:16687
- **PostgreSQL**: localhost:5432 (user: productuser, password: productpass, db: products)

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
