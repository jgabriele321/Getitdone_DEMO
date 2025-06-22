FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Set the working directory
WORKDIR /build

# Copy the Go module files first
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . ./

# Build the application with CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -tags sqlite_foreign_keys -ldflags '-s -w' -o app ./cmd/bot

# Start from a fresh Alpine image
FROM alpine:latest

# Install necessary packages
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    sqlite \
    sqlite-libs

# Set environment variables
ENV TZ=UTC
ENV PORT=8080

# Create a non-root user
RUN adduser -D -s /bin/sh appuser

# Set the working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /build/app ./app

# Copy the environment example file
COPY env.example .env.example

# Create data directory for SQLite and set permissions
RUN mkdir -p /app/data && \
    ls -la /app && \
    chmod +x /app/app && \
    chown -R appuser:appuser /app

# Use the non-root user
USER appuser

# Set database path to writable location
ENV DATABASE_PATH=/app/data/cache.db

# Expose the port
EXPOSE ${PORT}

# Run the application
CMD ["./app"]