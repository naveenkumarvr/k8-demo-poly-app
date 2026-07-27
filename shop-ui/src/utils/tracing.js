import { WebTracerProvider, StackContextManager } from '@opentelemetry/sdk-trace-web';
import { SimpleSpanProcessor, ConsoleSpanExporter, BatchSpanProcessor } from '@opentelemetry/sdk-trace-base';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
import { registerInstrumentations } from '@opentelemetry/instrumentation';
import { FetchInstrumentation } from '@opentelemetry/instrumentation-fetch';

let tracer;

try {
    const provider = new WebTracerProvider();

    const defaultOtelEndpoint = 'http://localhost:4318/v1/traces';
    const otelEndpoint = import.meta.env.VITE_OTEL_EXPORTER_OTLP_ENDPOINT || defaultOtelEndpoint;
    const exporter = new OTLPTraceExporter({ url: otelEndpoint });

    provider.addSpanProcessor(new BatchSpanProcessor(exporter));

    // For development, also log to console
    if (import.meta.env.DEV) {
        provider.addSpanProcessor(new SimpleSpanProcessor(new ConsoleSpanExporter()));
    }

    provider.register({
        contextManager: new StackContextManager(),
    });

    registerInstrumentations({
        instrumentations: [
            new FetchInstrumentation({
                indent: 2,
                propagateTraceHeaderCorsUrls: /.*/,
                clearTimingResources: true,
            }),
        ],
    });

    tracer = provider.getTracer('shop-ui');
} catch (error) {
    console.warn('OpenTelemetry tracing initialization failed, continuing without tracing:', error);
    tracer = {
        startSpan: () => ({ end: () => {}, setStatus: () => {}, recordException: () => {} }),
        startActiveSpan: (name, fn) => fn({ end: () => {}, setStatus: () => {}, recordException: () => {} }),
    };
}

export { tracer };
