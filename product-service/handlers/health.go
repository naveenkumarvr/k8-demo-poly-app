package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthResponse represents the standard health check response
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// Healthz handles the /healthz endpoint
// This is the general health check endpoint
// Returns HTTP 200 if the service is running
func Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:  "ok",
		Service: "product-service",
	})
}

// Ready handles the /ready endpoint
// This is the Kubernetes readiness probe
// Indicates whether the service is ready to accept traffic
// In a real application, this would check:
// - Database connectivity
// - Required service dependencies
// - Cache availability
func Ready(c *gin.Context) {
	// For this demo, we're always ready
	// In production, add actual readiness checks here
	c.JSON(http.StatusOK, HealthResponse{
		Status:  "ready",
		Service: "product-service",
	})
}

// Live handles the /live endpoint
// This is the Kubernetes liveness probe
// Indicates whether the service needs to be restarted
// In a real application, this would check:
// - Memory leaks
// - Deadlocks
// - Critical goroutine failures
func Live(c *gin.Context) {
	// For this demo, we're always alive
	// In production, add actual liveness checks here
	c.JSON(http.StatusOK, HealthResponse{
		Status:  "alive",
		Service: "product-service",
	})
}
