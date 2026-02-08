package telemetry

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// TracerConfig holds the configuration for OpenTelemetry tracing
type TracerConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
}

// tracerProvider holds the global tracer provider for cleanup
var tracerProvider *sdktrace.TracerProvider

// InitTracer initializes the OpenTelemetry tracer with OTLP/gRPC exporter
// It sets up W3C Trace Context propagation and batch span processing
// Returns a shutdown function that should be called on application exit
func InitTracer(config TracerConfig) (func(context.Context) error, error) {
	ctx := context.Background()

	// Create resource with service information
	// These attributes identify the service in the observability backend
	res, err := resource.New(ctx,
		resource.WithAttributes(
			// Service identification attributes
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironment(config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP/gRPC trace exporter
	// This sends traces to the OTel Collector via gRPC
	// WithInsecure() is used for local development; use TLS in production
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(config.OTLPEndpoint),
		otlptracegrpc.WithInsecure(), // Remove in production, use TLS
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	// Create tracer provider with batch span processor
	// BatchSpanProcessor batches spans before export for better performance
	// This reduces the number of network calls to the collector
	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithMaxExportBatchSize(512),
			sdktrace.WithBatchTimeout(5*time.Second),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // Sample all traces for demo
	)

	// Set the global tracer provider
	// This allows accessing the tracer from anywhere in the application
	otel.SetTracerProvider(tracerProvider)

	// Set the global propagator to W3C Trace Context
	// This ensures trace context is correctly extracted from and injected into HTTP headers
	// W3C Trace Context format: traceparent: 00-<trace-id>-<span-id>-<trace-flags>
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, // W3C Trace Context
		propagation.Baggage{},      // W3C Baggage
	))

	log.Printf("OpenTelemetry tracer initialized: service=%s, version=%s, environment=%s, endpoint=%s",
		config.ServiceName, config.ServiceVersion, config.Environment, config.OTLPEndpoint)

	// Return shutdown function
	// This should be called on application shutdown to flush remaining spans
	return tracerProvider.Shutdown, nil
}

// Shutdown gracefully shuts down the tracer provider
// This should be called before application exit to ensure all spans are exported
func Shutdown(ctx context.Context) error {
	if tracerProvider != nil {
		return tracerProvider.Shutdown(ctx)
	}
	return nil
}
