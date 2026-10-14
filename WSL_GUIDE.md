# Guía de Compilación para Linux y Pruebas en Windows Subsystem for Linux (WSL)

Esta guía explica los cambios de arquitectura realizados y cómo compilar y probar todos los microservicios de TeamEdu en **Linux / WSL**.

---

## 1. Resumen de Cambios de Arquitectura

1. **Gestor Dinámico de Plugins (`ms-core`)**:
   - `ms-core/internal/bootstrap/plugins.go` se actualizó para filtrar por sistema operativo:
     - En **Linux**: Carga únicamente binarios ejecutables ELF (sin extensión) e ignora archivos `.exe`.
     - En **Windows**: Carga únicamente binarios `.exe` e ignora binarios de Linux.
   - Ambos conjuntos de binarios pueden convivir de forma segura en `ms-core/plugins_dist/`.
   - Se incorporó el plugin `chronotype` en `ms-core/Makefile`, `ms-core/Dockerfile` y scripts de compilación.

2. **Makefiles Adaptados y Multiplataforma**:
   - `ms-core/Makefile`, `ms-courses/Makefile`, `ms-notis/Makefile` y `api-gateway/Makefile`:
     - Detectan el sistema operativo (`OS=Windows_NT` vs Linux/WSL).
     - Añadido el target `build-linux` (y `build-all-linux` en `ms-core`) con `CGO_ENABLED=0 GOOS=linux`.
     - Comandos de limpieza (`clean`) compatibles tanto con POSIX (`rm -rf`) como con Windows.

3. **Configuración de Red en Frontend (`front-eduteam`)**:
   - `front-eduteam/nuxt.config.ts`: `devServer.host` ahora utiliza `process.env.HOST || '0.0.0.0'` para que el reenvío de puertos entre WSL2 y el navegador en Windows funcione de forma fluida.

---

## 2. Requisitos Previos en WSL

Si vas a ejecutar y probar los servicios directamente en tu distribución de WSL (ej. Ubuntu):

1. **Go (>= 1.22) y utilidades de compilación**:
   ```bash
   sudo apt update
   sudo apt install -y golang-go unzip build-essential
   # O verificar versión:
   go version
   ```

2. **Bun e instalación de dependencias de frontend**:
   ```bash
   curl -fsSL https://bun.sh/install | bash
   source ~/.bashrc
   bun --version

   # Instalar dependencias en front-eduteam para Linux:
   cd /mnt/c/Proyectos/RANDOM/USACH/TESIS/teamedu-main/front-eduteam
   bun install
   ```

3. **Bases de datos (PostgreSQL, MongoDB, Redis)**:
   Puedes levantarlas usando Docker (Docker Desktop con soporte WSL2 o Docker en WSL):
   ```bash
   docker compose up -d auth-db courses-db mongo-db redis-cache
   ```

---

## 3. Formas de Compilar para Linux

### Opción A: Desde Windows (Cross-compilación Go)
Go permite compilar para Linux de forma nativa sin herramientas adicionales:
```cmd
build-linux.bat
```
Esto compilará los 8 plugins (`enneagram`, `mbti`, `bfi44`, `peer_evaluation`, `sociogram`, `team_formation`, `group_compatibility`, `chronotype`) como binarios ELF de Linux en `ms-core/plugins_dist/`.

### Opción B: Dentro de WSL / Linux
```bash
chmod +x build-linux.sh start.sh
./build-linux.sh
```

O usando Make en `ms-core`:
```bash
cd ms-core
make build-all-linux
```

---

## 4. Cómo Iniciar y Probar en WSL

### Método 1: Desde Windows usando Windows Terminal con pestañas de WSL
Ejecuta el script:
```cmd
start-wsl.bat
```
Este script:
1. Compila automáticamente los plugins para Linux.
2. Abre Windows Terminal con 6 pestañas independientes, cada una ejecutando su servicio dentro de WSL:
   - Frontend (`bun dev`)
   - API Gateway (`go run main.go`)
   - MS-Auth (`go run main.go`)
   - MS-Courses (`go run main.go`)
   - MS-Core (`go run main.go`)
   - MS-Notis (`go run main.go`)

### Método 2: Directamente dentro de la terminal de WSL
Abre tu terminal de WSL, dirígete al proyecto:
```bash
cd /mnt/c/Proyectos/RANDOM/USACH/TESIS/teamedu-main
./start.sh
```
Este script:
- Verifica que Go y Bun estén instalados.
- Compila los plugins de Linux.
- Inicia los 6 servicios en segundo plano registrando logs en `logs/<servicio>.log`.
- Con `Ctrl+C` detiene de forma limpia todos los procesos sin dejar puertos tomados.

---

## 5. URLs de los Servicios

Una vez iniciados, puedes acceder desde tu navegador en Windows o mediante curl en WSL:

| Servicio | Puerto | URL |
|---|---|---|
| Frontend | 2999 | http://127.0.0.1:2999 |
| API Gateway | 3000 | http://127.0.0.1:3000 |
| MS Auth | 3001 | http://127.0.0.1:3001 |
| MS Courses | 3002 | http://127.0.0.1:3002 |
| MS Core | 3003 | http://127.0.0.1:3003 |
| MS Notis | 3004 | http://127.0.0.1:3004 |
| Healthcheck Core | 3003 | http://127.0.0.1:3003/health |

> **Nota de rendimiento en WSL2**: Accede siempre mediante `http://127.0.0.1:2999` en tu navegador en lugar de `localhost`. En Windows 10/11, `localhost` intenta resolver primero vía IPv6 (`::1`), sufriendo un retraso de 2.2 segundos por petición antes de reintentar por IPv4. Con `127.0.0.1` la respuesta es inmediata (<5ms).

---

## 6. Solución de Problemas Frecuentes

### Error: `/bin/sh: bash: not found (código 127)`
- **Causa**: La única distribución registrada en WSL actualmente es `docker-desktop`. Esta es una imagen interna y mínima creada por Docker Desktop que no incluye `bash` (solo `/bin/sh` básico) ni utilidades de desarrollo (Go, Bun, apt).
- **Solución**:
  1. Instala Ubuntu en WSL desde PowerShell o CMD:
     ```cmd
     wsl --install -d Ubuntu
     ```
  2. Cuando se complete la descarga, se abrirá la consola de Ubuntu para configurar tu usuario y contraseña.
  3. Dentro de Ubuntu, instala Go y Bun:
     ```bash
     sudo apt update && sudo apt install -y golang-go
     curl -fsSL https://bun.sh/install | bash
     ```
  4. Vuelve a ejecutar `start-wsl.bat`. El script detectará Ubuntu automáticamente y abrirá cada pestaña dirigiéndose a dicha distribución.

### Error: `Cannot find native binding ... @oxc-parser/binding-linux-x64-gnu`
- **Causa**: Al compartir `node_modules` entre Windows y WSL, las librerías nativas compiladas (como `@oxc-parser`, `@esbuild`, `@rollup`, `lightningcss`, `@parcel/watcher`) solo tenían descargados los binarios de Windows (`*-win32-x64*`).
- **Solución**: En la terminal de Ubuntu ejecuta:
  ```bash
  cd /mnt/c/Proyectos/RANDOM/USACH/TESIS/teamedu-main/front-eduteam
  bun install --force
  ```
  O descarga directamente las dependencias nativas para Linux:
  ```bash
  bun add -d @oxc-parser/binding-linux-x64-gnu @rollup/rollup-linux-x64-gnu @esbuild/linux-x64 lightningcss-linux-x64-gnu @parcel/watcher-linux-x64-glibc
  ```
