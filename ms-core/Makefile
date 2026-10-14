# Makefile para ms-core

# ---------------------------------------------------------------------------
# Variables
# ---------------------------------------------------------------------------
BINARY_NAME    = ms-core
MAIN_PATH      = ./main.go
BUILD_DIR      = ./bin

# Directorio con los fuentes de los plugins dinámicos
PLUGINS_SRC_DIR = ./plugins_src
# Directorio de destino para los .so compilados (debe coincidir con PLUGIN_DIR en .env)
PLUGINS_DIST_DIR = ./plugins_dist

# Imagen Go utilizada tanto para el core como para compilar los .so
# IMPORTANTE: debe ser la misma en el Dockerfile para garantizar paridad de entorno
GO_IMAGE = golang:1.24-alpine

# Detección de sistema operativo
ifeq ($(OS),Windows_NT)
    BINARY_EXT = .exe
    RM_DIR     = if exist $(1) rmdir /s /q $(1)
    RM_FILE    = if exist $(1) del /f $(1)
else
    BINARY_EXT =
    RM_DIR     = rm -rf $(1)
    RM_FILE    = rm -f $(1)
endif

# Lista de plugins (subdirectorios bajo PLUGINS_SRC_DIR)
PLUGINS = enneagram mbti bfi44 peer_evaluation sociogram team_formation group_compatibility chronotype

# ---------------------------------------------------------------------------
# Targets principales
# ---------------------------------------------------------------------------
.PHONY: all build build-linux build-all build-all-linux run clean test help \
        plugins plugins-windows plugins-docker plugin-% \
        docker-build docker-run \
        test-coverage deps lint fmt dev install-tools

all: build

# ---------------------------------------------------------------------------
# Core — binario principal
# ---------------------------------------------------------------------------

## build: Compila el binario principal según el sistema operativo actual
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT)"

## build-linux: Compila el binario principal para Linux (CGO_ENABLED=0 GOOS=linux)
build-linux:
	@echo "Building $(BINARY_NAME) for Linux..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## build-all: Compila el servidor principal y todos los plugins Windows en un solo paso
build-all: build plugins-windows
	@echo "Server + all Windows plugins compiled and ready."

## build-all-linux: Compila el servidor principal y todos los plugins Linux en un solo paso
build-all-linux: build-linux plugins
	@echo "Server + all Linux plugins compiled and ready."

## run: Ejecuta la aplicación directamente con go run
run:
	@echo "Running $(BINARY_NAME)..."
	@go run $(MAIN_PATH)

## run-binary: Compila y ejecuta el binario
run-binary: build
	@echo "Running compiled binary..."
	@$(BUILD_DIR)/$(BINARY_NAME)$(BINARY_EXT)

# ---------------------------------------------------------------------------
# Plugins — compilación local (CGO_ENABLED=0, funciona en Windows/macOS/Linux)
# ---------------------------------------------------------------------------

## plugins: Compila todos los plugins como binarios independientes (Linux)
##          Para desarrollo local en Windows usa plugins-windows.
plugins: $(addprefix plugin-,$(PLUGINS))

## plugin-%: Compila un plugin para Linux. Ejemplo: make plugin-enneagram
plugin-%:
	@echo "Compiling plugin: $*"
	@mkdir -p $(PLUGINS_DIST_DIR)
	@cd $(PLUGINS_SRC_DIR)/$* && \
		CGO_ENABLED=0 GOOS=linux \
		go build \
			-o ../../$(PLUGINS_DIST_DIR)/$* \
			.
	@echo "  -> $(PLUGINS_DIST_DIR)/$*"

## plugins-windows: Compila todos los plugins como .exe para desarrollo local en Windows
plugins-windows:
	@echo "Compiling plugins for Windows (development)..."
	@mkdir -p $(PLUGINS_DIST_DIR)
	@for plugin in $(PLUGINS); do \
		echo "  -> Building $$plugin.exe ..."; \
		(cd $(PLUGINS_SRC_DIR)/$$plugin && \
			CGO_ENABLED=0 GOOS=windows \
			go build \
				-o ../../$(PLUGINS_DIST_DIR)/$$plugin.exe \
				.); \
	done
	@echo "All Windows plugins compiled into $(PLUGINS_DIST_DIR)/"

# ---------------------------------------------------------------------------
# Plugins dinámicos — compilación vía Docker (recomendado, garantiza paridad)
# ---------------------------------------------------------------------------

## plugins-docker: Compila todos los plugins .so dentro de un contenedor Docker.
##                 Garantiza que el .so es compatible con el binario del servidor.
##                 Funciona en Windows, macOS y Linux.
plugins-docker:
	@echo "Compiling all plugins inside Docker ($(GO_IMAGE))..."
	@mkdir -p $(PLUGINS_DIST_DIR)
	@for plugin in $(PLUGINS); do \
		echo "  -> Building $$plugin ..."; \
		docker run --rm \
			-v "$(CURDIR):/workspace" \
			-w /workspace \
			$(GO_IMAGE) \
			sh -c " \
				cd plugins_src/$$plugin && \
				CGO_ENABLED=0 GOOS=linux \
				go build \
					-o /workspace/$(PLUGINS_DIST_DIR)/$$plugin \
					. \
			"; \
	done
	@echo "All plugins compiled into $(PLUGINS_DIST_DIR)/"

