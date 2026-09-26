package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestStressTest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	handler := NewStressHandler(logger)

	t.Run("should handle default parameters", func(t *testing.T) {
		router := gin.New()
		router.POST("/stress", handler.StressTest)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/stress", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response StressResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, 1000, response.CPUIterations)
		assert.Equal(t, 100, response.MemoryMB)
		assert.Greater(t, response.PrimesCalculated, 0)
		assert.NotEmpty(t, response.ComputationTime)
	})

	t.Run("should handle custom parameters", func(t *testing.T) {
		router := gin.New()
		router.POST("/stress", handler.StressTest)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/stress?cpu_iterations=100&memory_mb=10", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response StressResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, 100, response.CPUIterations)
		assert.Equal(t, 10, response.MemoryMB)
	})

	t.Run("should reject invalid cpu_iterations", func(t *testing.T) {
		router := gin.New()
		router.POST("/stress", handler.StressTest)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/stress?cpu_iterations=20000", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should reject invalid memory_mb", func(t *testing.T) {
		router := gin.New()
		router.POST("/stress", handler.StressTest)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/stress?memory_mb=2000", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should handle zero values", func(t *testing.T) {
		router := gin.New()
		router.POST("/stress", handler.StressTest)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/stress?cpu_iterations=0&memory_mb=0", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response StressResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.Equal(t, 0, response.CPUIterations)
		assert.Equal(t, 0, response.MemoryMB)
		assert.Equal(t, 0, response.PrimesCalculated)
	})
}

func TestIsPrime(t *testing.T) {
	tests := []struct {
		n        int
		expected bool
	}{
		{1, false},
		{2, true},
		{3, true},
		{4, false},
		{5, true},
		{17, true},
		{18, false},
		{97, true},
		{100, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := isPrime(tt.n)
			assert.Equal(t, tt.expected, result, "isPrime(%d) should be %v", tt.n, tt.expected)
		})
	}
}
