# Build stage
FROM oven/bun:1 AS builder

WORKDIR /app

# Copy package files
COPY package.json bun.lockb ./

# Install dependencies
RUN bun install --frozen-lockfile

# Copy only the source files required for the build (explicit copy avoids accidentally including sensitive data)
COPY nuxt.config.ts tsconfig.json app.vue ./
COPY components/ ./components/
COPY composables/ ./composables/
COPY content/ ./content/
COPY encuestas/ ./encuestas/
COPY middleware/ ./middleware/
COPY pages/ ./pages/
COPY plugins/ ./plugins/
COPY public/ ./public/

# Build the application
RUN bun run build

# Final stage — Node.js runtime (requerido por el preset node-server de Nitro)
FROM node:22-alpine

WORKDIR /app

# Copy only the built output — Nitro node-server bundlea todas las dependencias
COPY --from=builder /app/.output ./.output

# Expose port
EXPOSE 2999

ENV PORT=2999
ENV HOST=0.0.0.0
ENV NODE_ENV=production

CMD ["node", ".output/server/index.mjs"]
