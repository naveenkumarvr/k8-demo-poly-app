package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"product-service/database"
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
	port := getEnv("PORT", "8090")
	podName := getEnv("POD_NAME", "local-dev")
	nodeName := getEnv("NODE_NAME", "local-dev")

	// Initialize structured JSON logger to stdout
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})).With(
		slog.String("service", serviceName),
		slog.String("version", serviceVersion),
		slog.String("environment", environment),
		slog.String("pod_name", podName),
		slog.String("node_name", nodeName),
	)
	slog.SetDefault(logger)

	databaseURL := getEnvOrFile("DATABASE_URL", "")
	if databaseURL == "" {
		logger.Error("DATABASE_URL environment variable or DATABASE_URL_FILE secret mount is required")
		os.Exit(1)
	}

	// Initialize OpenTelemetry tracer
	// The shutdown function ensures all spans are flushed before exit
	shutdown, err := telemetry.InitTracer(telemetry.TracerConfig{
		ServiceName:    serviceName,
		ServiceVersion: serviceVersion,
		Environment:    environment,
		OTLPEndpoint:   otlpEndpoint,
	})
	if err != nil {
		logger.Error("Failed to initialize tracer", slog.String("error", err.Error()))
		os.Exit(1)
	}
	// Ensure tracer shutdown on exit
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdown(ctx); err != nil {
			logger.Error("Error shutting down tracer", slog.String("error", err.Error()))
		}
	}()

	// Initialize database connection
	logger.Info("Connecting to database")
	dbClient, err := database.NewClient(context.Background(), database.Config{
		DatabaseURL: databaseURL,
		MaxRetries:  5,
		ServiceName: serviceName,
	})
	if err != nil {
		logger.Error("Failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dbClient.Close()
	logger.Info("Database connection established")

	// Create repository for database operations
	productRepo := database.NewProductRepository(dbClient)

	// Create product handler with repository
	productHandler := handlers.NewProductHandler(productRepo)

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
	// Prometheus RED metrics middleware (exposes /metrics for scraping)
	router.Use(middleware.PrometheusMiddleware())

	// Register API routes
	// Products endpoint - returns products from PostgreSQL
	// Supports optional ?category=<name> query parameter
	router.GET("/products", productHandler.GetProducts)
	router.GET("/products/:id", productHandler.GetProductByID)

	// Stress endpoint - CPU-intensive computation for HPA testing
	router.GET("/stress", handlers.StressTest)

	// Health check endpoints for Kubernetes probes
	router.GET("/healthz", handlers.Healthz(dbClient))
	router.GET("/ready", handlers.Ready)
	router.GET("/live", handlers.Live)

	// Prometheus metrics endpoint
	router.GET("/metrics", middleware.PrometheusHandler())

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
		logger.Info("Starting HTTP server", slog.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start server", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	// This handles SIGINT (Ctrl+C) and SIGTERM (Docker/Kubernetes stop)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server")

	// Graceful shutdown with 5 second timeout
	// This allows in-flight requests to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("Server exited")
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvOrFile retrieves a config value from an environment variable or from a file
// pointed to by the KEY_FILE environment variable. This supports both direct env vars
// and Kubernetes Secret/ConfigMap file mounts.
func getEnvOrFile(key, defaultValue string) string {
	if value := getEnv(key, ""); value != "" {
		return value
	}
	filePath := getEnv(key+"_FILE", "")
	if filePath == "" {
		return defaultValue
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return defaultValue
	}
	return string(data)
}
