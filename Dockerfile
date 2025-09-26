# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Copy migrations to expected location
COPY db/migrations /app/migrations

# Build the application for Linux/AMD64 (AWS App Runner architecture)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./server

# Stage 2: Create a lightweight image for deployment
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y curl ca-certificates && update-ca-certificates

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/server .

# Copy migrations to the deployment image
COPY --from=builder /app/migrations /app/migrations

# Expose the port the app runs on
EXPOSE 8080

# Start the application
CMD ["./server"]