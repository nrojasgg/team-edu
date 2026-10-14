package main

import (
	"log"
	"ms-notis/internal/config"
	"ms-notis/internal/handlers"
	"ms-notis/internal/routes"
	"ms-notis/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize services
	emailService := services.NewEmailService(cfg)

	// Initialize handlers
	notificationHandler := handlers.NewNotificationHandler(emailService)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName: "TeamEdu Notification Service v1.0.0",
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

	// Global middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigin + ",http://127.0.0.1:2999",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))
	app.Use(logger.New())
	app.Use(recover.New())

	// Setup routes
	routes.SetupRoutes(app, notificationHandler)

	// Start server
	log.Printf("🚀 Notification Service starting on port %s", cfg.Port)
	log.Printf("📧 SMTP configured: %s:%s", cfg.SMTPHost, cfg.SMTPPort)
	log.Fatal(app.Listen(":" + cfg.Port))
}
