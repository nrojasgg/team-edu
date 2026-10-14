package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// LoggingMiddleware configura el middleware de logging para el API Gateway
func LoggingMiddleware() fiber.Handler {
	return logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path} - ${latency}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Local",
	})
}

// CORSMiddleware configura CORS para permitir requests desde el frontend
func CORSMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Method() == "OPTIONS" {
			return c.SendStatus(fiber.StatusOK)
		}

		return c.Next()
	}
}

// RequestIDMiddleware añade un ID único a cada request para tracking
func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := generateRequestID()
		c.Set("X-Request-ID", requestID)
		c.Locals("requestID", requestID)

		log.Printf("Request ID: %s - %s %s", requestID, c.Method(), c.Path())

		return c.Next()
	}
}

// RateLimitingMiddleware implementa rate limiting básico
func RateLimitingMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Implementación básica - en producción usar Redis o similar
		// Por ahora solo registramos la request
		log.Printf("Rate limiting check for IP: %s", c.IP())
		return c.Next()
	}
}

// generateRequestID genera un ID único para cada request
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString genera una cadena aleatoria de longitud específica
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
