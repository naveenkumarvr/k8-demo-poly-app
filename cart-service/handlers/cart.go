package handlers

import (
	"net/http"

	"cart-service/redis"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// AddItemRequest represents the request body for adding an item to cart
type AddItemRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

// CartItem represents a single item in the cart response
type CartItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

// CartResponse represents the response for cart operations
type CartResponse struct {
	UserID     string     `json:"user_id"`
	Items      []CartItem `json:"items"`
	TotalItems int        `json:"total_items"`
}

// CartHandler holds dependencies for cart handlers
type CartHandler struct {
	redisClient *redis.Client
	logger      *zap.Logger
}

// NewCartHandler creates a new cart handler
func NewCartHandler(redisClient *redis.Client, logger *zap.Logger) *CartHandler {
	return &CartHandler{
		redisClient: redisClient,
		logger:      logger,
	}
}

// AddItem handles POST /v1/cart/:user_id
// Adds an item to the user's cart or increments quantity if it already exists
func (h *CartHandler) AddItem(c *gin.Context) {
	// Extract trace context for creating child spans
	ctx := c.Request.Context()
	tracer := otel.Tracer("cart-service")
	ctx, span := tracer.Start(ctx, "handler.AddItem")
	defer span.End()

	// Extract user_id from path parameter
	userID := c.Param("user_id")
	if userID == "" {
		span.SetStatus(codes.Error, "Missing user_id")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id is required",
		})
		return
	}

	span.SetAttributes(attribute.String("user_id", userID))

	// Parse request body
	var req AddItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.SetStatus(codes.Error, "Invalid request body")
		span.RecordError(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	span.SetAttributes(
		attribute.String("product_id", req.ProductID),
		attribute.Int("quantity", req.Quantity),
	)

	// Add item to cart via Redis
	if err := h.redisClient.AddItem(ctx, userID, req.ProductID, req.Quantity); err != nil {
		span.SetStatus(codes.Error, "Failed to add item")
		span.RecordError(err)
		h.logger.Error("Failed to add item to cart",
			zap.String("user_id", userID),
			zap.String("product_id", req.ProductID),
			zap.Int("quantity", req.Quantity),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add item to cart",
		})
		return
	}

	// Get updated cart to return in response
	items, err := h.redisClient.GetCart(ctx, userID)
	if err != nil {
		span.SetStatus(codes.Error, "Failed to retrieve cart")
		span.RecordError(err)
		// Item was added but we failed to retrieve the cart
		// Return success but with a warning
		c.JSON(http.StatusOK, gin.H{
			"message": "Item added successfully",
			"warning": "Failed to retrieve updated cart",
		})
		return
	}

	// Convert to response format
	responseItems := make([]CartItem, len(items))
	for i, item := range items {
		responseItems[i] = CartItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	response := CartResponse{
		UserID:     userID,
		Items:      responseItems,
		TotalItems: len(responseItems),
	}

	span.SetStatus(codes.Ok, "Item added successfully")
	span.SetAttributes(attribute.Int("total_items", len(responseItems)))

	c.JSON(http.StatusOK, response)
}

// GetCart handles GET /v1/cart/:user_id
// Returns all items in the user's cart
func (h *CartHandler) GetCart(c *gin.Context) {
	ctx := c.Request.Context()
	tracer := otel.Tracer("cart-service")
	ctx, span := tracer.Start(ctx, "handler.GetCart")
	defer span.End()

	userID := c.Param("user_id")
	if userID == "" {
		span.SetStatus(codes.Error, "Missing user_id")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id is required",
		})
		return
	}

	span.SetAttributes(attribute.String("user_id", userID))

	// Get cart items from Redis
	items, err := h.redisClient.GetCart(ctx, userID)
	if err != nil {
		span.SetStatus(codes.Error, "Failed to get cart")
		span.RecordError(err)
		h.logger.Error("Failed to get cart",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve cart",
		})
		return
	}

	// Convert to response format
	responseItems := make([]CartItem, len(items))
	for i, item := range items {
		responseItems[i] = CartItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	response := CartResponse{
		UserID:     userID,
		Items:      responseItems,
		TotalItems: len(responseItems),
	}

	span.SetStatus(codes.Ok, "Cart retrieved successfully")
	span.SetAttributes(attribute.Int("total_items", len(responseItems)))

	c.JSON(http.StatusOK, response)
}

// DeleteCart handles DELETE /v1/cart/:user_id
// Clears all items from the user's cart
func (h *CartHandler) DeleteCart(c *gin.Context) {
	ctx := c.Request.Context()
	tracer := otel.Tracer("cart-service")
	ctx, span := tracer.Start(ctx, "handler.DeleteCart")
	defer span.End()

	userID := c.Param("user_id")
	if userID == "" {
		span.SetStatus(codes.Error, "Missing user_id")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id is required",
		})
		return
	}

	span.SetAttributes(attribute.String("user_id", userID))

	// Clear cart in Redis
	if err := h.redisClient.ClearCart(ctx, userID); err != nil {
		span.SetStatus(codes.Error, "Failed to clear cart")
		span.RecordError(err)
		h.logger.Error("Failed to clear cart",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to clear cart",
		})
		return
	}

	span.SetStatus(codes.Ok, "Cart cleared successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "Cart cleared successfully",
		"user_id": userID,
	})
}
