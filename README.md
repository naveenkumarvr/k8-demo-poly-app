# Poly-Shop Microservices Suite

Production-ready multi-service e-commerce platform demonstrating microservices architecture, distributed tracing, and container orchestration.

## 🏗️ Architecture

```mermaid
graph TB
    Client[Client] -->|:8090| PS[Product Service]
    Client -->|:8080| CS[Cart Service]
    
    PS -->|SQL| PG[(PostgreSQL<br/>:5432)]
    CS -->|Commands| Redis[(Redis<br/>:6379)]
    
    PS -->|OTLP gRPC| Jaeger[Jaeger Collector<br/>:4317]
    CS -->|OTLP gRPC| Jaeger
    
    Jaeger -->|UI| UI[Jaeger UI<br/>:16686]
    
    subgraph poly-shop-net
        PS
        CS
        PG
        Redis
        Jaeger
    end
```

## 📦 Services

| Service | Port | Database | Purpose |
|---------|------|----------|---------|
| **product-service** | 8090 | PostgreSQL | Product catalog management with 16 sample products |
| **cart-service** | 8080 | Redis | Shopping cart operations with persistent storage |
| **jaeger** | 16686 | - | Distributed tracing UI |
| **postgres** | 5432 | - | Product database (16 products, 4 categories) |
| **redis** | 6379 | - | Cart cache with LRU eviction |

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- 2GB free RAM
- Ports available: 8080, 8090, 6379, 5432, 16686, 4317

### Start All Services

```bash
# Navigate to poly-app root
cd d:\code\poly-app

# Build and start all services
docker-compose up --build -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f
```

**Expected Output**:
```
✔ Container poly-shop-redis           Healthy
✔ Container poly-shop-postgres        Healthy  
✔ Container poly-shop-jaeger          Started
✔ Container poly-shop-product-service Healthy
✔ Container poly-shop-cart-service    Healthy
```

### Verify Services

```bash
# Product Service health
curl http://localhost:8090/healthz

# Cart Service health
curl http://localhost:8080/healthz

# Get all products (PostgreSQL)
curl http://localhost:8090/products

# Add item to cart (Redis)
curl -X POST http://localhost:8080/v1/cart/user-123 `
  -H "Content-Type: application/json" `
  -d '{\"product_id\":\"prod-456\",\"quantity\":2}'

# Get cart
curl http://localhost:8080/v1/cart/user-123
```

### Access UIs

- **Jaeger Tracing UI**: http://localhost:16686
- **Product API Docs**: http://localhost:8090 (see [product-service/README.md](product-service/README.md))
- **Cart API Docs**: http://localhost:8080 (see [cart-service/README.md](cart-service/README.md))

### Database Access

**PostgreSQL** (Product Catalog):
```bash
docker-compose exec postgres psql -U productuser -d products

# View products
SELECT id, name, price, category FROM products LIMIT 5;
```

**Redis** (Cart Data):
```bash
docker-compose exec redis redis-cli

# View keys
KEYS *

# Get cart data
HGETALL cart:user-123
```

## 🧪 Testing Distributed Tracing

### End-to-End User Journey

```bash
# 1. Browse products (Product Service → PostgreSQL)
curl http://localhost:8090/products?category=Electronics

# 2. Add to cart (Cart Service → Redis)
curl -X POST http://localhost:8080/v1/cart/shopper-001 `
  -H "Content-Type: application/json" `
  -d '{\"product_id\":\"1\",\"quantity\":1}'

# 3. View Jaeger UI
# Open http://localhost:16686
# Select service: product-service or cart-service
# Click "Find Traces"
```

**Trace Example**:
```
HTTP GET /products
├─ repository.GetProductsByCategory (PostgreSQL)
│  ├─ db.query.duration_ms: 12
│  └─ db.result.count: 5
└─ http.status_code: 200
```

## 🛠️ Development Commands

### Service Management

```bash
# Start specific service
docker-compose up -d product-service

# Rebuild after code changes
docker-compose up --build product-service

# View service logs
docker-compose logs -f cart-service

# Restart service
docker-compose restart product-service

# Stop all services
docker-compose down

# Stop and remove volumes (⚠️ deletes data)
docker-compose down -v
```

### Debugging

```bash
# Check service health
docker-compose ps

# Inspect network
docker network inspect poly-shop-net

# Exec into container
docker-compose exec cart-service sh

# View database logs
docker-compose logs postgres

# Check Redis connections
docker-compose exec redis redis-cli CLIENT LIST
```

### Data Management

```bash
# Backup PostgreSQL
docker-compose exec postgres pg_dump -U productuser products > backup.sql

# Backup Redis
docker-compose exec redis redis-cli SAVE
docker cp poly-shop-redis:/data/dump.rdb ./backup-redis.rdb

# Clear cart data
docker-compose exec redis redis-cli FLUSHDB
```

## 📊 Monitoring & Observability

### Health Checks

All services expose `/healthz` endpoints:

**Product Service**:
```json
{
  "status": "healthy",
  "service": "product-service",
  "database": "healthy",
  "pod_name": "poly-shop-product",
  "node_name": "docker-compose"
}
```

**Cart Service**:
```json
{
  "status": "healthy",
  "service": "cart-service",
  "redis": "healthy",
  "pod_name": "poly-shop-cart",
  "node_name": "docker-compose"
}
```

