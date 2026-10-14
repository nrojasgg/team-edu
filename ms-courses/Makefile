# Makefile para ms-courses

# Variables
BINARY_NAME=ms-courses
MAIN_PATH=./main.go
BUILD_DIR=./bin

# Detección de sistema operativo
ifeq ($(OS),Windows_NT)
    BINARY_EXT = .exe
else
    BINARY_EXT =
endif

# Comandos por defecto
.PHONY: all build build-linux run clean test help

all: build

# Construir el binario
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT)"

# Construir para Linux
build-linux:
	@echo "Building $(BINARY_NAME) for Linux..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Ejecutar la aplicación
run:
	@echo "Running $(BINARY_NAME)..."
	@go run $(MAIN_PATH)

# Limpiar archivos generados
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Ejecutar tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Ejecutar tests con coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Instalar dependencias
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

# Mostrar ayuda
help:
	@echo "Available commands:"
	@echo "  build         - Build the binary"
	@echo "  run           - Run the application"
	@echo "  clean         - Clean build files"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  deps          - Install dependencies"
	@echo "  help          - Show this help message"