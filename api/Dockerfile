# ----------------------
# Stage 1: Build the Go binary
# ----------------------
FROM golang:alpine AS builder

# Install git (if needed for dependencies)
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go.mod files first (for caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the Go binary
RUN go build -o main .

# ----------------------
# Stage 2: Production image
# ----------------------
FROM alpine:latest

# Install certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create app directory and user
RUN adduser -S -h /app appuser
RUN mkdir -p /app

# Copy binary and .env from builder stage with correct ownership
COPY --from=builder --chown=appuser:appuser /build/main /app/
COPY --from=builder --chown=appuser:appuser /build/.env /app/

# Switch to non-root user
USER appuser
WORKDIR /app

# Expose the application port
EXPOSE 3000

# Run the application
CMD ["./main"]
