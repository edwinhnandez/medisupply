# Build stage
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies (git not needed since go.mod/go.sum are present)
# RUN apk add --no-cache git

# Copy go mod files and vendor directory
COPY go.mod go.sum ./
COPY vendor ./vendor

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./purchase-order-service

# Final stage
FROM nginx:alpine

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/main .

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# Run the application
CMD ["./main"]
