# FauxFoundry Dockerfile
# Multi-stage build for optimal image size
# Created by copyleftdev

# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o fauxfoundry \
    ./cmd/fauxfoundry

# Runtime stage
FROM scratch

# Import from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy the binary
COPY --from=builder /app/fauxfoundry /fauxfoundry

# Copy examples and create necessary directories
COPY --from=builder /app/examples /app/examples
COPY --from=builder /app/README.md /app/README.md

# Create non-root user
USER nobody:nobody

# Create volumes for data
VOLUME ["/app/outputs", "/app/specs"]

# Set working directory
WORKDIR /app

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/fauxfoundry", "--version"]

# Default command
ENTRYPOINT ["/fauxfoundry"]
CMD ["--help"]

# Metadata
LABEL maintainer="copyleftdev <dj@codetestcode.com>"
LABEL description="FauxFoundry - Synthetic data generation powered by local LLMs"
LABEL version="1.0.0"
LABEL org.opencontainers.image.source="https://github.com/copyleftdev/faux-foundry"
LABEL org.opencontainers.image.documentation="https://github.com/copyleftdev/faux-foundry/blob/main/README.md"
LABEL org.opencontainers.image.licenses="MIT"
