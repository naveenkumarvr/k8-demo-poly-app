package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFibonacci(t *testing.T) {
	tests := []struct {
		input    int
		expected uint64
	}{
		{0, 0},
		{1, 1},
		{2, 1},
		{3, 2},
		{4, 3},
		{5, 5},
		{6, 8},
		{10, 55},
		{20, 6765},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("fibonacci(%d)", tt.input), func(t *testing.T) {
			result := fibonacci(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStressTest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return 200 OK with default parameter", func(t *testing.T) {
		router := gin.New()
		router.GET("/stress", StressTest)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/stress", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return valid JSON response", func(t *testing.T) {
		router := gin.New()
		router.GET("/stress", StressTest)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/stress?n=10", nil)

		router.ServeHTTP(w, req)

		var response StressResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "Response should be valid JSON")
		
		assert.Equal(t, 10, response.Input)
		assert.Equal(t, uint64(55), response.Result)
		assert.NotEmpty(t, response.ComputationTime)
		assert.NotEmpty(t, response.Message)
	})

	t.Run("should handle query parameter n=20", func(t *testing.T) {
		router := gin.New()
		router.GET("/stress", StressTest)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/stress?n=20", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response StressResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		
		assert.Equal(t, 20, response.Input)
		assert.Equal(t, uint64(6765), response.Result)
	})

	t.Run("should reject negative input", func(t *testing.T) {
		router := gin.New()
		router.GET("/stress", StressTest)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/stress?n=-5", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var errorResponse map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.Contains(t, errorResponse, "error")
	})

	t.Run("should reject invalid input", func(t *testing.T) {
		router := gin.New()
		router.GET("/stress", StressTest)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/stress?n=invalid", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should reject input greater than 50", func(t *testing.T) {
		router := gin.New()
		router.GET("/stress", StressTest)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/stress?n=51", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var errorResponse map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.Contains(t, errorResponse["error"], "too large")
	})

	t.Run("should accept maximum allowed value of 50", func(t *testing.T) {
		router := gin.New()
		router.GET("/stress", StressTest)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/stress?n=50", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should include computation time in response", func(t *testing.T) {
		router := gin.New()
		router.GET("/stress", StressTest)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/stress?n=30", nil)

		router.ServeHTTP(w, req)

		var response StressResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		
		// Computation time should be present and not empty
		assert.NotEmpty(t, response.ComputationTime)
		// Should contain time units (ms, µs, or s)
		assert.Contains(t, response.ComputationTime, "s")
	})
}

// Benchmark the Fibonacci function
func BenchmarkFibonacci(b *testing.B) {
	inputs := []int{10, 20, 25, 30}
	
	for _, n := range inputs {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				fibonacci(n)
			}
		})
	}
}

// Benchmark the stress endpoint
func BenchmarkStressTest(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/stress", StressTest)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/stress?n=20", nil)
		router.ServeHTTP(w, req)
	}
}
