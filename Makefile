.PHONY: build run test clean help

# Default target
help:
	@echo "Weather Service - Available commands:"
	@echo "  build    - Build the weather service binary"
	@echo "  run      - Run the weather service"
	@echo "  test     - Run tests against the running service"
	@echo "  clean    - Remove build artifacts"
	@echo "  deps     - Download dependencies"

# Download dependencies
deps:
	go mod tidy

# Build the service
build: deps
	go build -o weather-service main.go
	@echo "Build complete: weather-service"

# Run the service
run: build
	@echo "Starting weather service on :8080..."
	@echo "Open http://localhost:8080 in your browser"
	@echo "Press Ctrl+C to stop"
	./weather-service

# Run the service directly with Go (no build step)
dev:
	@echo "Starting weather service in development mode..."
	@echo "Open http://localhost:8080 in your browser"
	@echo "Press Ctrl+C to stop"
	go run main.go

# Test the running service
test:
	@echo "Testing weather service..."
	@chmod +x test.sh
	./test.sh

# Clean build artifacts
clean:
	rm -f weather-service
	@echo "Clean complete"

# Install dependencies and build
install: deps build
	@echo "Installation complete"
