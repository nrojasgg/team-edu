FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copiar archivos de dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fuente
COPY . .

# Construir la aplicación
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Imagen final
FROM alpine:latest

RUN apk --no-cache add ca-certificates

# Run as non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copiar el binario desde el builder
COPY --from=builder /app/main .

# Puerto en el que corre la aplicación
EXPOSE 8081

USER appuser

CMD ["./main"]
