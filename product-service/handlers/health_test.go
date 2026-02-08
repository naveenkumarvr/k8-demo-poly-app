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

func TestHealthz(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return 200 OK", func(t *testing.T) {
		router := gin.New()
		router.GET("/healthz", Healthz)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/healthz", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return valid JSON", func(t *testing.T) {
		router := gin.New()
		router.GET("/healthz", Healthz)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/healthz", nil)

		router.ServeHTTP(w, req)

		var response HealthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "ok", response.Status)
		assert.Equal(t, "product-service", response.Service)
	})

	t.Run("should have correct content type", func(t *testing.T) {
		router := gin.New()
		router.GET("/healthz", Healthz)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/healthz", nil)

		router.ServeHTTP(w, req)

		assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	})
}

func TestReady(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return 200 OK", func(t *testing.T) {
		router := gin.New()
		router.GET("/ready", Ready)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/ready", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return valid JSON with ready status", func(t *testing.T) {
		router := gin.New()
		router.GET("/ready", Ready)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/ready", nil)

		router.ServeHTTP(w, req)

		var response HealthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "ready", response.Status)
		assert.Equal(t, "product-service", response.Service)
	})

	t.Run("should be compatible with Kubernetes readiness probe", func(t *testing.T) {
		router := gin.New()
		router.GET("/ready", Ready)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/ready", nil)

		router.ServeHTTP(w, req)

		// Kubernetes expects HTTP 200 for healthy status
		assert.Equal(t, http.StatusOK, w.Code)
		// Should return JSON
		assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	})
}

func TestLive(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return 200 OK", func(t *testing.T) {
		router := gin.New()
		router.GET("/live", Live)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/live", nil)

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return valid JSON with alive status", func(t *testing.T) {
		router := gin.New()
		router.GET("/live", Live)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/live", nil)

		router.ServeHTTP(w, req)

		var response HealthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "alive", response.Status)
		assert.Equal(t, "product-service", response.Service)
	})

	t.Run("should be compatible with Kubernetes liveness probe", func(t *testing.T) {
		router := gin.New()
		router.GET("/live", Live)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/live", nil)

		router.ServeHTTP(w, req)

		// Kubernetes expects HTTP 200 for healthy status
		assert.Equal(t, http.StatusOK, w.Code)
		// Should return JSON
		assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	})
}

// Test all health endpoints together
func TestAllHealthEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.GET("/healthz", Healthz)
	router.GET("/ready", Ready)
	router.GET("/live", Live)

	endpoints := []struct {
		path           string
		expectedStatus string
	}{
		{"/healthz", "ok"},
		{"/ready", "ready"},
		{"/live", "alive"},
	}

	for _, endpoint := range endpoints {
		t.Run("Testing "+endpoint.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", endpoint.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			
			var response HealthResponse
			json.Unmarshal(w.Body.Bytes(), &response)
			assert.Equal(t, endpoint.expectedStatus, response.Status)
			assert.Equal(t, "product-service", response.Service)
		})
	}
}
