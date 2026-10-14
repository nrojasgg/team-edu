@echo off
echo ========================================
echo Iniciando todos los servicios de TeamEdu
echo en Windows Terminal con pestañas
echo ========================================
echo.

echo ----------------------------------------
echo Compilando plugins de MS-Core...
echo ----------------------------------------
echo.

cd /d "%~dp0ms-core"

for %%P in (enneagram mbti bfi44 peer_evaluation sociogram team_formation group_compatibility chronotype) do (
    echo Compilando plugin: %%P
    go build -o plugins_dist/%%P.exe ./plugins_src/%%P/
    if errorlevel 1 (
        echo [ERROR] Fallo al compilar el plugin: %%P
        pause
        exit /b 1
    )
    echo   OK: plugins_dist/%%P.exe
)

cd /d "%~dp0"

echo.
echo Plugins compilados correctamente.
echo ----------------------------------------
echo.

REM Iniciar Windows Terminal con múltiples pestañas
wt.exe ^
  --title "Frontend" -d "%~dp0front-eduteam" cmd /k "bun dev" ^
; new-tab --title "API Gateway" -d "%~dp0api-gateway" cmd /k "go run main.go" ^
; new-tab --title "MS Auth" -d "%~dp0ms-auth" cmd /k "go run main.go" ^
; new-tab --title "MS Courses" -d "%~dp0ms-courses" cmd /k "go run main.go" ^
; new-tab --title "MS Core" -d "%~dp0ms-core" cmd /k "go run main.go" ^
; new-tab --title "MS Notis" -d "%~dp0ms-notis" cmd /k "go run main.go"

echo.
echo ========================================
echo Todos los servicios han sido iniciados!
echo ========================================
echo.
echo Servicios en ejecucion en Windows Terminal:
echo - Frontend:        http://localhost:2999
echo - API Gateway:     http://localhost:3000
echo - MS-Auth:         http://localhost:3001
echo - MS-Courses:      http://localhost:3002
echo - MS-Core:         http://localhost:3003
echo - MS-Notis:        http://localhost:3004
echo.
echo Todos los servicios estan en pestañas separadas de Windows Terminal.
echo.
pause