### Jaeger Traces

**Access**: http://localhost:16686

**Services Available**:
- `product-service` - Product catalog operations
- `cart-service` - Cart CRUD operations

**Key Spans**:
- `repository.GetAllProducts` - Database queries
- `repository.GetProductsByCategory` - Filtered queries
- `redis.AddItem` - Cart operations
- `redis.GetCart` - Cart retrieval

### Resource Usage

```bash
# Monitor container stats
docker stats

# Expected usage (approximate):
# product-service:  20-30 MB RAM
# cart-service:     15-25 MB RAM
# postgres:         40-60 MB RAM
# redis:            10-20 MB RAM (maxmemory: 256MB)
# jaeger:           80-120 MB RAM
```

## 🔧 Configuration

### Environment Variables

**Product Service**:
- `PORT=8090`
- `DATABASE_URL=postgres://productuser:productpass@postgres:5432/products?sslmode=disable`
- `OTEL_EXPORTER_OTLP_ENDPOINT=jaeger:4317`

**Cart Service**:
- `PORT=8080`
- `REDIS_ADDR=redis:6379`
- `OTEL_EXPORTER_OTLP_ENDPOINT=jaeger:4317`

### Network Configuration

**Network**: `poly-shop-net`
- **Type**: Bridge
- **Subnet**: 172.25.0.0/16
- **DNS**: Service names resolve to container IPs

### Persistent Volumes

- `poly-shop-postgres-data` - Product database (survives restarts)
- `poly-shop-redis-data` - Cart data with AOF persistence

## 📁 Project Structure

```
poly-app/
├── docker-compose.yml          # Master orchestration
├── README.md                   # This file
├── product-service/
│   ├── Dockerfile
│   ├── docker-compose.yml      # Standalone config
│   ├── README.md
│   ├── WALKTHROUGH.md
│   ├── database/
│   │   ├── schema.sql          # Auto-loaded on init
│   │   └── seed.sql            # 16 sample products
│   └── ...
└── cart-service/
    ├── Dockerfile
    ├── docker-compose.yml      # Standalone config
    ├── README.md
    ├── walkthrough.md
    └── ...
```

## 🚨 Troubleshooting

### Services Won't Start

**Issue**: Containers exit or fail health checks

**Solutions**:
1. Check logs: `docker-compose logs <service-name>`
2. Verify ports not in use: `netstat -an | findstr "8080 8090 6379 5432"`
3. Ensure sufficient resources (2GB RAM minimum)
4. Rebuild: `docker-compose down && docker-compose up --build -d`

### Database Connection Errors

**Issue**: `connection refused` or `dial tcp: lookup postgres`

**Solutions**:
1. Wait for health checks: `docker-compose ps` (should show "healthy")
2. Check network: `docker network inspect poly-shop-net`
3. Verify DNS resolution: `docker-compose exec product-service ping postgres`

### Traces Not Appearing in Jaeger

**Issue**: No traces in Jaeger UI

**Solutions**:
1. Verify Jaeger is running: `docker-compose ps jaeger`
2. Check OTLP endpoint: Services should use `jaeger:4317`
3. Generate traffic: Make API requests to services
4. Check logs: `docker-compose logs jaeger | grep -i error`

### Redis/PostgreSQL Data Loss

**Issue**: Data disappears after restart

**Solutions**:
1. Don't use `docker-compose down -v` (removes volumes)
2. Check volumes exist: `docker volume ls | grep poly-shop`
3. Enable AOF persistence (already configured for Redis)

## 🔐 Security Notes

- Default credentials are for **development only**
- PostgreSQL: `productuser` / `productpass`
- No SSL/TLS (use `sslmode=disable`)
- Services run as non-root (distroless images)
- No exposed management interfaces

**For Production**:
- Change all default passwords
- Enable SSL/TLS
- Add network policies
- Implement authentication/authorization
- Use secrets management

## 📚 Documentation

- **Product Service**: [product-service/README.md](product-service/README.md)
- **Cart Service**: [cart-service/README.md](cart-service/README.md)
- **Product Walkthrough**: [product-service/WALKTHROUGH.md](product-service/WALKTHROUGH.md)
- **Cart Walkthrough**: [cart-service/walkthrough.md](cart-service/walkthrough.md)

## 🎯 Use Cases

### Local Development
```bash
docker-compose up -d
# Develop against live services
```

### Integration Testing
```bash
docker-compose up -d
# Run test suites against full stack
```

### Demo/Presentation
```bash
docker-compose up -d
# Show microservices, tracing, databases
```

###Kubernetes Practice
```bash
# Learn concepts locally before K8s deployment
# Both services have K8s manifests in their docs
```

## 🏁 Next Steps

- [ ] Add API Gateway (nginx/traefik)
- [ ] Implement service mesh (Istio/Linkerd)
- [ ] Add metrics (Prometheus + Grafana)
- [ ] Create Kubernetes manifests
- [ ] Add CI/CD pipelines
- [ ] Implement rate limiting
- [ ] Add authentication service

---

**Version**: 1.0.0  
**Services**: 2 (Product + Cart)  
**Infrastructure**: Redis + PostgreSQL + Jaeger  
**Network**: Custom bridge (poly-shop-net)
