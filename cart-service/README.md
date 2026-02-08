# Cart-Service

Production-ready stateful microservice for managing user shopping carts in the Poly-Shop suite. Built with Go, Redis, OpenTelemetry, and comprehensive observability.

## Overview

Cart-Service provides a REST API for managing user shopping carts with:
- **Redis Backend**: Fast, stateful cart storage using Redis hashes
- **OpenTelemetry**: Full distributed tracing with W3C Trace Context propagation
- **Structured Logging**: JSON logs with trace correlation using Zap
- **Sidecar Pattern**: Log shipping demonstration for centralized logging
- **Graceful Shutdown**: Context-based shutdown handling SIGINT/SIGTERM
- **Production-Ready**: Distroless container image, health checks, retry logic

## Architecture

```mermaid
graph LR
    Client[HTTP Client] -->|W3C Trace Context| CartService[Cart Service]
    CartService -->|OTLP/gRPC| Collector[OTel Collector]
    CartService -->|Redis Protocol| Redis[(Redis)]
    CartService -->|JSON Logs| LogFile[/var/log/app/cart-service.log]
    LogShipper[Log Shipper Sidecar] -.->|tail -F| LogFile
    LogShipper -->|stdout| LogAggregator[Log Aggregator]
    Collector -->|Traces| Jaeger[Jaeger UI]
```

### Redis Data Model

```
Key: "cart:{user_id}"
Type: Hash
Fields:
  {product_id_1}: {quantity_1}
  {product_id_2}: {quantity_2}
  ...
```

**Example**:
```
cart:user-123
  prod-abc: 2
  prod-xyz: 5
```

### Project Structure

```
cart-service/
├── main.go                 # Application entry point
├── Dockerfile              # Multi-stage Docker build
├── handlers/               # HTTP request handlers (Add, Get, Delete)
├── redis/                  # Redis client and repository implementation
├── middleware/             # Gin middleware (logging, tracing)
├── logger/                 # Structured logging configuration (Zap)
├── telemetry/              # OpenTelemetry trace configuration
├── docker-compose.yml      # Local development stack
└── scripts/                # k6 load testing scripts
```

## API Contract

### Cart Operations

#### Add Item to Cart
```http
POST /v1/cart/:user_id
Content-Type: application/json

{
  "product_id": "prod-123",
  "quantity": 2
}
```

**Response** (200 OK):
```json
{
  "user_id": "user-456",
  "items": [
    {"product_id": "prod-123", "quantity": 2}
  ],
  "total_items": 1
}
```

**Error Codes**:
- `400 Bad Request`: Invalid request body or quantity ≤ 0
- `500 Internal Server Error`: Redis connection failure

#### Get Cart
```http
GET /v1/cart/:user_id
```

**Response** (200 OK):
```json
{
  "user_id": "user-456",
  "items": [
    {"product_id": "prod-123", "quantity": 2},
    {"product_id": "prod-789", "quantity": 1}
  ],
  "total_items": 2
}
```

**Note**: Returns empty cart if user has no items.

#### Delete Cart
```http
DELETE /v1/cart/:user_id
```

**Response** (200 OK):
```json
{
  "message": "Cart cleared successfully",
  "user_id": "user-456"
}
```

### Health Check

#### Healthz
```http
GET /healthz
```

**Response** (200 OK when healthy):
```json
{
  "status": "healthy",
  "service": "cart-service",
  "pod_name": "cart-service-abc123",
  "node_name": "node-1",
  "redis": "healthy"
}
```

**Response** (503 Service Unavailable when unhealthy):
```json
{
  "status": "unhealthy",
  "service": "cart-service",
  "pod_name": "cart-service-abc123",
  "node_name": "node-1",
  "redis": "unhealthy"
}
```

### Stress Test

#### Artificial Load Generator
```http
POST /stress?cpu_iterations=1000&memory_mb=100
```

**Query Parameters**:
- `cpu_iterations` (default: 1000, max: 10000): Number of prime calculation iterations
- `memory_mb` (default: 100, max: 1000): MB of memory to allocate

**Response** (200 OK):
```json
{
  "cpu_iterations": 1000,
  "memory_mb": 100,
  "primes_calculated": 1229,
  "computation_time": "3.456s",
  "message": "Stress test completed successfully"
}
```

**Use Cases**:
- Horizontal Pod Autoscaler (HPA) testing
- Performance profiling
- Resource limit validation

## Local Development

### Prerequisites

