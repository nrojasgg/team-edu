# Makefile for MS Registro Curricular

# Variables
BINARY_NAME=api
SETUP_BINARY=setup
MAIN_PACKAGE=./main.go
SETUP_PACKAGE=./cmd/setup/main.go

# Default target
.PHONY: all
all: build

# Build the application
.PHONY: build
build:
	@echo "🔨 Building application..."
	@go build -o bin/$(BINARY_NAME) $(MAIN_PACKAGE)
	@go build -o bin/$(SETUP_BINARY) $(SETUP_PACKAGE)
	@echo "✅ Build complete!"

# Run the application
.PHONY: run
run: build
	@echo "🚀 Starting API server..."
	@./bin/$(BINARY_NAME)

# Run setup
.PHONY: setup
setup: build
	@echo "🔧 Running setup..."
	@./bin/$(SETUP_BINARY)

# Clean build artifacts
.PHONY: clean
clean:
	@echo "🧹 Cleaning..."
	@rm -rf bin/
	@echo "✅ Clean complete!"

# Run tests
.PHONY: test
test:
	@echo "🧪 Running tests..."
	@go test ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	@go test -cover ./...

# Install dependencies
.PHONY: deps
deps:
	@echo "📦 Downloading dependencies..."
	@go mod download
	@go mod tidy

# Build for production
.PHONY: build-prod
build-prod:
	@echo "🏗️ Building for production..."
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "✅ Production build complete!"

# Build Docker image
.PHONY: docker-build
docker-build:
	@echo "🐳 Building Docker image..."
	@docker build -t ms-registro-curricular .
	@echo "✅ Docker build complete!"

# Run with Docker
.PHONY: docker-run
docker-run: docker-build
	@echo "🐳 Running with Docker..."
	@docker run -p 3000:3000 --env-file .env ms-registro-curricular

# Format code
.PHONY: fmt
fmt:
	@echo "💅 Formatting code..."
	@go fmt ./...

# Lint code
.PHONY: lint
lint:
	@echo "🔍 Linting code..."
	@golangci-lint run

# Development mode
.PHONY: dev
dev:
	@echo "🔧 Starting development mode..."
	@air

# Help
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  build         - Build the application"
	@echo "  run           - Build and run the application"
	@echo "  setup         - Run the setup utility"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  deps          - Download dependencies"
	@echo "  build-prod    - Build for production"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run with Docker"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  dev           - Start development mode (requires air)"
	@echo "  help          - Show this help message"
