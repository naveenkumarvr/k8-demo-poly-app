package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProducts(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	t.Run("should return 200 OK", func(t *testing.T) {
		// Create test router and recorder
		router := gin.New()
		router.GET("/products", GetProducts)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/products", nil)

		// Execute request
		router.ServeHTTP(w, req)

		// Assert status code
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return valid JSON array", func(t *testing.T) {
		router := gin.New()
		router.GET("/products", GetProducts)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/products", nil)

		router.ServeHTTP(w, req)

		// Assert valid JSON array
		var products []Product
		err := json.Unmarshal(w.Body.Bytes(), &products)
		require.NoError(t, err, "Response should be valid JSON")
	})

	t.Run("should return at least 10 products", func(t *testing.T) {
		router := gin.New()
		router.GET("/products", GetProducts)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/products", nil)

		router.ServeHTTP(w, req)

		var products []Product
		json.Unmarshal(w.Body.Bytes(), &products)

		// Assert at least 10 products
		assert.GreaterOrEqual(t, len(products), 10, "Should return at least 10 products")
	})

	t.Run("each product should have required fields", func(t *testing.T) {
		router := gin.New()
		router.GET("/products", GetProducts)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/products", nil)

		router.ServeHTTP(w, req)

		var products []Product
		json.Unmarshal(w.Body.Bytes(), &products)

		// Check first product has all required fields
		require.NotEmpty(t, products, "Products array should not be empty")
		
		product := products[0]
		assert.NotZero(t, product.ID, "Product ID should not be zero")
		assert.NotEmpty(t, product.Name, "Product name should not be empty")
		assert.NotEmpty(t, product.Description, "Product description should not be empty")
		assert.Greater(t, product.Price, 0.0, "Product price should be positive")
		assert.NotEmpty(t, product.ImageURL, "Product image URL should not be empty")
	})

	t.Run("all products should have unique IDs", func(t *testing.T) {
		router := gin.New()
		router.GET("/products", GetProducts)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/products", nil)

		router.ServeHTTP(w, req)

		var products []Product
		json.Unmarshal(w.Body.Bytes(), &products)

		// Create map to track IDs
		idMap := make(map[int]bool)
		for _, product := range products {
			assert.False(t, idMap[product.ID], "Product ID %d should be unique", product.ID)
			idMap[product.ID] = true
		}
	})

	t.Run("should have valid price range", func(t *testing.T) {
		router := gin.New()
		router.GET("/products", GetProducts)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/products", nil)

		router.ServeHTTP(w, req)

		var products []Product
		json.Unmarshal(w.Body.Bytes(), &products)

		// All prices should be positive and reasonable
		for _, product := range products {
			assert.Greater(t, product.Price, 0.0, "Price should be positive")
			assert.Less(t, product.Price, 10000.0, "Price should be reasonable")
		}
	})
}

func TestMockProducts(t *testing.T) {
	t.Run("should return consistent data", func(t *testing.T) {
		// Call mockProducts multiple times
		products1 := mockProducts()
		products2 := mockProducts()

		// Should return the same data each time
		assert.Equal(t, len(products1), len(products2))
		assert.Equal(t, products1[0].ID, products2[0].ID)
		assert.Equal(t, products1[0].Name, products2[0].Name)
	})

	t.Run("should have at least 10 products", func(t *testing.T) {
		products := mockProducts()
		assert.GreaterOrEqual(t, len(products), 10)
	})
}

// Benchmark test to measure performance
func BenchmarkGetProducts(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/products", GetProducts)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/products", nil)
		router.ServeHTTP(w, req)
	}
}