- **Go 1.21+**
- **Docker** & **Docker Compose**
- **Redis** (if running locally without Docker)
- **k6** (for load testing)

### Setup

1. **Clone the repository**:
   ```bash
   cd d:\code\poly-app\cart-service
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Copy environment template**:
   ```bash
   cp .env.example .env
   ```

### Running Locally (without Docker)

1. **Start Redis**:
   ```bash
   # Windows (using Docker)
   docker run -d --name redis -p 6379:6379 redis:7-alpine
   ```

2. **Run the service**:
   ```bash
   go run main.go
   ```

3. **Test the API**:
   ```bash
   # Health check
   curl http://localhost:8080/healthz

   # Add item to cart
   curl -X POST http://localhost:8080/v1/cart/user-123 \
     -H "Content-Type: application/json" \
     -d '{"product_id":"prod-456","quantity":2}'

   # Get cart
   curl http://localhost:8080/v1/cart/user-123

   # Delete cart
   curl -X DELETE http://localhost:8080/v1/cart/user-123
   ```

### Running with Docker Compose

The recommended way for local development includes Redis, cart-service, log-shipper sidecar, and Jaeger for traces.

1. **Build and start all services**:
   ```bash
   docker-compose up --build
   ```

2. **View logs from different components**:
   ```bash
   # Cart service logs
   docker-compose logs -f cart-service

   # Log shipper (demonstrates sidecar pattern)
   docker-compose logs -f log-shipper

   # All logs
   docker-compose logs -f
   ```

3. **Access services**:
   - **Cart Service**: http://localhost:8080
   - **Jaeger UI**: http://localhost:16686

4. **Stop services**:
   ```bash
   docker-compose down
   ```

### Running with Docker (Standalone)

1. **Build the Docker image**:
   ```bash
   docker build -t cart-service:latest .
   ```

2. **Run with local Redis**:
   ```bash
   # Start Redis first
   docker run -d --name redis -p 6379:6379 redis:7-alpine

   # Run cart-service
   docker run -p 8080:8080 \
     -e REDIS_ADDR=host.docker.internal:6379 \
     -e ENVIRONMENT=development \
     --name cart-service \
     cart-service:latest
   ```

3. **View logs**:
   ```bash
   docker logs -f cart-service
   ```

4. **Stop and remove**:
   ```bash
   docker stop cart-service redis
   docker rm cart-service redis
   ```

## Testing

### Unit Tests

Tests use `miniredis` (in-memory Redis mock) for isolated testing without external dependencies.

```bash
# Run all tests
go test ./... -v

# Run with coverage
go test ./... -cover -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out -o coverage.html
# Open coverage.html in browser
```

**Expected Output**:
```
?       cart-service                            [no test files]
ok      cart-service/handlers                   2.345s  coverage: 87.5% of statements
ok      cart-service/logger                     0.123s  coverage: 92.3% of statements
ok      cart-service/middleware                 0.089s  coverage: 95.0% of statements
ok      cart-service/redis                      1.567s  coverage: 89.2% of statements
ok      cart-service/telemetry                  0.234s  coverage: 85.0% of statements
```

### Race Condition Detection

```bash
go test ./... -race -count=10
```

### Load Testing with k6

Simulates realistic user journeys: Add items → Get cart → Stress test → Clear cart.

1. **Start the service**:
   ```bash
   docker-compose up -d
   ```

2. **Run k6 load test**:
   ```bash
   k6 run scripts/k6-load-test.js
   ```

3. **Custom base URL**:
   ```bash
   k6 run -e BASE_URL=http://your-service:8080 scripts/k6-load-test.js
   ```

**Expected Results**:
```
✓ add_item: status is 200
✓ add_item: response time < 200ms
✓ get_cart: status is 200
✓ delete_cart: status is 200
✓ stress: response time < 5s
✓ healthz: redis is healthy

checks.........................: 99.8% ✓ 15234   ✗ 31
http_req_duration..............: avg=145ms  p(95)=180ms
http_req_failed................: 0.2%  ✓ 31      ✗ 15234
```

## Stress Endpoint Usage

The `/stress` endpoint generates artificial CPU and memory load for testing.

### Triggering a Stress Test

```bash
# Default parameters (1000 iterations, 100MB)
curl -X POST http://localhost:8080/stress

# Custom parameters
curl -X POST "http://localhost:8080/stress?cpu_iterations=2000&memory_mb=200"
```

### Monitoring Container Metrics

#### Docker Stats
```bash
docker stats cart-service

