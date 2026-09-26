# Multi-stage Dockerfile for Notification Worker
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /build

COPY go.mod ./
RUN go mod download && go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o notification-worker \
    .

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /build/notification-worker /app/notification-worker
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["/app/notification-worker"]
