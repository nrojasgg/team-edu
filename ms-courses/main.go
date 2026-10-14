package main

import (
	"log"
	"ms-courses/internal/config"
	"ms-courses/internal/database"
	"ms-courses/internal/handlers"
	"ms-courses/internal/repositories"
	"ms-courses/internal/routes"
	"ms-courses/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// @title MS Courses API
// @version 1.0
// @description API para gestión de cursos y secciones del sistema TeamEdu
// @host localhost:8081
// @BasePath /api/v1
func main() {
	// Cargar configuración
	cfg := config.Load()

	// Conectar a la base de datos
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Inicializar repositorios
	courseRepo := repositories.NewCourseRepository(db)
	semesterRepo := repositories.NewSemesterRepository(db)
	courseOfferingRepo := repositories.NewCourseOfferingRepository(db)
	sectionRepo := repositories.NewSectionRepository(db)
	studentSectionRepo := repositories.NewStudentSectionRepository(db)
	invitationCodeRepo := repositories.NewInvitationCodeRepository(db)

	// Inicializar servicios
	courseService := services.NewCourseService(courseRepo)
	semesterService := services.NewSemesterService(semesterRepo)
	courseOfferingService := services.NewCourseOfferingService(courseOfferingRepo, courseRepo, semesterRepo)
	sectionService := services.NewSectionService(sectionRepo, courseOfferingRepo)
	studentSectionService := services.NewStudentSectionService(studentSectionRepo)

	// Inicializar handlers
	courseHandler := handlers.NewCourseHandler(courseService)
	semesterHandler := handlers.NewSemesterHandler(semesterService)
	courseOfferingHandler := handlers.NewCourseOfferingHandler(courseOfferingService)
	sectionHandler := handlers.NewSectionHandler(sectionService)
	studentSectionHandler := handlers.NewStudentSectionHandler(studentSectionService)
	invitationCodeHandler := handlers.NewInvitationCodeHandler(invitationCodeRepo)

	// Crear aplicación Fiber
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middlewares
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Configurar rutas
	h := &routes.Handlers{
		Course:         courseHandler,
		Semester:       semesterHandler,
		CourseOffering: courseOfferingHandler,
		Section:        sectionHandler,
		StudentSection: studentSectionHandler,
		InvitationCode: invitationCodeHandler,
	}

	routes.SetupRoutes(app, h, cfg.JWTSecret)

	// Iniciar servidor
	log.Printf("MS Courses starting on port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
