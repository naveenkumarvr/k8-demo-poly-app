package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

// CartEvent represents a cart event published by cart-service.
type CartEvent struct {
	Event     string `json:"event"`
	UserID    string `json:"user_id"`
	TotalItems int   `json:"total_items,omitempty"`
}

func main() {
	serviceName := getEnv("SERVICE_NAME", "notification-worker")
	serviceVersion := getEnv("SERVICE_VERSION", "1.0.0")
	environment := getEnv("ENVIRONMENT", "development")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	channel := getEnv("EVENT_CHANNEL", "cart-events")
	podName := getEnv("POD_NAME", "local-dev")
	nodeName := getEnv("NODE_NAME", "local-dev")

	// Structured JSON logger to stdout (12-Factor App)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})).With(
		slog.String("service", serviceName),
		slog.String("version", serviceVersion),
		slog.String("environment", environment),
		slog.String("pod_name", podName),
		slog.String("node_name", nodeName),
	)
	slog.SetDefault(logger)

	logger.Info("Starting notification worker", slog.String("redis_addr", redisAddr), slog.String("channel", channel))

	rdb := redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Verify Redis connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("Failed to connect to Redis", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("Connected to Redis")

	// Subscribe to event channel
	sub := rdb.Subscribe(ctx, channel)
	defer sub.Close()

	logger.Info("Subscribed to Redis Pub/Sub channel", slog.String("channel", channel))

	// Graceful shutdown on SIGTERM/SIGINT
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Event processing loop
	go func() {
		ch := sub.Channel()
		for msg := range ch {
			var event CartEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				logger.Error("Failed to parse cart event", slog.String("payload", msg.Payload), slog.String("error", err.Error()))
				continue
			}

			// Simulated background work: log / notify / analytics
			logger.Info("Processed cart event asynchronously",
				slog.String("event", event.Event),
				slog.String("user_id", event.UserID),
				slog.Int("total_items", event.TotalItems),
			)
		}
	}()

	<-quit
	logger.Info("Shutting down notification worker")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := rdb.Close(); err != nil {
		logger.Error("Error closing Redis connection", slog.String("error", err.Error()))
	}
	_ = shutdownCtx
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
