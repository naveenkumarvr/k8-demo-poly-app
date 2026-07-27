package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// redisPinger is the minimal Redis interface required by health checks
type redisPinger interface {
	Ping(ctx context.Context) error
}

// HealthHandler holds dependencies for health check handlers
type HealthHandler struct {
	redisClient redisPinger
	logger      *zap.Logger
	podName     string
	nodeName    string
}

// HealthResponse represents the response for health check endpoints
type HealthResponse struct {
	Status   string `json:"status"`
	Service  string `json:"service"`
	PodName  string `json:"pod_name"`
	NodeName string `json:"node_name"`
	Redis    string `json:"redis,omitempty"`
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(redisClient redisPinger, logger *zap.Logger, podName, nodeName string) *HealthHandler {
	return &HealthHandler{
		redisClient: redisClient,
		logger:      logger,
		podName:     podName,
		nodeName:    nodeName,
	}
}

// Healthz handles GET /healthz
// Kubernetes liveness/readiness probe that checks Redis connectivity
// Returns 200 OK if Redis is reachable, 503 Service Unavailable otherwise
func (h *HealthHandler) Healthz(c *gin.Context) {
	// Create a context with timeout for Redis ping
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	// Check Redis connectivity
	redisStatus := "healthy"
	err := h.redisClient.Ping(ctx)
	if err != nil {
		redisStatus = "unhealthy"
		h.logger.Error("Health check failed: Redis unreachable",
			zap.Error(err),
		)

		c.JSON(http.StatusServiceUnavailable, HealthResponse{
			Status:   "unhealthy",
			Service:  "cart-service",
			PodName:  h.podName,
			NodeName: h.nodeName,
			Redis:    redisStatus,
		})
		return
	}

	// All checks passed
	c.JSON(http.StatusOK, HealthResponse{
		Status:   "healthy",
		Service:  "cart-service",
		PodName:  h.podName,
		NodeName: h.nodeName,
		Redis:    redisStatus,
	})
}

// Ready handles GET /ready
// Kubernetes readiness probe: returns 200 only when Redis is reachable
func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.redisClient.Ping(ctx); err != nil {
		h.logger.Warn("Readiness probe failed: Redis unreachable", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, HealthResponse{
			Status:   "not_ready",
			Service:  "cart-service",
			PodName:  h.podName,
			NodeName: h.nodeName,
			Redis:    "unhealthy",
		})
		return
	}

	c.JSON(http.StatusOK, HealthResponse{
		Status:   "ready",
		Service:  "cart-service",
		PodName:  h.podName,
		NodeName: h.nodeName,
		Redis:    "healthy",
	})
}

// Live handles GET /live
// Kubernetes liveness probe: returns 200 if the process is alive
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:   "alive",
		Service:  "cart-service",
		PodName:  h.podName,
		NodeName: h.nodeName,
	})
}
