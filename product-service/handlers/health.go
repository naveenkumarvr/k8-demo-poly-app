package handlers

import (
	"net/http"
	"os"

	"product-service/database"

	"github.com/gin-gonic/gin"
)

// HealthResponse represents the standard health check response
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// Healthz is a simple health check endpoint
// Returns 200 OK if the service is healthy
// In production, this should check:
// - Application status
// - Database connectivity
// - External service connectivity
func Healthz(dbClient *database.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check database health
		dbStatus := "healthy"
		statusCode := http.StatusOK

		if dbClient != nil {
			if err := dbClient.Ping(c.Request.Context()); err != nil {
				dbStatus = "unhealthy"
				statusCode = http.StatusServiceUnavailable
			}
		}

		response := gin.H{
			"status":    "healthy",
			"service":   "product-service",
			"pod_name":  os.Getenv("POD_NAME"),
			"node_name": os.Getenv("NODE_NAME"),
			"database":  dbStatus,
		}

		if statusCode != http.StatusOK {
			response["status"] = "unhealthy"
		}

		c.JSON(statusCode, response)
	}
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
