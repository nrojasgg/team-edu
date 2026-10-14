#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="$SCRIPT_DIR/logs"

mkdir -p "$LOG_DIR"

# Colores para la terminal
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}========================================================${NC}"
echo -e "${CYAN}   Iniciando todos los servicios de TeamEdu (Linux/WSL)  ${NC}"
echo -e "${CYAN}========================================================${NC}"
echo ""

# 1. Verificar herramientas necesarias
echo -e "${BLUE}[1/4] Verificando dependencias en WSL/Linux...${NC}"

if ! command -v go &> /dev/null; then
    echo -e "${RED}[ERROR] Go no está instalado en este entorno Linux/WSL.${NC}"
    echo "Instálalo con: sudo apt install golang-go o descarga Go oficial."
    exit 1
fi

if ! command -v bun &> /dev/null; then
    echo -e "${YELLOW}[AVISO] Bun no está en el PATH.${NC}"
    # Verificar si bun existe en ~/.bun/bin
    if [ -f "$HOME/.bun/bin/bun" ]; then
        export PATH="$HOME/.bun/bin:$PATH"
        echo -e "${GREEN}Bun encontrado en $HOME/.bun/bin${NC}"
    else
        echo -e "${RED}[ERROR] Bun no encontrado. Instálalo con: curl -fsSL https://bun.sh/install | bash${NC}"
        exit 1
    fi
fi

echo -e "${GREEN}  ✓ Go ($(go version | awk '{print $3}')) detectado${NC}"
echo -e "${GREEN}  ✓ Bun ($(bun --version)) detectado${NC}"
echo ""

# 2. Compilar plugins de ms-core para Linux
echo -e "${BLUE}[2/4] Compilando plugins de MS-Core para Linux...${NC}"
bash "$SCRIPT_DIR/build-linux.sh"
echo ""

# 3. Manejo de señales para limpieza limpia al salir (Ctrl+C)
PIDS=()

cleanup() {
    echo ""
    echo -e "${YELLOW}Deteniendo todos los servicios de TeamEdu...${NC}"
    for pid in "${PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            kill -TERM "$pid" 2>/dev/null || kill -9 "$pid" 2>/dev/null
        fi
    done
    wait 2>/dev/null
    echo -e "${GREEN}Todos los servicios se han detenido correctamente.${NC}"
    exit 0
}

trap cleanup SIGINT SIGTERM EXIT

# 4. Iniciar servicios en segundo plano con logs redirigidos
echo -e "${BLUE}[3/4] Lanzando servicios...${NC}"

export CGO_ENABLED=0

start_service() {
    local name="$1"
    local dir="$2"
    local cmd="$3"
    local logfile="$LOG_DIR/${name}.log"

    echo -e "  -> Iniciando ${CYAN}$name${NC}..."
    (
        cd "$SCRIPT_DIR/$dir"
        eval "$cmd" > "$logfile" 2>&1
    ) &
    local pid=$!
    PIDS+=("$pid")
    echo -e "     PID: $pid | Log: logs/${name}.log"
}

start_service "frontend"     "front-eduteam" "bunx nuxt dev --no-fork"
start_service "api-gateway"  "api-gateway"   "go run main.go"
start_service "ms-auth"      "ms-auth"       "go run main.go"
start_service "ms-courses"   "ms-courses"    "go run main.go"
start_service "ms-core"      "ms-core"       "go run main.go"
start_service "ms-notis"     "ms-notis"      "go run main.go"

echo ""
echo -e "${GREEN}========================================================${NC}"
echo -e "${GREEN}   ¡Todos los servicios han sido iniciados en WSL!      ${NC}"
echo -e "${GREEN}========================================================${NC}"
echo ""
echo -e "Servicios disponibles (accesibles desde Windows y WSL):"
echo -e "  - Frontend:        ${CYAN}http://127.0.0.1:2999${NC}"
echo -e "  - API Gateway:     ${CYAN}http://127.0.0.1:3000${NC}"
echo -e "  - MS-Auth:         ${CYAN}http://127.0.0.1:3001${NC}"
echo -e "  - MS-Courses:      ${CYAN}http://127.0.0.1:3002${NC}"
echo -e "  - MS-Core:         ${CYAN}http://127.0.0.1:3003${NC}"
echo -e "  - MS-Notis:        ${CYAN}http://127.0.0.1:3004${NC}"
echo ""
echo -e "Logs en tiempo real guardados en: ${YELLOW}$LOG_DIR/*.log${NC}"
echo -e "${YELLOW}Presiona Ctrl+C en esta terminal para detener todos los servicios.${NC}"
echo ""

# Mantener script corriendo y monitorear servicios
while true; do
    for i in "${!PIDS[@]}"; do
        pid="${PIDS[$i]}"
        if ! kill -0 "$pid" 2>/dev/null; then
            echo -e "${RED}[ALERTA] Uno de los procesos (PID $pid) se ha detenido inesperadamente.${NC}"
        fi
    done
    sleep 5
done
