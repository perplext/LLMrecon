# Multi-stage build for LLMrecon
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make ca-certificates tzdata

# Set working directory
WORKDIR /src

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build arguments for version information
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown

# Build the application with version info
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -o llmrecon \
    ./src/main.go

# Runtime stage - distroless for better security
FROM gcr.io/distroless/static-debian12:latest AS runtime

# Copy CA certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Create necessary directories
USER 65534:65534

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /src/llmrecon /usr/local/bin/llmrecon

# Copy templates and examples (readable by user)
COPY --chown=65534:65534 examples/ ./examples/
COPY --chown=65534:65534 templates/ ./templates/

# Health check with minimal impact
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/usr/local/bin/llmrecon", "version"]

# Default entrypoint
ENTRYPOINT ["/usr/local/bin/llmrecon"]
CMD ["--help"]

# Alternative: Full Alpine runtime (uncomment if distroless causes issues)
FROM alpine:latest AS alpine-runtime

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl jq

# Create non-root user for security
RUN addgroup -g 1001 -S llmrecon && \
    adduser -u 1001 -S llmrecon -G llmrecon

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /src/llmrecon /usr/local/bin/llmrecon

# Copy templates and examples
COPY --chown=llmrecon:llmrecon examples/ ./examples/
COPY --chown=llmrecon:llmrecon templates/ ./templates/
COPY --chown=llmrecon:llmrecon docs/ ./docs/

# Create necessary directories
RUN mkdir -p /app/config /app/reports /app/logs /app/cache && \
    chown -R llmrecon:llmrecon /app

# Switch to non-root user
USER llmrecon

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD llmrecon version || exit 1

# Default entrypoint  
ENTRYPOINT ["llmrecon"]
CMD ["--help"]

# Labels for container metadata
LABEL org.opencontainers.image.title="LLMrecon" \
      org.opencontainers.image.description="Advanced LLM Security Testing Framework" \
      org.opencontainers.image.vendor="LLMrecon Project" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.source="https://github.com/perplext/LLMrecon" \
      org.opencontainers.image.documentation="https://github.com/perplext/LLMrecon/blob/main/README.md" \
      org.opencontainers.image.created="${DATE}"