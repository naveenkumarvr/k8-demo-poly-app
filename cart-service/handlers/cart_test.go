package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"cart-service/redis"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	redisclient "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// setupTest creates a miniredis instance and returns a configured cart handler
// This provides an isolated Redis environment for each test
func setupTest(t *testing.T) (*CartHandler, *miniredis.Miniredis, func()) {
	// Create miniredis server (in-memory Redis mock)
	mr := miniredis.NewMiniRedis()
	if err := mr.Start(); err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}

	// Create Redis client pointing to miniredis
	rdb := redisclient.NewClient(&redisclient.Options{
		Addr: mr.Addr(),
	})

	// Create logger (use nop logger for tests to avoid output clutter)
	logger := zap.NewNop()

	// Create redis.Client wrapper
	redisClient := &redis.Client{}
	// Use reflection or direct assignment if Client struct is exported
	// For simplicity, we'll create a new InitRedis that accepts a client
	// But since we can't modify redis package in tests, we'll create wrapper

	// Quick workaround: manually create what we need
	ctx := context.Background()
	err := rdb.Ping(ctx).Err()
	require.NoError(t, err, "miniredis should be reachable")

	// Create a test client wrapper
	testClient := &testRedisClient{
		rdb:    rdb,
		logger: logger,
	}

	handler := &CartHandler{
		redisClient: testClient,
		logger:      logger,
	}

	// Cleanup function
	cleanup := func() {
		rdb.Close()
		mr.Close()
	}

	return handler, mr, cleanup
}

// testRedisClient wraps the Redis client for testing
type testRedisClient struct {
	rdb    *redisclient.Client
	logger *zap.Logger
}

func (c *testRedisClient) AddItem(ctx context.Context, userID, productID string, quantity int) error {
	key := "cart:" + userID
	return c.rdb.HIncrBy(ctx, key, productID, int64(quantity)).Err()
}

func (c *testRedisClient) GetCart(ctx context.Context, userID string) ([]redis.CartItem, error) {
	key := "cart:" + userID
	result, err := c.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	items := make([]redis.CartItem, 0, len(result))
	for productID, quantityStr := range result {
		// Parse quantity
		var quantity int
		_, _ = fmt.Sscanf(quantityStr, "%d", &quantity)
		items = append(items, redis.CartItem{
			ProductID: productID,
			Quantity:  quantity,
		})
	}
	return items, nil
}

func (c *testRedisClient) ClearCart(ctx context.Context, userID string) error {
	key := "cart:" + userID
	return c.rdb.Del(ctx, key).Err()
}

func TestAddItem(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should add item to empty cart", func(t *testing.T) {
		handler, _, cleanup := setupTest(t)
		defer cleanup()

		router := gin.New()
		router.POST("/v1/cart/:user_id", handler.AddItem)

		reqBody := AddItemRequest{
			ProductID: "prod-123",
			Quantity:  2,
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/cart/user-1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response CartResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "user-1", response.UserID)
		assert.Equal(t, 1, response.TotalItems)
		assert.Equal(t, "prod-123", response.Items[0].ProductID)
		assert.Equal(t, 2, response.Items[0].Quantity)
	})

	t.Run("should increment existing item quantity", func(t *testing.T) {
		handler, _, cleanup := setupTest(t)
		defer cleanup()

		router := gin.New()
		router.POST("/v1/cart/:user_id", handler.AddItem)

		// Add item first time
		reqBody1 := AddItemRequest{ProductID: "prod-123", Quantity: 2}
		body1, _ := json.Marshal(reqBody1)
		req1, _ := http.NewRequest("POST", "/v1/cart/user-1", bytes.NewBuffer(body1))
		req1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// Add same item again
		reqBody2 := AddItemRequest{ProductID: "prod-123", Quantity: 3}
		body2, _ := json.Marshal(reqBody2)
		req2, _ := http.NewRequest("POST", "/v1/cart/user-1", bytes.NewBuffer(body2))
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusOK, w2.Code)

		var response CartResponse
		json.Unmarshal(w2.Body.Bytes(), &response)

		assert.Equal(t, 1, response.TotalItems)
		assert.Equal(t, 5, response.Items[0].Quantity) // 2 + 3
	})

	t.Run("should reject invalid quantity", func(t *testing.T) {
		handler, _, cleanup := setupTest(t)
		defer cleanup()

		router := gin.New()
		router.POST("/v1/cart/:user_id", handler.AddItem)

		reqBody := AddItemRequest{
			ProductID: "prod-123",
			Quantity:  0, // Invalid
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/cart/user-1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should reject missing product_id", func(t *testing.T) {
		handler, _, cleanup := setupTest(t)
		defer cleanup()

		router := gin.New()
		router.POST("/v1/cart/:user_id", handler.AddItem)

		reqBody := AddItemRequest{
			Quantity: 2,
			// ProductID missing
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/cart/user-1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetCart(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return empty cart", func(t *testing.T) {
		handler, _, cleanup := setupTest(t)
		defer cleanup()

		router := gin.New()
		router.GET("/v1/cart/:user_id", handler.GetCart)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/cart/user-1", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response CartResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, "user-1", response.UserID)
		assert.Equal(t, 0, response.TotalItems)
		assert.Empty(t, response.Items)
	})

	t.Run("should return cart with items", func(t *testing.T) {
		handler, _, cleanup := setupTest(t)
		defer cleanup()

		// Add items first
		ctx := context.Background()
		handler.redisClient.AddItem(ctx, "user-1", "prod-1", 2)
		handler.redisClient.AddItem(ctx, "user-1", "prod-2", 3)

		router := gin.New()
		router.GET("/v1/cart/:user_id", handler.GetCart)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/cart/user-1", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response CartResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, "user-1", response.UserID)
		assert.Equal(t, 2, response.TotalItems)
	})
}

func TestDeleteCart(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should clear cart successfully", func(t *testing.T) {
		handler, _, cleanup := setupTest(t)
		defer cleanup()

		// Add items first
		ctx := context.Background()
		handler.redisClient.AddItem(ctx, "user-1", "prod-1", 2)

		router := gin.New()
		router.DELETE("/v1/cart/:user_id", handler.DeleteCart)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/v1/cart/user-1", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify cart is empty
		items, _ := handler.redisClient.GetCart(ctx, "user-1")
		assert.Empty(t, items)
	})
}
