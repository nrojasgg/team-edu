# =============================================================================
# ms-core — Dockerfile
#
# CGO_ENABLED=0 — no se requiere CGO. Los plugins son binarios independientes
# compilados en la misma etapa de build y copiados a plugins_dist/.
# Se usa hashicorp/go-plugin (subprocess + net/rpc) para cargarlos en runtime.
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1: builder
# -----------------------------------------------------------------------------
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Descargar dependencias antes de copiar el código fuente (capa de caché)
COPY go.mod go.sum ./
RUN go mod download

# Copiar el código fuente
COPY . .

# Compilar el binario principal (CGO no requerido)
# -ldflags="-s -w" reduce el tamaño del binario eliminando símbolos de debug.
RUN CGO_ENABLED=0 GOOS=linux \
    go build \
      -ldflags="-s -w" \
      -o /app/bin/ms-core \
      ./main.go

# Compilar cada plugin como binario independiente
RUN mkdir -p /app/plugins_dist && \
    CGO_ENABLED=0 GOOS=linux \
      go build -ldflags="-s -w" \
      -o /app/plugins_dist/enneagram \
      ./plugins_src/enneagram/ && \
    CGO_ENABLED=0 GOOS=linux \
      go build -ldflags="-s -w" \
      -o /app/plugins_dist/mbti \
      ./plugins_src/mbti/ && \
    CGO_ENABLED=0 GOOS=linux \
      go build -ldflags="-s -w" \
      -o /app/plugins_dist/bfi44 \
      ./plugins_src/bfi44/ && \
    CGO_ENABLED=0 GOOS=linux \
      go build -ldflags="-s -w" \
      -o /app/plugins_dist/peer_evaluation \
      ./plugins_src/peer_evaluation/ && \
    CGO_ENABLED=0 GOOS=linux \
      go build -ldflags="-s -w" \
      -o /app/plugins_dist/sociogram \
      ./plugins_src/sociogram/ && \
    CGO_ENABLED=0 GOOS=linux \
      go build -ldflags="-s -w" \
      -o /app/plugins_dist/team_formation \
      ./plugins_src/team_formation/ && \
    CGO_ENABLED=0 GOOS=linux \
      go build -ldflags="-s -w" \
      -o /app/plugins_dist/group_compatibility \
      ./plugins_src/group_compatibility/ && \
    CGO_ENABLED=0 GOOS=linux \
      go build -ldflags="-s -w" \
      -o /app/plugins_dist/chronotype \
      ./plugins_src/chronotype/

# -----------------------------------------------------------------------------
# Stage 2: runtime
# -----------------------------------------------------------------------------
FROM alpine:3.21

WORKDIR /app

# ca-certificates permite conexiones HTTPS (MongoDB Atlas).
# Run as non-root user
RUN apk --no-cache add ca-certificates && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

# Copiar el binario principal y los binarios de plugins
COPY --from=builder /app/bin/ms-core .
COPY --from=builder /app/plugins_dist ./plugins_dist

# Puerto expuesto por la aplicación Fiber
EXPOSE 3003

# Variable de entorno por defecto para el directorio de plugins.
# Puede sobreescribirse con -e PLUGIN_DIR=/ruta en docker run o en
# el bloque environment del docker-compose.
ENV PLUGIN_DIR=/app/plugins_dist

USER appuser

CMD ["./ms-core"]
