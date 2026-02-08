package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// StressResponse represents the response from the stress test endpoint
type StressResponse struct {
	Input           int    `json:"input"`
	Result          uint64 `json:"result"`
	ComputationTime string `json:"computation_time"`
	Message         string `json:"message"`
}

// fibonacci calculates the nth Fibonacci number recursively
// This is intentionally inefficient for CPU stress testing
// Time complexity: O(2^n) - exponential growth
func fibonacci(n int) uint64 {
	if n <= 1 {
		return uint64(n)
	}
	// Recursive calls create exponential work
	// This will heavily utilize CPU for larger values of n
	return fibonacci(n-1) + fibonacci(n-2)
}

// StressTest handles the GET /stress endpoint
// This endpoint is designed for Horizontal Pod Autoscaler (HPA) testing
// by performing CPU-intensive recursive calculations
func StressTest(c *gin.Context) {
	// Get the current context from Gin
	ctx := c.Request.Context()

	// Create a custom span to track the stress computation
	tracer := otel.Tracer("product-service")
	ctx, span := tracer.Start(ctx, "stress_test_computation")
	defer span.End()

	// Parse the 'n' query parameter, default to 42 if not provided
	// Example: /stress?n=40
	nStr := c.DefaultQuery("n", "42")
	n, err := strconv.Atoi(nStr)
	if err != nil || n < 0 {
		span.SetStatus(codes.Error, "Invalid input parameter")
		span.SetAttributes(attribute.String("error", "invalid_parameter"))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid parameter 'n'",
			"message": "Parameter 'n' must be a non-negative integer",
		})
		return
	}

	// Limit the input to prevent excessive computation
	// Fibonacci(50) takes several minutes on a single CPU core
	if n > 50 {
		span.SetStatus(codes.Error, "Input too large")
		span.SetAttributes(attribute.Int("input.value", n))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Input too large",
			"message": "Maximum allowed value is 50",
		})
		return
	}

	// Add span attribute for the input value
	span.SetAttributes(attribute.Int("fibonacci.input", n))

	// Record the start time
	startTime := time.Now()

	// Perform the CPU-intensive Fibonacci calculation
	// For n=42, this typically takes 2-5 seconds on a modern CPU
	// For n=45, this can take 10-30 seconds
	// This creates measurable CPU load for HPA testing
	result := fibonacci(n)

	// Calculate the computation time
	duration := time.Since(startTime)

	// Add span attributes for observability
	span.SetAttributes(
		attribute.Int64("computation.duration_ms", duration.Milliseconds()),
		attribute.String("fibonacci.result", strconv.FormatUint(result, 10)),
	)

	// Mark the span as successful
	span.SetStatus(codes.Ok, "Stress computation completed")

	// Return the result
	response := StressResponse{
		Input:           n,
		Result:          result,
		ComputationTime: duration.String(),
		Message:         "CPU stress test completed successfully",
	}

	c.JSON(http.StatusOK, response)
}
