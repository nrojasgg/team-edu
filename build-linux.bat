@echo off
setlocal enabledelayedexpansion

echo ========================================================
echo   TeamEdu - Compilacion de Plugins para Linux (x86_64)
echo ========================================================
echo.

set "SCRIPT_DIR=%~dp0"
cd /d "%SCRIPT_DIR%ms-core"

REM Configurar variables para cross-compilacion Go a Linux
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0

if not exist "plugins_dist" mkdir "plugins_dist"

echo Compilando plugins para Linux en: ms-core\plugins_dist\
echo.

set PLUGINS=enneagram mbti bfi44 peer_evaluation sociogram team_formation group_compatibility chronotype

for %%P in (%PLUGINS%) do (
    echo [Compilando plugin Linux] %%P ...
    go build -ldflags="-s -w" -o plugins_dist/%%P ./plugins_src/%%P/
    if errorlevel 1 (
        echo [ERROR] Fallo al compilar el plugin %%P para Linux.
        pause
        exit /b 1
    )
    echo   -> OK: plugins_dist/%%P (binario ELF Linux)
)

cd /d "%SCRIPT_DIR%"

echo.
echo ========================================================
echo   Plugins compilados exitosamente para Linux!
echo ========================================================
echo.
