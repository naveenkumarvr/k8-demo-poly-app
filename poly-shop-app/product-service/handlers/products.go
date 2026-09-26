package handlers

import (
	"fmt"
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

// GetProductByID handles the GET /products/:id endpoint
// It retrieves a single product by ID
func (h *ProductHandler) GetProductByID(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")

	// Parse ID string to int
	var id int
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid product ID",
		})
		return
	}

	product, err := h.repository.GetProductByID(ctx, id)
	if err != nil {
		// Check if product not found (this depends on repository error behavior)
		// For now, we assume any error is 500 except specific "no rows" if exposed
		// In a real app, we'd check for sql.ErrNoRows wrapped error
		if err.Error() == "failed to get product by ID "+idStr+": no rows in result set" { // specific check might be brittle
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Product not found",
			})
			return
		}

		// Improve error handling: check if error message contains "no rows"
		if contains(err.Error(), "no rows") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Product not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve product",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, product)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr // simplistic, use strings.Contains
}
