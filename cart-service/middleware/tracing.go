package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// TracingMiddleware returns a Gin middleware that adds OpenTelemetry tracing
// It automatically:
// - Extracts W3C Trace Context from incoming HTTP headers
// - Creates a span for each HTTP request
// - Adds HTTP-specific attributes (method, route, status code)
// - Injects trace context into outgoing requests
func TracingMiddleware(serviceName string) gin.HandlerFunc {
	// Use the official OTel Gin instrumentation
	// This handles all the complexity of context propagation
	return otelgin.Middleware(serviceName)
}
