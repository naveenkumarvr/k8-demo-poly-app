package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// StressHandler holds dependencies for stress test handlers
type StressHandler struct {
	logger *zap.Logger
}

// StressResponse represents the response from the stress test endpoint
type StressResponse struct {
	CPUIterations    int    `json:"cpu_iterations"`
	MemoryMB         int    `json:"memory_mb"`
	PrimesCalculated int    `json:"primes_calculated"`
	ComputationTime  string `json:"computation_time"`
	Message          string `json:"message"`
}

// NewStressHandler creates a new stress handler
func NewStressHandler(logger *zap.Logger) *StressHandler {
	return &StressHandler{
		logger: logger,
	}
}

// StressTest handles POST /stress
// Artificial CPU/Memory load generator for performance profiling and HPA testing
// Query parameters:
// - cpu_iterations: Number of iterations for prime calculation (default: 1000)
// - memory_mb: Amount of memory to allocate in MB (default: 100)
func (h *StressHandler) StressTest(c *gin.Context) {
	ctx := c.Request.Context()
	tracer := otel.Tracer("cart-service")
	ctx, span := tracer.Start(ctx, "handler.StressTest")
	defer span.End()

	// Parse query parameters
	cpuIterations, _ := strconv.Atoi(c.DefaultQuery("cpu_iterations", "1000"))
	memoryMB, _ := strconv.Atoi(c.DefaultQuery("memory_mb", "100"))

	// Validate parameters
	if cpuIterations < 0 || cpuIterations > 10000 {
		span.SetStatus(codes.Error, "Invalid cpu_iterations")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid cpu_iterations",
			"message": "cpu_iterations must be between 0 and 10000",
		})
		return
	}

	if memoryMB < 0 || memoryMB > 1000 {
		span.SetStatus(codes.Error, "Invalid memory_mb")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid memory_mb",
			"message": "memory_mb must be between 0 and 1000",
		})
		return
	}

	span.SetAttributes(
		attribute.Int("cpu_iterations", cpuIterations),
		attribute.Int("memory_mb", memoryMB),
	)

	h.logger.Info("Starting stress test",
		zap.Int("cpu_iterations", cpuIterations),
		zap.Int("memory_mb", memoryMB),
	)

	startTime := time.Now()

	// CPU Stress: Calculate prime numbers
	primesFound := 0
	if cpuIterations > 0 {
		primesFound = calculatePrimes(cpuIterations)
	}

	// Memory Stress: Allocate and populate byte slices
	if memoryMB > 0 {
		allocateMemory(memoryMB)
	}

	duration := time.Since(startTime)

	span.SetAttributes(
		attribute.Int("primes_calculated", primesFound),
		attribute.Int64("duration_ms", duration.Milliseconds()),
	)
	span.SetStatus(codes.Ok, "Stress test completed")

	h.logger.Info("Stress test completed",
		zap.Int("cpu_iterations", cpuIterations),
		zap.Int("memory_mb", memoryMB),
		zap.Int("primes_calculated", primesFound),
		zap.Duration("duration", duration),
	)

	response := StressResponse{
		CPUIterations:    cpuIterations,
		MemoryMB:         memoryMB,
		PrimesCalculated: primesFound,
		ComputationTime:  duration.String(),
		Message:          "Stress test completed successfully",
	}

	c.JSON(http.StatusOK, response)
}

// calculatePrimes performs CPU-intensive prime number calculation
// Uses trial division algorithm to find all primes up to maxNum over multiple iterations
func calculatePrimes(iterations int) int {
	const maxNum = 10000
	totalPrimes := 0

	for i := 0; i < iterations; i++ {
		primeCount := 0
		for num := 2; num <= maxNum; num++ {
			if isPrime(num) {
				primeCount++
			}
		}
		totalPrimes = primeCount
	}

	return totalPrimes
}

// isPrime checks if a number is prime using trial division
func isPrime(n int) bool {
	if n <= 1 {
		return false
	}
	if n <= 3 {
		return true
	}
	if n%2 == 0 || n%3 == 0 {
		return false
	}
	for i := 5; i*i <= n; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
	}
	return true
}

// allocateMemory allocates large byte slices to stress memory
// Also performs JSON marshalling to add CPU overhead
func allocateMemory(sizeMB int) {
	// Allocate byte slices
	// Each chunk is 1MB
	chunks := make([][]byte, sizeMB)
	for i := 0; i < sizeMB; i++ {
		// Allocate 1MB chunk
		chunk := make([]byte, 1024*1024)

		// Fill with pseudo-random data to prevent optimization
		for j := range chunk {
			chunk[j] = byte((i + j) % 256)
		}

		chunks[i] = chunk
	}

	// Perform heavy JSON marshalling to add CPU load
	// This prevents the compiler from optimizing away the memory allocation
	largeObject := make(map[string]interface{})
	largeObject["chunks_count"] = sizeMB
	largeObject["timestamp"] = time.Now().Unix()
	largeObject["data"] = make([]map[string]int, 100)

	for i := 0; i < 100; i++ {
		largeObject["data"].([]map[string]int)[i] = map[string]int{
			"index": i,
			"value": i * 1000,
		}
	}

	// Marshal to JSON (CPU-intensive operation)
	_, _ = json.Marshal(largeObject)

	// Keep chunks alive until this function returns
	// This ensures memory stays allocated for the duration of the test
	_ = fmt.Sprintf("%d chunks allocated", len(chunks))
}
