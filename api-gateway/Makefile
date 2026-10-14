# Makefile para API Gateway

.PHONY: build run clean test install deps

# Variables
BINARY_NAME=api-gateway
MAIN_PATH=./main.go

# Comandos por defecto
all: clean deps build

# Instalar dependencias
deps:
	go mod download
	go mod tidy

# Compilar la aplicación
build:
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

# Ejecutar la aplicación
run:
	go run $(MAIN_PATH)

# Ejecutar con hot reload (requiere air)
dev:
	air

# Limpiar archivos compilados
clean:
	rm -rf bin/
	go clean

# Ejecutar tests
test:
	go test -v ./...

# Ejecutar tests con coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Formatear código
fmt:
	go fmt ./...

# Verificar calidad del código
lint:
	golangci-lint run

# Instalar herramientas de desarrollo
install-tools:
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Ejecutar con variables de entorno específicas
run-prod:
	GATEWAY_PORT=3000 REGISTRO_CURRICULAR_URL=http://localhost:3001 go run $(MAIN_PATH)
