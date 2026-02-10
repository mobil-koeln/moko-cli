# Build stage
FROM golang:1.25-alpine AS builder

# Install git for go mod download
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s -X main.version=$(cat VERSION 2>/dev/null || echo 0.3.0)" -o moko ./cmd/moko

# Runtime stage
FROM alpine:3.19

# Install ca-certificates for HTTPS requests and tzdata for timezone support
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' moko

# Copy binary from builder
COPY --from=builder /app/moko /usr/local/bin/moko

# Use non-root user
USER moko

# Set working directory
WORKDIR /home/moko

# Create cache directory
RUN mkdir -p /home/moko/.cache/moko

# Set timezone
ENV TZ=Europe/Berlin

# Default command
ENTRYPOINT ["moko"]
CMD ["--help"]
