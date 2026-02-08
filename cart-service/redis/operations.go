package redis

import (
	"context"
	"fmt"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// CartItem represents an item in a user's cart
type CartItem struct {
	ProductID string
	Quantity  int
}

// AddItem adds an item to a user's cart or increments the quantity if it already exists
// Redis data structure: Hash key = "cart:{userID}", field = productID, value = quantity
// Uses HINCRBY to atomically increment the quantity
// Creates a child span for observability
func (c *Client) AddItem(ctx context.Context, userID, productID string, quantity int) error {
	// Create a child span for this operation
	tracer := otel.Tracer("cart-service")
	ctx, span := tracer.Start(ctx, "redis.AddItem")
	defer span.End()

	// Add span attributes for observability
	span.SetAttributes(
		attribute.String("user_id", userID),
		attribute.String("product_id", productID),
		attribute.Int("quantity", quantity),
	)

	if quantity <= 0 {
		span.SetStatus(codes.Error, "Invalid quantity")
		return fmt.Errorf("quantity must be positive, got %d", quantity)
	}

	// Redis key for user's cart
	key := fmt.Sprintf("cart:%s", userID)

	// Use HINCRBY to atomically increment the quantity
	// This handles both adding new items and updating existing ones
	err := c.rdb.HIncrBy(ctx, key, productID, int64(quantity)).Err()
	if err != nil {
		span.SetStatus(codes.Error, "Redis HINCRBY failed")
		span.RecordError(err)
		c.logger.Error("Failed to add item to cart",
			zap.String("user_id", userID),
			zap.String("product_id", productID),
			zap.Int("quantity", quantity),
			zap.Error(err),
		)
		return fmt.Errorf("failed to add item to cart: %w", err)
	}

	span.SetStatus(codes.Ok, "Item added successfully")
	c.logger.Info("Item added to cart",
		zap.String("user_id", userID),
		zap.String("product_id", productID),
		zap.Int("quantity", quantity),
	)

	return nil
}

// GetCart retrieves all items in a user's cart
// Uses HGETALL to fetch all product_id:quantity pairs
// Returns an empty slice if cart doesn't exist
func (c *Client) GetCart(ctx context.Context, userID string) ([]CartItem, error) {
	// Create a child span for this operation
	tracer := otel.Tracer("cart-service")
	ctx, span := tracer.Start(ctx, "redis.GetCart")
	defer span.End()

	span.SetAttributes(attribute.String("user_id", userID))

	key := fmt.Sprintf("cart:%s", userID)

	// Use HGETALL to fetch all fields and values
	// Returns map[string]string where key=productID, value=quantity
	result, err := c.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		span.SetStatus(codes.Error, "Redis HGETALL failed")
		span.RecordError(err)
		c.logger.Error("Failed to get cart",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	// Convert map to slice of CartItem
	items := make([]CartItem, 0, len(result))
	for productID, quantityStr := range result {
		quantity, err := strconv.Atoi(quantityStr)
		if err != nil {
			// Skip invalid entries
			c.logger.Warn("Invalid quantity in cart, skipping",
				zap.String("user_id", userID),
				zap.String("product_id", productID),
				zap.String("quantity_str", quantityStr),
				zap.Error(err),
			)
			continue
		}

		items = append(items, CartItem{
			ProductID: productID,
			Quantity:  quantity,
		})
	}

	span.SetAttributes(attribute.Int("item_count", len(items)))
	span.SetStatus(codes.Ok, "Cart retrieved successfully")

	return items, nil
}

// ClearCart removes all items from a user's cart
// Uses DEL to delete the entire hash
func (c *Client) ClearCart(ctx context.Context, userID string) error {
	// Create a child span for this operation
	tracer := otel.Tracer("cart-service")
	ctx, span := tracer.Start(ctx, "redis.ClearCart")
	defer span.End()

	span.SetAttributes(attribute.String("user_id", userID))

	key := fmt.Sprintf("cart:%s", userID)

	// Use DEL to remove the entire hash
	err := c.rdb.Del(ctx, key).Err()
	if err != nil {
		span.SetStatus(codes.Error, "Redis DEL failed")
		span.RecordError(err)
		c.logger.Error("Failed to clear cart",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to clear cart: %w", err)
	}

	span.SetStatus(codes.Ok, "Cart cleared successfully")
	c.logger.Info("Cart cleared", zap.String("user_id", userID))

	return nil
}

// ItemCount returns the number of distinct items (not total quantity) in a cart
// Uses HLEN to count hash fields
func (c *Client) ItemCount(ctx context.Context, userID string) (int64, error) {
	// Create a child span for this operation
	tracer := otel.Tracer("cart-service")
	ctx, span := tracer.Start(ctx, "redis.ItemCount")
	defer span.End()

	span.SetAttributes(attribute.String("user_id", userID))

	key := fmt.Sprintf("cart:%s", userID)

	count, err := c.rdb.HLen(ctx, key).Result()
	if err != nil {
		span.SetStatus(codes.Error, "Redis HLEN failed")
		span.RecordError(err)
		return 0, fmt.Errorf("failed to get item count: %w", err)
	}

	span.SetAttributes(attribute.Int64("item_count", count))
	span.SetStatus(codes.Ok, "Item count retrieved")

	return count, nil
}
