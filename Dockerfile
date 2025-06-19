# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /src

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create app user
RUN addgroup -g 1001 -S app && \
    adduser -u 1001 -S app -G app

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /src/build/LLMrecon /app/LLMrecon

# Create directories for config and reports
RUN mkdir -p /app/config /app/reports && \
    chown -R app:app /app

# Switch to app user
USER app

# Expose any necessary ports (if applicable)
# EXPOSE 8080

# Set entrypoint
ENTRYPOINT ["/app/LLMrecon"]
CMD ["--help"]