# Expected output during stress test:
# NAME           CPU %   MEM USAGE / LIMIT   MEM %
# cart-service   95.2%   350MiB / 512MiB     68.4%
```

#### Kubernetes (if deployed)
```bash
# Watch pod metrics
kubectl top pod -l app=cart-service

# Expected output:
# NAME                CPU(cores)   MEMORY(bytes)
# cart-service-abc   850m         320Mi
```

### What to Look For

1. **CPU Spike**: Should reach 80-100% for 3-5 seconds
2. **Memory Increase**: Should allocate the specified MB temporarily
3. **Graceful Recovery**: CPU and memory should return to baseline after completion
4. **No Errors**: The service should remain responsive during stress

### HPA Testing

If using Kubernetes Horizontal Pod Autoscaler:

```bash
# Continuously stress the service
for i in {1..20}; do
  curl -X POST "http://localhost:8080/stress?cpu_iterations=3000&memory_mb=200"
  sleep 2
done

# Watch HPA scale up
kubectl get hpa -w
```

## Observability

### Distributed Tracing

1. **Start Jaeger** (included in docker-compose):
   ```bash
   docker-compose up -d otel-collector
   ```

2. **Generate traces**:
   ```bash
   curl -X POST http://localhost:8080/v1/cart/trace-demo \
     -H "Content-Type: application/json" \
     -d '{"product_id":"trace-test","quantity":1}'
   ```

3. **View traces in Jaeger UI**:
   - Open http://localhost:16686
   - Select service: `cart-service`
   - Click "Find Traces"

**Expected Trace Structure**:
```
HTTP POST /v1/cart/:user_id (parent span)
├── handler.AddItem
│   ├── redis.AddItem
│   │   └── redis: hincrby (otelredis instrumentation)
│   └── redis.GetCart
│       └── redis: hgetall (otelredis instrumentation)
```

### Log Correlation

Logs include `trace_id` for correlation with distributed traces:

```json
{
  "timestamp": "2026-02-08T18:45:23.456Z",
  "level": "info",
  "msg": "HTTP request completed",
  "service": "cart-service",
  "pod_name": "cart-service-abc123",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "method": "POST",
  "path": "/v1/cart/user-123",
  "status": 200,
  "duration": "15.234ms"
}
```

### Sidecar Logging Pattern

The docker-compose setup demonstrates the sidecar pattern:

**View sidecar logs**:
```bash
docker-compose logs -f log-shipper
```

**Expected**: You'll see structured JSON logs from the cart-service being tailed by the sidecar, simulating shipping to a centralized logging system like ELK or Splunk.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVICE_NAME` | `cart-service` | Service name for tracing |
| `SERVICE_VERSION` | `1.0.0` | Service version |
| `ENVIRONMENT` | `development` | Environment (development, production) |
| `PORT` | `8080` | HTTP server port |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `localhost:4317` | OTel collector endpoint |
| `POD_NAME` | `local-dev` | Kubernetes pod name (auto-injected in K8s) |
| `NODE_NAME` | `local-dev` | Kubernetes node name (auto-injected in K8s) |

## Concurrency & Design Patterns

### Exponential Backoff Retry

Redis connection uses exponential backoff for reliability:

```go
// Initial delay: 100ms
// Max delay: 2s
// Max retries: 5
// Jitter: ±10%
delay = min(100ms * 2^attempt, 2s) * (1 ± 10%)
```

### Graceful Shutdown

The service implements context-based graceful shutdown:

1. Receives SIGINT or SIGTERM signal
2. Stops accepting new requests
3. Waits up to 5s for in-flight requests to complete
4. Closes Redis connection
5. Flushes remaining OpenTelemetry spans
6. Exits cleanly

**Testing**:
```bash
# Start service
docker run --name cart-test cart-service:latest

# Send SIGTERM
docker stop cart-test

# Check logs
docker logs cart-test
# Expected: "Shutting down server..." "Server exited cleanly"
```

### OpenTelemetry Span Hierarchy

Every HTTP request creates a parent span, with child spans for Redis operations:

```
HTTP Request (parent)
  └── handler.AddItem
      └── redis.AddItem (child)
          └── redis: hincrby (auto-instrumented by otelredis)
```

Child spans inherit the parent's trace context, enabling end-to-end tracing.

## Troubleshooting

### Service won't start

**Error**: `Failed to connect to Redis`

