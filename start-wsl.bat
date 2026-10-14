@echo off
setlocal

echo ========================================================
echo   Iniciando todos los servicios de TeamEdu en WSL
echo   (Windows Subsystem for Linux + Windows Terminal)
echo ========================================================
echo.

REM 1. Configurar distribucion de WSL (Ubuntu)
set "WSL_DISTRO=Ubuntu"

REM Verificar que Ubuntu responde en WSL
wsl.exe -d %WSL_DISTRO% echo test >nul 2>&1
if errorlevel 1 (
    echo [ERROR] No se pudo conectar a la distribucion %WSL_DISTRO% en WSL.
    echo Asegurate de que Ubuntu este instalado ejecutando:
    echo.
    echo     wsl --install -d Ubuntu
    echo.
    pause
    exit /b 1
)

echo [WSL] Distribucion verificada: %WSL_DISTRO%
echo.

REM 2. Compilar plugins para Linux antes de iniciar servicios
echo --------------------------------------------------------
echo Compilando plugins de MS-Core para Linux...
echo --------------------------------------------------------
call "%~dp0build-linux.bat"
if errorlevel 1 (
    echo [ERROR] Fallo la compilacion de los plugins para Linux.
    pause
    exit /b 1
)

echo.
echo --------------------------------------------------------
echo Abriendo pestanas de WSL (%WSL_DISTRO%) en Windows Terminal...
echo --------------------------------------------------------
echo.

REM Iniciar Windows Terminal con pestanas ejecutandose en WSL
wt.exe ^
  --title "Frontend (WSL)" wsl.exe -d %WSL_DISTRO% --cd "%~dp0front-eduteam" bash -lic "bunx nuxt dev --no-fork" ^
; new-tab --title "API Gateway (WSL)" wsl.exe -d %WSL_DISTRO% --cd "%~dp0api-gateway" bash -lic "CGO_ENABLED=0 go run main.go" ^
; new-tab --title "MS Auth (WSL)" wsl.exe -d %WSL_DISTRO% --cd "%~dp0ms-auth" bash -lic "CGO_ENABLED=0 go run main.go" ^
; new-tab --title "MS Courses (WSL)" wsl.exe -d %WSL_DISTRO% --cd "%~dp0ms-courses" bash -lic "CGO_ENABLED=0 go run main.go" ^
; new-tab --title "MS Core (WSL)" wsl.exe -d %WSL_DISTRO% --cd "%~dp0ms-core" bash -lic "CGO_ENABLED=0 go run main.go" ^
; new-tab --title "MS Notis (WSL)" wsl.exe -d %WSL_DISTRO% --cd "%~dp0ms-notis" bash -lic "CGO_ENABLED=0 go run main.go"

echo.
echo ========================================================
echo   Todos los servicios han sido iniciados en WSL!
echo ========================================================
echo.
echo Servicios en ejecucion en Windows Terminal (WSL - %WSL_DISTRO%):
echo - Frontend:        http://127.0.0.1:2999
echo - API Gateway:     http://127.0.0.1:3000
echo - MS-Auth:         http://127.0.0.1:3001
echo - MS-Courses:      http://127.0.0.1:3002
echo - MS-Core:         http://127.0.0.1:3003
echo - MS-Notis:        http://127.0.0.1:3004
echo.
echo Cada servicio esta en una pestana separada de WSL.
echo.
pause