## plugin-docker-%: Compila un único plugin vía Docker. Ejemplo: make plugin-docker-enneagram
plugin-docker-%:
	@echo "Compiling plugin $* inside Docker ($(GO_IMAGE))..."
	@mkdir -p $(PLUGINS_DIST_DIR)
	@docker run --rm \
		-v "$(CURDIR):/workspace" \
		-w /workspace \
		$(GO_IMAGE) \
		sh -c " \
			cd plugins_src/$* && \
			CGO_ENABLED=0 GOOS=linux \
			go build \
				-o /workspace/$(PLUGINS_DIST_DIR)/$* \
				. \
		"
	@echo "  -> $(PLUGINS_DIST_DIR)/$*"

# ---------------------------------------------------------------------------
# Docker — imagen del servidor
# ---------------------------------------------------------------------------

## docker-build: Construye la imagen Docker del servidor ms-core
docker-build:
	@echo "Building Docker image ms-core:latest ..."
	@docker build -t ms-core:latest .

## docker-run: Compila plugins, construye imagen y arranca el contenedor
docker-run: plugins-docker docker-build
	@echo "Running Docker container..."
	@docker run -p 3003:3003 \
		--env-file .env \
		-v "$(CURDIR)/$(PLUGINS_DIST_DIR):/app/plugins_dist:ro" \
		ms-core:latest

# ---------------------------------------------------------------------------
# Calidad de código
# ---------------------------------------------------------------------------

## test: Ejecuta todos los tests
test:
	@echo "Running tests..."
	@go test -v ./...

## test-coverage: Tests con reporte de cobertura HTML
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: Ejecuta golangci-lint
lint:
	@echo "Running linter..."
	@golangci-lint run

## fmt: Formatea el código fuente
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# ---------------------------------------------------------------------------
# Dependencias y herramientas
# ---------------------------------------------------------------------------

## deps: Descarga y organiza dependencias de Go
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

## install-tools: Instala herramientas de desarrollo (air, golangci-lint)
install-tools:
	@echo "Installing development tools..."
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

## dev: Inicia el servidor con hot reload (requiere air)
dev:
	@echo "Starting development server with hot reload..."
	@air

# ---------------------------------------------------------------------------
# Limpieza
# ---------------------------------------------------------------------------

## clean: Elimina binarios, plugins compilados y reportes de cobertura
clean:
	@echo "Cleaning build artifacts..."
ifeq ($(OS),Windows_NT)
	@if exist $(BUILD_DIR) rmdir /s /q $(BUILD_DIR)
	@if exist $(PLUGINS_DIST_DIR) rmdir /s /q $(PLUGINS_DIST_DIR)
	@if exist coverage.out del /f coverage.out
	@if exist coverage.html del /f coverage.html
else
	@rm -rf $(BUILD_DIR) $(PLUGINS_DIST_DIR) coverage.out coverage.html
endif
	@echo "Clean complete"

# ---------------------------------------------------------------------------
# Ayuda
# ---------------------------------------------------------------------------

## help: Muestra esta ayuda
help:
	@echo ""
	@echo "ms-core — comandos disponibles:"
	@echo ""
	@echo "  Core:"
	@echo "    build              Compila el binario principal"
	@echo "    run                Ejecuta con go run"
	@echo "    run-binary         Compila y ejecuta el binario"
	@echo ""
	@echo "  Plugins (local — requiere CGO + gcc en Linux/macOS):"
	@echo "    plugins            Compila todos los plugins .so localmente"
	@echo "    plugin-<name>      Compila un plugin específico (ej: plugin-enneagram)"
	@echo ""
	@echo "  Plugins (Docker — recomendado, funciona en Windows):"
	@echo "    plugins-docker     Compila todos los plugins .so via Docker"
	@echo "    plugin-docker-<n>  Compila un plugin específico via Docker"
	@echo ""
	@echo "  Docker:"
	@echo "    docker-build       Construye la imagen Docker del servidor"
	@echo "    docker-run         Plugins + imagen + contenedor en un paso"
	@echo ""
	@echo "  Calidad:"
	@echo "    test               Ejecuta todos los tests"
	@echo "    test-coverage      Tests con reporte de cobertura HTML"
	@echo "    lint               golangci-lint"
	@echo "    fmt                Formatea código fuente"
	@echo ""
	@echo "  Utilitarios:"
	@echo "    deps               go mod tidy + download"
	@echo "    dev                Hot reload con air"
	@echo "    install-tools      Instala air y golangci-lint"
	@echo "    clean              Elimina artefactos generados"
	@echo "    help               Muestra esta ayuda"
	@echo ""