**Solution**:
- Check Redis is running: `docker ps | grep redis`
- Verify REDIS_ADDR: `echo $REDIS_ADDR`
- Test connection: `redis-cli -h localhost -p 6379 ping`

### Health check fails

**Error**: `{"status":"unhealthy","redis":"unhealthy"}`

**Solution**:
- Check Redis connectivity with 2s timeout
- Verify network connectivity between cart-service and Redis
- Check Redis logs: `docker logs redis`

### Tests fail

**Error**: `miniredis: address already in use`

**Solution**:
- Multiple tests running in parallel, conflict resolved by using unique ports
- If persistent, restart Go test cache: `go clean -testcache`

### Traces not appearing in Jaeger

**Solution**:
- Verify OTel collector is running: `docker-compose ps otel-collector`
- Check OTEL_EXPORTER_OTLP_ENDPOINT is correct
- Ensure service has W3C Trace Context headers (or creates new trace)

## First-Time Setup / Recovery

### Missing go.mod File (Fresh Setup or Accidentally Deleted)

If you don't have a `go.mod` file or accidentally deleted it, follow these steps:

**Symptoms**:
```
go: go.mod file not found in current directory or any parent directory
Docker build fails: "/go.mod": not found
```

**Solution** - Full Recovery Procedure:

```bash
# 1. Navigate to cart-service directory
cd d:\code\poly-app\cart-service

# 2. Initialize Go module (creates go.mod)
go mod init cart-service

# 3. Add all required dependencies
go get github.com/gin-gonic/gin@v1.9.1
go get github.com/redis/go-redis/v9@v9.17.3
go get github.com/redis/go-redis/extra/redisotel/v9@v9.17.3
go get go.uber.org/zap@v1.27.1
go get go.opentelemetry.io/otel@v1.22.0
go get go.opentelemetry.io/otel/sdk@v1.22.0
go get go.opentelemetry.io/otel/trace@v1.22.0
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc@v1.21.0
go get go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin@v0.46.1

# Test dependencies
go get github.com/stretchr/testify@v1.8.4
go get github.com/alicebob/miniredis/v2@v2.36.1

# 4. Clean up and generate go.sum
go mod tidy

# 5. Verify go.mod was created correctly
# Should show: "module cart-service" and "go 1.21"
cat go.mod  # Linux/Mac
type go.mod  # Windows

# 6. Now proceed with Docker build
docker-compose build --no-cache cart-service

# 7. Start services (use sequential startup to avoid DNS issues)
docker-compose down
docker-compose up -d redis otel-collector
Start-Sleep -Seconds 8
docker-compose up -d cart-service log-shipper

# 8. Verify everything is running
docker-compose ps

# 9. Test the service
curl http://localhost:8080/healthz -UseBasicParsing
```

**Quick Recovery** (if you have a backup or know the exact dependencies):
```bash
# Copy this exact content to go.mod
cat > go.mod << 'EOF'
module cart-service

go 1.21

require (
	github.com/alicebob/miniredis/v2 v2.36.1
	github.com/gin-gonic/gin v1.9.1
	github.com/redis/go-redis/extra/redisotel/v9 v9.17.3
	github.com/redis/go-redis/v9 v9.17.3
	github.com/stretchr/testify v1.8.4
	go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin v0.46.1
	go.opentelemetry.io/otel v1.22.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.21.0
	go.opentelemetry.io/otel/sdk v1.22.0
	go.opentelemetry.io/otel/trace v1.22.0
	go.uber.org/zap v1.27.1
)
EOF

# Then run
go mod tidy

# Continue with docker build...
```

**For PowerShell** (Windows):
```powershell
@"
module cart-service

go 1.21

require (
	github.com/alicebob/miniredis/v2 v2.36.1
	github.com/gin-gonic/gin v1.9.1
	github.com/redis/go-redis/extra/redisotel/v9 v9.17.3
	github.com/redis/go-redis/v9 v9.17.3
	github.com/stretchr/testify v1.8.4
	go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin v0.46.1
	go.opentelemetry.io/otel v1.22.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.21.0
	go.opentelemetry.io/otel/sdk v1.22.0
	go.opentelemetry.io/otel/trace v1.22.0
	go.uber.org/zap v1.27.1
)
"@ | Out-File -FilePath go.mod -Encoding UTF8

# Then run
go mod tidy
```

**Verification**:
- `go.mod` file exists with correct version `go 1.21`
- `go.sum` file is generated (contains checksums)
- `go build .` compiles without errors
- Docker build succeeds

