# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copiar go mod y sum files
COPY go.mod go.sum ./

# Descargar dependencias
RUN go mod download

# Copiar el código fuente
COPY . .

# Compilar la aplicación
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:3.21

# Instalar ca-certificates para HTTPS
RUN apk --no-cache add ca-certificates

# Run as non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copiar el binario compilado
COPY --from=builder /app/main .

# Exponer el puerto
EXPOSE 3000

USER appuser

# Comando para ejecutar la aplicación
CMD ["./main"]
