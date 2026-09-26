# Multi-Stage Dockerfile for Checkout Service with OpenTelemetry
# 
# Stage 1: Build with Maven
# Stage 2: Runtime with JRE + OpenTelemetry Java Agent
#
# Key Features:
# - Java 21 with Spring Boot 3.x
# - OpenTelemetry auto-instrumentation via Java Agent
# - JVM heap tuned for Kubernetes resource limits
# - Non-root user for security

# ============================================================================
# Stage 1: Builder - Compile Java application with Maven
# ============================================================================
FROM maven:3.9-eclipse-temurin-21-alpine AS builder

WORKDIR /build

# Copy Maven files for dependency caching
# Docker caches this layer if pom.xml hasn't changed
COPY pom.xml .

# Download dependencies (cached layer)
RUN mvn dependency:go-offline -B

# Copy source code
COPY src ./src

# Build application
# - Skip tests during Docker build (run tests separately via CI/CD)
# - Create executable JAR with embedded Tomcat
RUN mvn clean package -DskipTests -B

# ============================================================================
# Stage 2: Runtime - OpenTelemetry instrumented JRE
# ============================================================================
FROM eclipse-temurin:21-jre-alpine

# Install wget for health checks (Kubernetes probes)
RUN apk add --no-cache wget

WORKDIR /app

# Copy JAR from builder stage
COPY --from=builder /build/target/*.jar /app/checkout-service.jar

# ============================================================================
# OpenTelemetry Java Agent Setup
# ============================================================================
# Download OpenTelemetry Java Agent
# This agent performs BYTE-CODE INSTRUMENTATION at JVM startup:
# 
# How it works:
# 1. Agent is loaded via -javaagent JVM argument BEFORE main() runs
# 2. Agent uses Java Instrumentation API to transform classes during loading
# 3. Injects tracing logic into:
#    - HTTP clients (RestTemplate, HttpClient)
#    - HTTP servers (Spring MVC controllers)
#    - JDBC drivers (database queries)
#    - Logging frameworks (correlate logs with traces)
# 4. Automatically propagates W3C Trace Context headers
# 5. Exports spans to OTLP endpoint (Jaeger)
# 
# NO CODE CHANGES REQUIRED - 100% automatic!
# 
# Version: 2.2.0 (latest stable as of implementation)
# ============================================================================

ARG OTEL_AGENT_VERSION=2.2.0
RUN wget -q https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/download/v${OTEL_AGENT_VERSION}/opentelemetry-javaagent.jar \
    -O /app/opentelemetry-javaagent.jar

# ============================================================================
# JVM Configuration for Kubernetes
# ============================================================================
# Heap Size Tuning:
# - Xms (initial heap): 512MB
# - Xmx (max heap): 768MB
# - Container memory limit should be ~1GB (768MB heap + ~256MB non-heap)
# 
# Why set max heap close to container limit?
# 1. Prevents OOMKill due to JVM exceeding container memory
# 2. Allows testing realistic Java-in-K8s memory behavior
# 3. Triggers GC before reaching container limit
# 
# Non-Heap Memory Breakdown:
# - Metaspace: ~128MB (class metadata)
# - Thread stacks: ~50MB (depends on thread count)
# - Code cache: ~50MB (JIT compiled code)
# - Native memory: ~30MB (nio buffers, etc.)
# 
# Monitoring JVM Memory Inside Container:
# docker exec <container> java -XshowSettings:vm -version
# docker exec <container> jcmd 1 VM.native_memory summary
# ============================================================================

ENV JAVA_OPTS="-Xms512m -Xmx768m \
    -XX:+UseContainerSupport \
    -XX:MaxRAMPercentage=75.0 \
    -XX:+HeapDumpOnOutOfMemoryError \
    -XX:HeapDumpPath=/tmp/heapdump.hprof"

# ============================================================================
# OpenTelemetry Configuration
# ============================================================================
# JAVA_TOOL_OPTIONS: JVM picks this up automatically (higher priority than CMD)
# -javaagent: Loads OTel agent BEFORE application starts
# 
# Environment variables for OTel (set at runtime via Docker Compose/K8s):
# - OTEL_EXPORTER_OTLP_ENDPOINT: Where to send traces (e.g., jaeger:4317)
# - OTEL_SERVICE_NAME: Service identifier in traces
# - OTEL_RESOURCE_ATTRIBUTES: Additional metadata (pod name, namespace, etc.)
# ============================================================================

ENV JAVA_TOOL_OPTIONS="-javaagent:/app/opentelemetry-javaagent.jar"

# Create non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

# Expose application port
EXPOSE 8085

# Health check (for Docker Compose, K8s uses its own probes)
HEALTHCHECK --interval=10s --timeout=5s --start-period=20s --retries=3 \
    CMD wget --spider -q http://localhost:8085/actuator/health || exit 1

# Run application
# The JAVA_TOOL_OPTIONS automatically adds the -javaagent argument
# Startup sequence:
# 1. JVM loads OpenTelemetry agent
# 2. Agent instruments Spring Boot classes
# 3. StartupDelayBean sleeps for 15 seconds
# 4. Spring Boot starts on port 8085
# 5. Actuator endpoints become available
# 6. Health checks pass → container marked "healthy"
ENTRYPOINT ["sh", "-c", "java $JAVA_OPTS -jar /app/checkout-service.jar"]
