package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cart-service/handlers"
	"cart-service/logger"
	"cart-service/middleware"
	"cart-service/redis"
	"cart-service/telemetry"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Load configuration from environment variables
	serviceName := getEnv("SERVICE_NAME", "cart-service")
	serviceVersion := getEnv("SERVICE_VERSION", "1.0.0")
	environment := getEnv("ENVIRONMENT", "development")
	otlpEndpoint := getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	port := getEnv("PORT", "8080")

	// Kubernetes pod metadata (defaults to "local-dev" for local testing)
	podName := getEnv("POD_NAME", "local-dev")
	nodeName := getEnv("NODE_NAME", "local-dev")

	// Initialize logger first so we can use it for subsequent initialization
	// This creates structured JSON logs to stdout and /var/log/app/cart-service.log
	zapLogger, err := logger.InitLogger(serviceName, podName, nodeName, environment)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync() // Flush any buffered log entries

	zapLogger.Info("Starting cart-service",
		zap.String("service_name", serviceName),
		zap.String("version", serviceVersion),
		zap.String("environment", environment),
		zap.String("pod_name", podName),
		zap.String("node_name", nodeName),
	)

	// Initialize OpenTelemetry tracer
	// The shutdown function ensures all spans are flushed before exit
	shutdownTracer, err := telemetry.InitTracer(telemetry.TracerConfig{
		ServiceName:    serviceName,
		ServiceVersion: serviceVersion,
		Environment:    environment,
		OTLPEndpoint:   otlpEndpoint,
	})
	if err != nil {
		zapLogger.Fatal("Failed to initialize tracer", zap.Error(err))
	}
	// Ensure tracer shutdown on exit to flush remaining spans
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTracer(ctx); err != nil {
			zapLogger.Error("Error shutting down tracer", zap.Error(err))
		}
	}()

	// Initialize Redis client with retry logic
	// This uses exponential backoff for connection reliability
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	redisClient, err := redis.InitRedis(ctx, redisAddr, zapLogger)
	if err != nil {
		zapLogger.Fatal("Failed to initialize Redis client", zap.Error(err))
	}
	// Ensure Redis connection is closed on exit
	defer func() {
		if err := redisClient.Close(); err != nil {
			zapLogger.Error("Error closing Redis connection", zap.Error(err))
		}
	}()

	// Set Gin mode based on environment
	if environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.New()

	// Add middleware in order of execution:
	// 1. Recovery middleware - recovers from panics and returns 500
	router.Use(gin.Recovery())

	// 2. OpenTelemetry tracing middleware - creates parent span and extracts W3C Trace Context
	// This must come before logging middleware to ensure trace_id is available in logs
	router.Use(middleware.TracingMiddleware(serviceName))

	// 3. Zap logging middleware - logs all requests with trace_id correlation
	router.Use(middleware.ZapMiddleware(zapLogger))

	// Initialize handlers with dependencies
	cartHandler := handlers.NewCartHandler(redisClient, zapLogger)
	healthHandler := handlers.NewHealthHandler(redisClient, zapLogger, podName, nodeName)
	stressHandler := handlers.NewStressHandler(zapLogger)

	// Register API routes
	// Cart operations - v1 API versioning
	v1 := router.Group("/v1")
	{
		v1.POST("/cart/:user_id", cartHandler.AddItem)
		v1.GET("/cart/:user_id", cartHandler.GetCart)
		v1.DELETE("/cart/:user_id", cartHandler.DeleteCart)
	}

	// Health check endpoint for Kubernetes liveness/readiness probes
	router.GET("/healthz", healthHandler.Healthz)

	// Stress test endpoint for HPA testing and performance profiling
	router.POST("/stress", stressHandler.StressTest)

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
	// This allows us to handle OS signals while the server runs
	go func() {
		zapLogger.Info("Starting HTTP server",
			zap.String("port", port),
			zap.String("environment", environment),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	// This handles SIGINT (Ctrl+C) and SIGTERM (Docker/Kubernetes stop)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	// This allows in-flight requests and Redis operations to complete
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		zapLogger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	zapLogger.Info("Server exited cleanly")
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
