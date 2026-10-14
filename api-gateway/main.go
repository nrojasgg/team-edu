package main

import (
	"api-gateway/internal/config"
	"api-gateway/internal/proxy"
	"api-gateway/internal/routes"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Cargar configuración
	cfg := config.LoadConfig()

	// Crear instancia de Fiber
	app := fiber.New(fiber.Config{
		AppName: "TeamEdu API Gateway v1.0.0",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			log.Printf("Error: %v", err)

			return c.Status(code).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		},
	})

	// Middleware de recuperación ante panics
	app.Use(recover.New())

	// Crear servicio de proxy
	proxyService := proxy.NewProxyServiceComplete(cfg.RegistroCurricularURL, cfg.CoursesURL, cfg.CoreURL, cfg.NotificationsURL)

	// Configurar rutas
	routes.SetupRoutes(app, proxyService)

	// Iniciar servidor
	log.Printf("🚀 API Gateway starting on port %s", cfg.GatewayPort)
	log.Printf("📡 Proxying to Registro Curricular at: %s", cfg.RegistroCurricularURL)
	log.Printf("📡 Proxying to Courses at: %s", cfg.CoursesURL)
	log.Printf("📡 Proxying to Core at: %s", cfg.CoreURL)

	if err := app.Listen(":" + cfg.GatewayPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
