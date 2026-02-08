# Multi-stage Dockerfile for Product Service
# Uses Alpine for building and Distroless for runtime

# Stage 1: Builder
# Use official Golang image for compilation
FROM golang:1.21-alpine AS builder

# Install build dependencies
# ca-certificates: Required for HTTPS connections
# git: Required for fetching Go dependencies
RUN apk add --no-cache ca-certificates git

# Set working directory
WORKDIR /build

# Copy go.mod and go.sum first for better layer caching
# This allows Docker to cache dependencies if they haven't changed
COPY go.mod ./

# Download dependencies and generate go.sum if it doesn't exist
# This layer is cached unless go.mod changes
RUN go mod download && go mod tidy

# Copy the entire source code
COPY . .

# Build the application
# CGO_ENABLED=0: Build a statically linked binary (required for distroless)
# -ldflags="-w -s": Strip debug information to reduce binary size
#   -w: Omit DWARF symbol table
#   -s: Omit symbol table and debug information
# -o: Output binary name
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o product-service \
    .

# Stage 2: Runtime
# Use Google Distroless base image for minimal attack surface
# This image contains only the application and its runtime dependencies
# No shell, package manager, or other unnecessary tools
FROM gcr.io/distroless/static-debian12:nonroot

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /build/product-service /app/product-service

# Copy CA certificates for HTTPS connections
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Use non-root user (UID 65532 is 'nonroot' in distroless)
# This enhances security by preventing privilege escalation
USER 65532:65532

# Expose the application port
EXPOSE 8080

# Set the entrypoint
# Note: Distroless images don't have a shell, so we use the exec form
ENTRYPOINT ["/app/product-service"]
