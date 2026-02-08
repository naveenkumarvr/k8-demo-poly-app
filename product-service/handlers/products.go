package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// Product represents a single product in the inventory
type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"image_url"`
}

// mockProducts returns a static list of mock products for the demo
// In a real application, this would query from a database
func mockProducts() []Product {
	return []Product{
		{
			ID:          1,
			Name:        "Wireless Headphones",
			Description: "Premium noise-cancelling wireless headphones with 30-hour battery life",
			Price:       199.99,
			ImageURL:    "https://images.example.com/headphones.jpg",
		},
		{
			ID:          2,
			Name:        "Smart Watch",
			Description: "Fitness tracker with heart rate monitor and GPS",
			Price:       299.99,
			ImageURL:    "https://images.example.com/smartwatch.jpg",
		},
		{
			ID:          3,
			Name:        "Laptop Stand",
			Description: "Ergonomic aluminum laptop stand with adjustable height",
			Price:       49.99,
			ImageURL:    "https://images.example.com/laptop-stand.jpg",
		},
		{
			ID:          4,
			Name:        "Mechanical Keyboard",
			Description: "RGB backlit mechanical keyboard with Cherry MX switches",
			Price:       129.99,
			ImageURL:    "https://images.example.com/keyboard.jpg",
		},
		{
			ID:          5,
			Name:        "Webcam 4K",
			Description: "Ultra HD webcam with auto-focus and built-in microphone",
			Price:       89.99,
			ImageURL:    "https://images.example.com/webcam.jpg",
		},
		{
			ID:          6,
			Name:        "USB-C Hub",
			Description: "7-in-1 USB-C hub with HDMI, USB 3.0, and SD card reader",
			Price:       39.99,
			ImageURL:    "https://images.example.com/usb-hub.jpg",
		},
		{
			ID:          7,
			Name:        "Portable SSD 1TB",
			Description: "High-speed portable SSD with 1TB storage capacity",
			Price:       149.99,
			ImageURL:    "https://images.example.com/ssd.jpg",
		},
		{
			ID:          8,
			Name:        "Wireless Mouse",
			Description: "Ergonomic wireless mouse with precision tracking",
			Price:       34.99,
			ImageURL:    "https://images.example.com/mouse.jpg",
		},
		{
			ID:          9,
			Name:        "Monitor 27-inch",
			Description: "27-inch 4K UHD monitor with HDR support",
			Price:       399.99,
			ImageURL:    "https://images.example.com/monitor.jpg",
		},
		{
			ID:          10,
			Name:        "Desk Lamp LED",
			Description: "Adjustable LED desk lamp with touch control and USB charging",
			Price:       29.99,
			ImageURL:    "https://images.example.com/desk-lamp.jpg",
		},
		{
			ID:          11,
			Name:        "Gaming Chair",
			Description: "Ergonomic gaming chair with lumbar support and adjustable armrests",
			Price:       249.99,
			ImageURL:    "https://images.example.com/gaming-chair.jpg",
		},
		{
			ID:          12,
			Name:        "Bluetooth Speaker",
			Description: "Portable Bluetooth speaker with 360-degree sound and waterproof design",
			Price:       79.99,
			ImageURL:    "https://images.example.com/speaker.jpg",
		},
	}
}

// GetProducts handles the GET /products endpoint
// It simulates a database fetch with a custom OpenTelemetry span
func GetProducts(c *gin.Context) {
	// Get the current context from Gin (which already has trace context from middleware)
	ctx := c.Request.Context()

	// Create a custom span to simulate database fetching
	// This demonstrates manual span creation for specific operations
	tracer := otel.Tracer("product-service")
	ctx, span := tracer.Start(ctx, "fetch_products_from_database")
	defer span.End()

	// Simulate database latency (50-100ms) to make traces more realistic
	// In a real application, this would be actual database query time
	sleepDuration := 75 * time.Millisecond
	time.Sleep(sleepDuration)

	// Fetch the mock products
	products := mockProducts()

	// Add span attributes for better observability
	// These attributes help in filtering and analyzing traces
	span.SetAttributes(
		attribute.Int("product.count", len(products)),
		attribute.String("database.operation", "SELECT"),
		attribute.String("database.table", "products"),
		attribute.Int64("fetch.duration_ms", sleepDuration.Milliseconds()),
	)

	// Mark the span as successful
	span.SetStatus(codes.Ok, "Products fetched successfully")

	// Return the products as JSON
	c.JSON(http.StatusOK, products)
}
