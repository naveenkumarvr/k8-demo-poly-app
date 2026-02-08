import { WebTracerProvider } from '@opentelemetry/sdk-trace-web';
import { SimpleSpanProcessor, ConsoleSpanExporter, BatchSpanProcessor } from '@opentelemetry/sdk-trace-base';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
import { registerInstrumentations } from '@opentelemetry/instrumentation';
import { FetchInstrumentation } from '@opentelemetry/instrumentation-fetch';
import { ZoneContextManager } from '@opentelemetry/context-zone';
import { Resource } from '@opentelemetry/resources';
import { SEMRESATTRS_SERVICE_NAME } from '@opentelemetry/semantic-conventions';

const provider = new WebTracerProvider({
    resource: new Resource({
        [SEMRESATTRS_SERVICE_NAME]: 'shop-ui',
    }),
});

const exporter = new OTLPTraceExporter({
    url: 'http://localhost:4318/v1/traces', // Default Jaeger/OTLP endpoint
});

provider.addSpanProcessor(new BatchSpanProcessor(exporter));

// For development, also log to console
if (import.meta.env.DEV) {
    provider.addSpanProcessor(new SimpleSpanProcessor(new ConsoleSpanExporter()));
}

provider.register({
    contextManager: new ZoneContextManager(),
});

registerInstrumentations({
    instrumentations: [
        new FetchInstrumentation({
            indent: 2,
            propagateTraceHeaderCorsUrls: /.*/, // Propagate trace headers to all URLs
            clearTimingResources: true,
        }),
    ],
});

export const tracer = provider.getTracer('shop-ui');
