package handlers

import (
	"net/http"

	"product-service/database"

	"github.com/gin-gonic/gin"
)

// ProductHandler handles product-related HTTP requests
type ProductHandler struct {
	repository database.ProductRepository
}

// NewProductHandler creates a new product handler with a repository
func NewProductHandler(repository database.ProductRepository) *ProductHandler {
	return &ProductHandler{
		repository: repository,
	}
}

// GetProducts handles the GET /products endpoint
// It retrieves products from PostgreSQL with optional category filtering
func (h *ProductHandler) GetProducts(c *gin.Context) {
	// Get the current context from Gin (which already has trace context from middleware)
	ctx := c.Request.Context()

	// Check for optional category query parameter
	category := c.Query("category")

	var products []database.Product
	var err error

	if category != "" {
		// Filter by category
		products, err = h.repository.GetProductsByCategory(ctx, category)
	} else {
		// Get all products
		products, err = h.repository.GetAllProducts(ctx)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve products",
			"message": err.Error(),
		})
		return
	}

	// Return the products as JSON
	c.JSON(http.StatusOK, products)
}
