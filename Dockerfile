# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

# Install git and ca-certificates (needed for go mod download with private repos)
RUN apk add --no-cache git ca-certificates tzdata

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies (this layer will be cached unless go.mod/go.sum changes)
RUN go mod download && go mod verify

# Copy the source code
COPY . .

# Build the application with optimizations for production
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o server ./server

# Stage 2: Create a minimal production image
FROM scratch

# Copy ca-certificates from builder for HTTPS requests to AWS services
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data for proper time handling
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the built binary from the builder stage
COPY --from=builder /app/server /server

# Create a non-root user ID for security (App Runner will use this)
USER 1000

# Expose the port the app runs on
EXPOSE 8080

# Health check for AWS App Runner
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/server", "--health-check"] || exit 1

# Start the application
ENTRYPOINT ["/server"]