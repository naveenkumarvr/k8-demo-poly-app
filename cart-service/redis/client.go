package redis

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Client wraps the Redis client with additional functionality
type Client struct {
	rdb    *redis.Client
	logger *zap.Logger
}

// RetryConfig holds configuration for exponential backoff retry logic
type RetryConfig struct {
	InitialDelay time.Duration // Starting delay (e.g., 100ms)
	MaxDelay     time.Duration // Maximum delay (e.g., 2s)
	MaxRetries   int           // Maximum number of retries (e.g., 5)
	JitterPct    float64       // Jitter percentage (e.g., 0.1 for ±10%)
}

// DefaultRetryConfig returns the default retry configuration
// Initial delay: 100ms, Max delay: 2s, Max retries: 5, Jitter: ±10%
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     2 * time.Second,
		MaxRetries:   5,
		JitterPct:    0.1,
	}
}

// InitRedis initializes a Redis client with connection pooling and instrumentation
// The client is instrumented with OpenTelemetry for automatic span creation
// Connection is verified by pinging Redis with retry logic
func InitRedis(ctx context.Context, addr string, logger *zap.Logger) (*Client, error) {
	// Create Redis client with connection pool settings
	rdb := redis.NewClient(&redis.Options{
		Addr:            addr,
		Password:        "", // No password for local development
		DB:              0,  // Use default DB
		MaxRetries:      3,  // Automatic retry for failed commands
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		PoolSize:        10,              // Maximum number of socket connections
		MinIdleConns:    2,               // Minimum number of idle connections
		ConnMaxIdleTime: 5 * time.Minute, // Close idle connections after this duration
	})

	// Add OpenTelemetry instrumentation
	// This automatically creates child spans for all Redis operations (HGET, HSET, etc.)
	// Each Redis command will appear as a child span in the trace
	if err := redisotel.InstrumentTracing(rdb); err != nil {
		return nil, fmt.Errorf("failed to instrument Redis with OpenTelemetry: %w", err)
	}

	// Verify connection with retry logic
	retryConfig := DefaultRetryConfig()
	if err := pingWithRetry(ctx, rdb, retryConfig, logger); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis at %s after %d retries: %w", addr, retryConfig.MaxRetries, err)
	}

	logger.Info("Redis client initialized successfully",
		zap.String("addr", addr),
		zap.Int("pool_size", 10),
		zap.Duration("max_idle_time", 5*time.Minute),
	)

	return &Client{
		rdb:    rdb,
		logger: logger,
	}, nil
}

// pingWithRetry attempts to ping Redis with exponential backoff retry logic
// Implements: Starting delay 100ms, max delay 2s, max 5 retries, ±10% jitter
func pingWithRetry(ctx context.Context, rdb *redis.Client, config RetryConfig, logger *zap.Logger) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Try to ping Redis
		err := rdb.Ping(ctx).Err()
		if err == nil {
			if attempt > 0 {
				logger.Info("Redis connection successful after retry",
					zap.Int("attempts", attempt+1),
				)
			}
			return nil
		}

		lastErr = err

		// If this was the last attempt, don't wait
		if attempt == config.MaxRetries {
			break
		}

		// Calculate exponential backoff delay
		// Formula: min(initialDelay * 2^attempt, maxDelay)
		delay := time.Duration(float64(config.InitialDelay) * math.Pow(2, float64(attempt)))
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}

		// Add jitter to prevent thundering herd
		// Jitter range: delay * (1 ± jitterPct)
		jitter := 1.0 + (rand.Float64()*2-1)*config.JitterPct
		delay = time.Duration(float64(delay) * jitter)

		logger.Warn("Redis connection failed, retrying with exponential backoff",
			zap.Error(err),
			zap.Int("attempt", attempt+1),
			zap.Int("max_retries", config.MaxRetries),
			zap.Duration("retry_delay", delay),
		)

		// Wait before retrying
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return lastErr
}

// GetClient returns the underlying Redis client
// This is useful for operations not wrapped by the Client methods
func (c *Client) GetClient() *redis.Client {
	return c.rdb
}

// Ping checks if Redis is reachable
// This is used by health check endpoints
func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// Close closes the Redis connection
// Should be called during graceful shutdown
func (c *Client) Close() error {
	c.logger.Info("Closing Redis connection")
	return c.rdb.Close()
}
