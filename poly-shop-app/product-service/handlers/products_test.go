package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"product-service/database"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProductRepository is a test double for database.ProductRepository
type mockProductRepository struct {
	products []database.Product
	err      error
}

func (m *mockProductRepository) GetAllProducts(ctx context.Context) ([]database.Product, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.products, nil
}

func (m *mockProductRepository) GetProductByID(ctx context.Context, id int) (*database.Product, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, p := range m.products {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockProductRepository) GetProductsByCategory(ctx context.Context, category string) ([]database.Product, error) {
	if m.err != nil {
		return nil, m.err
	}
	var filtered []database.Product
	for _, p := range m.products {
		if p.Category == category {
			filtered = append(filtered, p)
		}
	}
	return filtered, nil
}

func (m *mockProductRepository) CreateProduct(ctx context.Context, product *database.Product) error {
	return nil
}

func testProducts() []database.Product {
	now := time.Now()
	return []database.Product{
		{
			ID:          1,
			Name:        "Test Product",
			Description: "A great product",
			Price:       10.0,
			Stock:       100,
			Category:    "electronics",
			ImageURL:    "/test.jpg",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          2,
			Name:        "Another Product",
			Description: "Another great product",
			Price:       20.0,
			Stock:       50,
			Category:    "electronics",
			ImageURL:    "/another.jpg",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
}

func TestGetProducts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return 200 OK", func(t *testing.T) {
		handler := NewProductHandler(&mockProductRepository{products: testProducts()})
		router := gin.New()
		router.GET("/products", handler.GetProducts)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/products", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return valid JSON array", func(t *testing.T) {
		handler := NewProductHandler(&mockProductRepository{products: testProducts()})
		router := gin.New()
		router.GET("/products", handler.GetProducts)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/products", nil)
		router.ServeHTTP(w, req)

		var products []database.Product
		err := json.Unmarshal(w.Body.Bytes(), &products)
		require.NoError(t, err, "Response should be valid JSON")
		assert.Len(t, products, 2)
	})

	t.Run("should return products from repository", func(t *testing.T) {
		handler := NewProductHandler(&mockProductRepository{products: testProducts()})
		router := gin.New()
		router.GET("/products", handler.GetProducts)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/products", nil)
		router.ServeHTTP(w, req)

		var products []database.Product
		json.Unmarshal(w.Body.Bytes(), &products)

		require.Len(t, products, 2)
		assert.Equal(t, "Test Product", products[0].Name)
		assert.Equal(t, 10.0, products[0].Price)
		assert.Equal(t, "/test.jpg", products[0].ImageURL)
	})

	t.Run("should filter by category", func(t *testing.T) {
		handler := NewProductHandler(&mockProductRepository{products: testProducts()})
		router := gin.New()
		router.GET("/products", handler.GetProducts)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/products?category=electronics", nil)
		router.ServeHTTP(w, req)

		var products []database.Product
		json.Unmarshal(w.Body.Bytes(), &products)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Len(t, products, 2)
	})

	t.Run("should return 500 when repository fails", func(t *testing.T) {
		handler := NewProductHandler(&mockProductRepository{err: errors.New("database error")})
		router := gin.New()
		router.GET("/products", handler.GetProducts)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/products", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// Benchmark test to measure performance
func BenchmarkGetProducts(b *testing.B) {
	gin.SetMode(gin.TestMode)
	handler := NewProductHandler(&mockProductRepository{products: testProducts()})
	router := gin.New()
	router.GET("/products", handler.GetProducts)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/products", nil)
		router.ServeHTTP(w, req)
	}
}
