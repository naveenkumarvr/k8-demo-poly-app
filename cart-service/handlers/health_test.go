package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cart-service/redis"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	redisclient "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// setupHealthTest creates a test environment with Redis client
func setupHealthTest(t *testing.T) (*HealthHandler, *miniredis.Miniredis, func()) {
	mr := miniredis.NewMiniRedis()
	if err := mr.Start(); err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}

	rdb := redisclient.NewClient(&redisclient.Options{
		Addr: mr.Addr(),
	})

	logger := zap.NewNop()

	// Create a test wrapper
	testClient := &testHealthRedisClient{
		rdb:    rdb,
		logger: logger,
	}

	handler := &HealthHandler{
		redisClient: testClient,
		logger:      logger,
		podName:     "test-pod",
		nodeName:    "test-node",
	}

	cleanup := func() {
		rdb.Close()
		mr.Close()
	}

	return handler, mr, cleanup
}

// testHealthRedisClient wraps Redis client for health tests
type testHealthRedisClient struct {
	rdb    *redisclient.Client
	logger *zap.Logger
}

func (c *testHealthRedisClient) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

func (c *testHealthRedisClient) AddItem(ctx context.Context, userID, productID string, quantity int) error {
	return nil
}

func (c *testHealthRedisClient) GetCart(ctx context.Context, userID string) ([]redis.CartItem, error) {
	return nil, nil
}

func (c *testHealthRedisClient) ClearCart(ctx context.Context, userID string) error {
	return nil
}

func TestHealthz(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return healthy when Redis is reachable", func(t *testing.T) {
		handler, _, cleanup := setupHealthTest(t)
		defer cleanup()

		router := gin.New()
		router.GET("/healthz", handler.Healthz)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/healthz", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response HealthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "healthy", response.Status)
		assert.Equal(t, "cart-service", response.Service)
		assert.Equal(t, "test-pod", response.PodName)
		assert.Equal(t, "test-node", response.NodeName)
		assert.Equal(t, "healthy", response.Redis)
	})

	t.Run("should return unhealthy when Redis is down", func(t *testing.T) {
		handler, mr, cleanup := setupHealthTest(t)
		defer cleanup()

		// Stop miniredis to simulate Redis being down
		mr.Close()

		router := gin.New()
		router.GET("/healthz", handler.Healthz)

		// Use a short timeout context
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/healthz", nil)
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		req = req.WithContext(ctx)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)

		var response HealthResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, "unhealthy", response.Status)
		assert.Equal(t, "unhealthy", response.Redis)
	})
}
