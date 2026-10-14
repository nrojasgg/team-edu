# ms-notis

Microservicio de notificaciones para TeamEdu.

## Configuración

1. Copia `.env` y actualiza las credenciales SMTP:
```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=tu-email@gmail.com
SMTP_PASSWORD=tu-contraseña-de-aplicación
```

Para Gmail, necesitas generar una "Contraseña de aplicación" en tu cuenta de Google.

2. Instala dependencias:
```bash
go mod download
```

3. Ejecuta el servicio:
```bash
go run main.go
```

## Endpoints

### POST /api/v1/notifications/password-reset
Envía un correo de restablecimiento de contraseña.

**Body:**
```json
{
  "email": "student@usach.cl",
  "name": "Juan Pérez",
  "reset_token": "abc123",
  "reset_url": "http://localhost:2999/reset-password"
}
```

### POST /api/v1/notifications/test-reminder
Envía un recordatorio de cuestionario pendiente.

**Body:**
```json
{
  "email": "student@usach.cl",
  "name": "Juan Pérez",
  "test_name": "Cuestionario de Eneagrama",
  "test_url": "http://localhost:2999/enneagram/test/123",
  "due_date": "2026-01-20"
}
```

### GET /health
Health check del servicio.

## Desarrollo

```bash
make run    # Ejecutar en modo desarrollo
make build  # Compilar binario
make clean  # Limpiar archivos compilados
```
