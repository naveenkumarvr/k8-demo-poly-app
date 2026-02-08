package database

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Client wraps the PostgreSQL connection pool with retry logic
type Client struct {
	pool   *pgxpool.Pool
	tracer trace.Tracer
}

// Config holds database configuration
type Config struct {
	DatabaseURL string
	MaxRetries  int
	ServiceName string
}

// NewClient creates a new database client with connection pooling and retry logic
// It implements exponential backoff similar to cart-service Redis client
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 5
	}

	var pool *pgxpool.Pool
	var err error

	// Exponential backoff retry logic
	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		// Parse config and configure connection pool
		config, parseErr := pgxpool.ParseConfig(cfg.DatabaseURL)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse database URL: %w", parseErr)
		}

		// Connection pool settings
		config.MaxConns = 25                      // Maximum number of connections
		config.MinConns = 5                       // Minimum number of idle connections
		config.MaxConnLifetime = 30 * time.Minute // Connection lifetime
		config.MaxConnIdleTime = 5 * time.Minute  // Idle connection timeout

		// Attempt to create connection pool
		pool, err = pgxpool.NewWithConfig(ctx, config)
		if err == nil {
			// Verify connection with ping
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err = pool.Ping(pingCtx)
			cancel()

			if err == nil {
				// Success!
				break
			}
		}

		// Connection failed, retry with exponential backoff
		if attempt < cfg.MaxRetries {
			// Calculate delay: min(100ms * 2^attempt, 2s) with ±10% jitter
			baseDelay := 100 * time.Millisecond
			maxDelay := 2 * time.Second
			delay := time.Duration(float64(baseDelay) * float64(uint(1)<<uint(attempt)))
			if delay > maxDelay {
				delay = maxDelay
			}

			// Add jitter (±10%)
			jitter := time.Duration(rand.Float64()*0.2-0.1) * delay
			delay += jitter

			time.Sleep(delay)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", cfg.MaxRetries, err)
	}

	// Initialize OpenTelemetry tracer
	tracer := otel.Tracer(cfg.ServiceName)

	return &Client{
		pool:   pool,
		tracer: tracer,
	}, nil
}

// Ping checks if the database connection is alive
func (c *Client) Ping(ctx context.Context) error {
	ctx, span := c.tracer.Start(ctx, "database.Ping")
	defer span.End()

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err := c.pool.Ping(pingCtx)
	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Bool("db.healthy", false))
		return err
	}

	span.SetAttributes(attribute.Bool("db.healthy", true))
	return nil
}

// Close gracefully closes the database connection pool
func (c *Client) Close() {
	c.pool.Close()
}

// Pool returns the underlying connection pool for use by repositories
func (c *Client) Pool() *pgxpool.Pool {
	return c.pool
}
