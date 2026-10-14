#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "========================================================"
echo "  TeamEdu - Compilación de Plugins para Linux"
echo "========================================================"
echo ""

cd "$SCRIPT_DIR/ms-core"
mkdir -p plugins_dist

export CGO_ENABLED=0
export GOOS=linux

PLUGINS=(
    enneagram
    mbti
    bfi44
    peer_evaluation
    sociogram
    team_formation
    group_compatibility
    chronotype
)

for p in "${PLUGINS[@]}"; do
    echo "[Compilando plugin Linux] $p ..."
    go build -ldflags="-s -w" -o "plugins_dist/$p" "./plugins_src/$p/"
    chmod +x "plugins_dist/$p"
    echo "  -> OK: plugins_dist/$p (binario ejecutable Linux)"
done

cd "$SCRIPT_DIR"

echo ""
echo "========================================================"
echo "  Todos los plugins para Linux se compilaron con éxito!"
echo "========================================================"
echo ""
