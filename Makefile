# Gripmock Makefile

# Variables
BINARY_NAME=gripmock
DOCKER_IMAGE=tkpd/gripmock
PLATFORMS=linux/amd64,linux/arm64
GOPATH:=$(shell go env GOPATH)

# Include test makefile
include Makefile.test

.PHONY: all build clean test push help

# Default target
all: build

# Build the Go binary
build:
	@echo "Building Go binary..."
	go build -o $(BINARY_NAME) .

# Build Docker image
docker-build:
	@if [ "$(VERSION)" = "" ]; then \
		echo "Error: VERSION is required. Usage: make docker-build VERSION=x.y.z"; \
		exit 1; \
	fi
	@echo "Building Docker image..."
	docker buildx build --load -t $(DOCKER_IMAGE):$(VERSION) --platform linux/amd64 .

# Push Docker image
docker-push:
	@if [ "$(VERSION)" = "" ]; then \
		echo "Error: VERSION is required. Usage: make docker-push VERSION=x.y.z"; \
		exit 1; \
	fi
	@echo "Pushing Docker image..."
	docker buildx build --push -t $(DOCKER_IMAGE):$(VERSION) --platform $(PLATFORMS) .

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	go clean
	docker rm -f gripmock-test || true

# Show help
help:
	@echo "Available targets:"
	@echo "  all            - Build the project (default)"
	@echo "  build          - Build the Go binary"
	@echo "  docker-build   - Build Docker image (requires VERSION=x.y.z)"
	@echo "  docker-push    - Push Docker image (requires VERSION=x.y.z)"
	@echo "  test           - Run tests"
	@echo "  clean          - Clean build artifacts"
	@echo "  help           - Show this help message"
	@echo ""
	@echo "Test targets:"
	@echo "  test-simple    - Run simple example test"
	@echo "  build-test-image - Build test Docker image" 