# TeamEdu — Agent Instructions

## Project Overview

Microservices-based educational platform for team formation. Go backend (Fiber framework) + Nuxt 4 frontend.

**Architecture**: API Gateway (port 3000) → 4 backend services + 1 frontend. MongoDB for ms-core, PostgreSQL for ms-auth/ms-courses, Redis for caching/queues.

## Critical Rules

1. **Never run `bun` commands automatically** after generating frontend code
2. **Never compile or run tests automatically** after generating backend code
3. Use `bun`, not `npm` — the project uses Bun as package manager

## Service Ports

| Service | Port | Database |
|---------|------|----------|
| api-gateway | 3000 | — |
| ms-auth | 3001 | PostgreSQL (auth_db) |
| ms-courses | 3002 | PostgreSQL (courses_db) |
| ms-core | 3003 | MongoDB + Redis |
| ms-notis | 3004 | — |
| front-eduteam | 2999 | — |

## Quick Start

### All services (Windows)
```bash
start.bat
```
Opens Windows Terminal with tabs for each service on Windows.

### All services (WSL / Linux)
- Desde Windows con pestañas de Windows Terminal en WSL:
  ```bash
  start-wsl.bat
  ```
- Directamente dentro de la terminal de WSL:
  ```bash
  ./start.sh
  ```
- Compilar plugins para Linux:
  ```bash
  build-linux.bat     # Desde Windows (cross-compilando)
  ./build-linux.sh    # Desde WSL
  ```

### Individual services
```bash
# Frontend
cd front-eduteam && bun dev

# Backend services
cd api-gateway && go run main.go
cd ms-auth && go run main.go
cd ms-courses && go run main.go
cd ms-core && go run main.go
cd ms-notis && go run main.go
```

## ms-core Plugin System

ms-core uses `hashicorp/go-plugin` for dynamic plugins. Plugins are separate binaries communicating via `net/rpc`.

### Build plugins
```bash
cd ms-core

# Windows (development)
make plugins-windows

# Docker (recommended for parity)
make plugins-docker

# Single plugin
make plugin-enneagram

# Server + all plugins
make build-all
```

### Plugin locations
- Source: `ms-core/plugins_src/`
- Compiled: `ms-core/plugins_dist/`
- Plugins: enneagram, mbti, bfi44, peer_evaluation, sociogram, team_formation, group_compatibility, chronotype (dev)

### Plugin interface
All plugins implement: `Name()`, `Version()`, `Initialize()`, `Execute()`, `Shutdown()`

## Testing

### Frontend (Nuxt)
```bash
cd front-eduteam
bun test                    # Run tests
bun test:coverage           # With coverage
```
Tests use vitest with happy-dom environment. Config: `vitest.config.ts`

### Backend (Go)
```bash
cd <service>
make test                   # Run tests
make test-coverage          # With coverage report
```

## Linting & Formatting

### Go services
```bash
make lint                   # golangci-lint
make fmt                    # go fmt
```

### Frontend
```bash
cd front-eduteam
bun lint                    # If configured
```

## Development

### Hot reload (requires `air`)
```bash
cd <service>
make dev
```
Install: `go install github.com/cosmtrek/air@latest`

### Dependencies
```bash
# Go
cd <service> && make deps

# Frontend
cd front-eduteam && bun install
```

## Database Init

Docker Compose auto-initializes databases from:
- `db-auth/` → PostgreSQL auth_db
- `db-courses/` → PostgreSQL courses_db

MongoDB collections created by ms-core plugins at runtime.

## Environment Variables

Each service has `.env` file. Key variables:
- `JWT_SECRET` — shared across services
- `DATABASE_URL` — PostgreSQL connection
- `MONGO_URI` — MongoDB connection
- `REDIS_URL` — Redis connection
- `PLUGIN_DIR` — ms-core plugin directory

## Common Gotchas

1. **Plugin build mismatch**: Use `make plugins-docker` to ensure plugin binaries match the Docker image's Go version
2. **Port conflicts**: Each service runs on a fixed port; ensure no other processes use 2999-3004
3. **ms-core startup**: Requires MongoDB + Redis running; plugins must be compiled first
4. **Frontend SSR**: Nuxt SSR bundles Vue deps inline for Docker (see `nuxt.config.ts` nitro config)

## Project Structure

```
teamedu-main/
├── api-gateway/          # Go Fiber reverse proxy
├── ms-auth/              # Authentication (PostgreSQL)
├── ms-courses/           # Course management (PostgreSQL)
├── ms-core/              # Core logic + plugins (MongoDB)
│   ├── plugins_src/      # Plugin source code
│   ├── plugins_dist/     # Compiled plugins
│   └── internal/         # Core server code
├── ms-notis/             # Notifications (SMTP)
├── front-eduteam/        # Nuxt 4 frontend
├── db-auth/              # PostgreSQL init scripts
├── db-courses/           # PostgreSQL init scripts
├── scripts/              # Backup & seed scripts
└── docker-compose.yml    # Full stack orchestration
```
