package routes

import (
	"ms-notis/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, notificationHandler *handlers.NotificationHandler) {
	// API v1 routes
	api := app.Group("/api/v1")

	// Notification routes
	notifications := api.Group("/notifications")
	notifications.Post("/password-reset", notificationHandler.SendPasswordReset)
	notifications.Post("/test-reminder", notificationHandler.SendTestReminder)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "ms-notis",
			"message": "Notification service is running",
		})
	})

	// Root endpoint
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"name":        "TeamEdu Notification Service",
			"version":     "1.0.0",
			"description": "Microservice for sending email notifications",
			"endpoints": fiber.Map{
				"health":         "/health",
				"password-reset": "/api/v1/notifications/password-reset",
				"test-reminder":  "/api/v1/notifications/test-reminder",
			},
		})
	})
}
