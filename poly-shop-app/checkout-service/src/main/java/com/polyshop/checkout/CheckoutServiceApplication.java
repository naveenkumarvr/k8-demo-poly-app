package com.polyshop.checkout;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.Bean;
import org.springframework.web.client.RestTemplate;

/**
 * Poly-Shop Checkout Service - Main Application
 * 
 * A production-ready Spring Boot microservice for payment simulation and currency conversion.
 * This service demonstrates:
 * 
 * 1. OpenTelemetry Auto-Instrumentation:
 *    - The OpenTelemetry Java Agent is attached via JVM argument (-javaagent:/app/opentelemetry-javaagent.jar)
 *    - The agent uses BYTE-CODE INSTRUMENTATION to automatically add tracing to:
 *      * HTTP requests/responses (RestTemplate, controllers)
 *      * Database queries (JPA/Hibernate)
 *      * Method calls (with annotations)
 *    - NO CODE CHANGES REQUIRED - tracing is automatic!
 *    - Trace context is propagated via W3C Trace Context headers
 * 
 * 2. Kubernetes Health Probes:
 *    - Liveness Probe: /actuator/health/liveness (checks if app is running)
 *    - Readiness Probe: /actuator/health/readiness (checks if app is ready for traffic)
 *    - Startup Probe: /actuator/health/readiness (with initial delay for slow startup)
 * 
 * 3. Behavioral Testing:
 *    - Slow Startup: StartupDelayBean sleeps for 15 seconds to test K8s startupProbe
 *    - Memory Load: Admin endpoint can allocate large byte arrays to test OOMKill scenarios
 * 
 * @see <a href="https://opentelemetry.io/docs/instrumentation/java/automatic/">OTel Java Agent Docs</a>
 * @version 1.0.0
 */
@SpringBootApplication
public class CheckoutServiceApplication {

    public static void main(String[] args) {
        /*
         * Spring Boot startup sequence:
         * 1. Application context initialization
         * 2. Bean creation (@Bean methods, @Component scan)
         * 3. StartupDelayBean @PostConstruct runs (15-second sleep)
         * 4. Embedded Tomcat starts on port 8085
         * 5. Actuator endpoints become available
         * 
         * OpenTelemetry Agent Lifecycle:
         * - Attached BEFORE main() via JAVA_TOOL_OPTIONS
         * - Transforms byte-code of Spring classes during class loading
         * - Adds tracing logic to RestTemplate, controllers, JPA repositories
         * - Exports spans to Jaeger at OTEL_EXPORTER_OTLP_ENDPOINT
         */
        SpringApplication.run(CheckoutServiceApplication.class, args);
        System.out.println("✅ Checkout Service started successfully!");
        System.out.println("📊 OpenTelemetry agent is instrumenting HTTP and DB calls");
        System.out.println("🔍 View traces at http://localhost:16686 (Jaeger UI)");
    }

    /**
     * RestTemplate Bean Configuration
     * 
     * This RestTemplate is used to call cart-service for cart verification.
     * The OpenTelemetry Java Agent automatically instruments RestTemplate to:
     * 1. Create a new span for each HTTP request
     * 2. Inject trace context headers (traceparent, tracestate)
     * 3. Propagate the trace ID to downstream services
     * 
     * Example trace flow:
     * POST /checkout → checkout-service span
     *   └─ GET http://cart-service:8080/v1/cart/{userId} → cart-service span
     * 
     * Both spans share the same trace ID, enabling distributed tracing!
     * 
     * @return RestTemplate instance for HTTP calls
     */
    @Bean
    public RestTemplate restTemplate() {
        return new RestTemplate();
    }
}