---

## Common Deployment Issues & Fixes

### Issue 1: Docker Build Fails - "go.mod: not found"

**Symptoms**:
```
ERROR: failed to compute cache key: "/go.mod": not found
```

**Root Cause**: `.dockerignore` file starts with `*` which blocks ALL files including `go.mod`, `go.sum`, and source code from the Docker build context.

**Solution**:
Update `.dockerignore` to use exclusion-based approach instead of starting with `*`:

```dockerignore
# WRONG - This blocks everything
*
!.dockerignore

# CORRECT - Only exclude what you don't need
.git
.gitignore
.env
*.md
!go.mod
*_test.go
coverage.out
cart-service.exe
docker-compose.yml
scripts/
```

**Verification**: Run `docker build -t cart-service:test .` - should succeed

---

### Issue 2: Docker Build Fails - Invalid Go Version

**Symptoms**:
```
go: module cart-service@upgrade found (v1.25.7), but does not contain package...
exit code: 1
```

**Root Cause**: `go.mod` contains invalid Go version (e.g., `go 1.25.7` when Go is only at 1.21/1.22).

**Solution**:
Update `go.mod` to use correct Go version:

```go
module cart-service

go 1.21  // Changed from go 1.25.7

require (
    // ... dependencies
)
```

Then run:
```bash
go mod tidy
```

**Verification**: Run `go build .` - should compile without errors

---

### Issue 3: Cart-Service Container Fails - "network is unreachable"

**Symptoms**:
```
dial tcp: lookup redis on 192.168.65.7:53: connect: network is unreachable
OpenTelemetry error: connect: network is unreachable after 5 attempts
Container exits immediately after start
```

**Root Cause**: Docker DNS resolver not ready when cart-service starts, even with `depends_on`. The distroless container lacks tools to debug, but the issue is Docker network initialization timing.

**Solution Option 1** (Recommended - Sequential Startup):
Start services in sequence to ensure network/DNS is ready:

```bash
# Stop everything
docker-compose down

# Start infrastructure first
docker-compose up -d redis otel-collector

# Wait 5-10 seconds for network/DNS initialization
Start-Sleep -Seconds 5

# Start application services
docker-compose up -d cart-service log-shipper
```

**Solution Option 2** (Restart After Failure):
If containers fail on first start:

```bash
# Restart the failed service
docker-compose restart cart-service

# Or restart all
docker-compose restart
```

**Solution Option 3** (Force Recreate):
```bash
docker-compose down
docker-compose up -d --force-recreate
```

**Verification**:
```bash
# Check all containers are running
docker-compose ps

# Check cart-service logs (should show successful Redis connection)
docker logs cart-service --tail=20

# Test health endpoint
curl http://localhost:8080/healthz -UseBasicParsing
```

Expected output: `{"status":"healthy","redis":"healthy",...}`

---

### Issue 4: Port 8080 Already in Use

**Symptoms**:
```
Error: Bind for 0.0.0.0:8080 failed: port is already allocated
```

**Root Cause**: Another service (e.g., product-service) is using port 8080.

**Solution**:
```bash
# Find what's using the port
docker ps | Select-String "8080"

# Stop the conflicting container
docker stop <container-name>

# Or change cart-service port in docker-compose.yml
ports:
  - "8081:8080"  # Changed from 8080:8080
```

---

## Quick Start (With Fixes Applied)

To avoid all above issues, follow this exact sequence:

```bash
# 1. Navigate to cart-service directory
cd d:\code\poly-app\cart-service

# 2. Verify go.mod has correct version
# Should be: go 1.21 (NOT 1.25.7 or other invalid version)

# 3. Verify .dockerignore doesn't start with '*'
# Should use exclusion-based approach

# 4. Clean any previous containers
docker-compose down

# 5. Build fresh image
docker-compose build --no-cache cart-service

# 6. Start infrastructure services first
docker-compose up -d redis otel-collector

# 7. Wait for network/DNS initialization
Start-Sleep -Seconds 8

# 8. Start application services
docker-compose up -d cart-service log-shipper

# 9. Verify all services are running
docker-compose ps

# 10. Test the service
curl http://localhost:8080/healthz -UseBasicParsing
```

Expected: 4 containers running (redis, otel-collector, cart-service, log-shipper), health check returns 200 OK.

## License

Poly-Shop Suite - Educational/Demo Purpose

## Author

Developed for Kubernetes Administration & Microservices Practice
