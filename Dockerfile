# Use Go's official image as the builder
FROM golang:1.23 AS builder

# Set the current working directory inside the container
WORKDIR /app

# Copy the Go module files and download the dependencies
COPY go.mod go.sum ./
RUN go mod tidy

# Copy the source code into the container
COPY . .

# Build the Go binary for Linux architecture (cross-compile for Linux)
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o kvstore ./cmd/kvstorecli

# Use a minimal base image for the final container (Alpine)
FROM debian:bullseye-slim

# Install necessary dependencies (e.g., ca-certificates) to run the Go binary
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy the built binary from the builder stage to the target image
COPY --from=builder /app/go-template /usr/local/bin/go-template

# Ensure the binary has execute permissions
RUN chmod +x /usr/local/bin/go-template

# Expose the application port (assuming your Go service runs on port 8080)
EXPOSE 8080

# Command to run the binary
CMD ["/usr/local/bin/go-template"]
