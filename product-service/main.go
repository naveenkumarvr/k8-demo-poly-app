package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"product-service/handlers"
	"product-service/middleware"
	"product-service/telemetry"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration from environment variables
	serviceName := getEnv("SERVICE_NAME", "product-service")
	serviceVersion := getEnv("SERVICE_VERSION", "1.0.0")
	environment := getEnv("ENVIRONMENT", "development")
	otlpEndpoint := getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")
	port := getEnv("PORT", "8080")

	// Initialize OpenTelemetry tracer
	// The shutdown function ensures all spans are flushed before exit
	shutdown, err := telemetry.InitTracer(telemetry.TracerConfig{
		ServiceName:    serviceName,
		ServiceVersion: serviceVersion,
		Environment:    environment,
		OTLPEndpoint:   otlpEndpoint,
	})
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	// Ensure tracer shutdown on exit
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer: %v", err)
		}
	}()

	// Set Gin mode based on environment
	if environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.New()

	// Add middleware
	// Recovery middleware recovers from panics and returns 500
	router.Use(gin.Recovery())
	// Logger middleware logs all HTTP requests
	router.Use(gin.Logger())
	// OpenTelemetry tracing middleware
	// This must be added after Recovery and Logger to ensure proper trace context
	router.Use(middleware.TracingMiddleware(serviceName))

	// Register API routes
	// Products endpoint - returns mock inventory data
	router.GET("/products", handlers.GetProducts)

	// Stress endpoint - CPU-intensive computation for HPA testing
	router.GET("/stress", handlers.StressTest)

	// Health check endpoints for Kubernetes probes
	router.GET("/healthz", handlers.Healthz)
	router.GET("/ready", handlers.Ready)
	router.GET("/live", handlers.Live)

	// Create HTTP server with timeouts
	// These timeouts prevent resource exhaustion from slow clients
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine to enable graceful shutdown
	go func() {
		log.Printf("Starting %s on port %s (environment: %s)", serviceName, port, environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	// This handles SIGINT (Ctrl+C) and SIGTERM (Docker/Kubernetes stop)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	// This allows in-flight requests to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
