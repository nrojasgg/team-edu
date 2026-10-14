package middleware

import (
	"context"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
)

var ctx = context.Background()

func RedisCacheMiddleware(redisClient *redis.Client, ttl time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Si Redis no está disponible, continuar sin caché
		if redisClient == nil {
			return c.Next()
		}

		// Solo cacheamos peticiones GET
		if c.Method() != fiber.MethodGet {
			return c.Next()
		}

		// Usamos la URL completa como key (incluyendo query params)
		cacheKey := "cache:mscore:" + c.OriginalURL()

		// 1. Intentar obtener el dato de Redis
		cachedResponse, err := redisClient.Get(ctx, cacheKey).Bytes()
		if err == nil {
			// Cache Hit: Devolver la respuesta inmediatamente
			c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
			return c.Send(cachedResponse)
		}

		// 2. Cache Miss: Interceptar la respuesta original de Fiber
		err = c.Next()
		if err != nil {
			return err
		}

		// Solo cachear respuestas exitosas (200 OK)
		if c.Response().StatusCode() == fiber.StatusOK {
			body := append([]byte(nil), c.Response().Body()...) // Copiar body para evitar sobrescrituras

			// Guardar de forma asíncrona para no bloquear el retorno de esta request
			go func(key string, data []byte, duration time.Duration) {
				err := redisClient.Set(context.Background(), key, data, duration).Err()
				if err != nil {
					log.Printf("Error saving to Redis cache for key %s: %v", key, err)
				}
			}(cacheKey, body, ttl)
		}

		return nil
	}
}
