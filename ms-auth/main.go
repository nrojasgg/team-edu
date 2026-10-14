// Package main is the entry point of the application
package main

import (
	"log"
	"ms-registro-curricular/internal/config"
	"ms-registro-curricular/internal/database"
	"ms-registro-curricular/internal/handlers"
	"ms-registro-curricular/internal/middleware"
	"ms-registro-curricular/internal/repositories"
	"ms-registro-curricular/internal/routes"
	"ms-registro-curricular/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"

	_ "ms-registro-curricular/docs"
)

// @title Registro Curricular API
// @version 1.0
// @description API for curriculum registration system
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@registrocurricular.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:3000
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	roleRepo := repositories.NewRoleRepository(db)
	permissionRepo := repositories.NewPermissionRepository(db)

	// Initialize services
	authService := services.NewAuthService(
		userRepo,
		roleRepo,
		cfg.JWTSecret,
		cfg.GoogleClientID,
		cfg.GoogleClientSecret,
		cfg.GoogleRedirectURL,
		cfg.ApiGatewayURL,
	)
	userService := services.NewUserService(userRepo, cfg.JWTSecret)
	roleService := services.NewRoleService(roleRepo)
	permissionService := services.NewPermissionService(permissionRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService, roleService)
	roleHandler := handlers.NewRoleHandler(roleService)
	permissionHandler := handlers.NewPermissionHandler(permissionService)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
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
	routes.SetupRoutes(app, &routes.Handlers{
		Auth:       authHandler,
		User:       userHandler,
		Role:       roleHandler,
		Permission: permissionHandler,
	}, cfg.JWTSecret)

	// Setup Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)  // default
	app.Get("/docs/*", swagger.New(swagger.Config{ // custom config
		URL:         "http://localhost:" + cfg.Port + "/swagger/doc.json",
		DeepLinking: false,
		Title:       "Registro Curricular API Documentation",
	}))

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	log.Printf("Swagger UI available at: http://localhost:%s/swagger/", cfg.Port)
	log.Printf("Alternative docs at: http://localhost:%s/docs/", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
