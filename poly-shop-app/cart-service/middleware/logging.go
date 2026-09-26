package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// ZapMiddleware returns a Gin middleware that logs HTTP requests using Zap
// Logs include trace_id for correlation with distributed traces
// This middleware should be added after the tracing middleware to capture trace IDs
func ZapMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Calculate request duration
		duration := time.Since(start)
		status := c.Writer.Status()

		// Extract trace ID from span context for log correlation
		// This allows correlating logs with traces in observability systems
		var traceID string
		spanContext := trace.SpanContextFromContext(c.Request.Context())
		if spanContext.IsValid() {
			traceID = spanContext.TraceID().String()
		}

		// Determine log level based on status code
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		// Add trace_id if available for correlation
		if traceID != "" {
			fields = append(fields, zap.String("trace_id", traceID))
		}

		// Add error if present
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("error", c.Errors.String()))
		}

		// Log based on status code
		if status >= 500 {
			logger.Error("HTTP request failed", fields...)
		} else if status >= 400 {
			logger.Warn("HTTP request client error", fields...)
		} else {
			logger.Info("HTTP request completed", fields...)
		}
	}
}
