ifeq ($(OS),Windows_NT)
    BINARY_EXT = .exe
    RM_CMD = if exist ms-notis.exe del /f ms-notis.exe
else
    BINARY_EXT =
    RM_CMD = rm -f ms-notis ms-notis.exe
endif

.PHONY: run build build-linux clean install dev

run:
	go run main.go

build:
	go build -o ms-notis$(BINARY_EXT) main.go

build-linux:
	CGO_ENABLED=0 GOOS=linux go build -o ms-notis main.go

clean:
	@$(RM_CMD)

install:
	go mod download

dev:
	go run main.go
