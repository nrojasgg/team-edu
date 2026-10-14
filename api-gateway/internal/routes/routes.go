package routes

import (
	"api-gateway/internal/middleware"
	"api-gateway/internal/proxy"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configura todas las rutas del API Gateway
func SetupRoutes(app *fiber.App, proxyService *proxy.ProxyService) {
	// Aplicar middlewares globales
	app.Use(middleware.LoggingMiddleware())
	app.Use(middleware.CORSMiddleware())
	app.Use(middleware.RequestIDMiddleware())
	app.Use(middleware.RateLimitingMiddleware())

	// Grupo de rutas para el API v1
	api := app.Group("/api/v1")

	// Ruta especial para obtener toda la información de cursos en una sola llamada
	api.Get("/courses-complete", proxyService.GetCoursesComplete)

	// Ruta especial para obtener respuestas de cuestionarios con información de estudiantes
	api.Get("/core/enneagram/questionnaires/:id/responses", proxyService.GetQuestionnaireResponses)

	// Ruta especial para obtener cuestionarios con formato simplificado (admin)
	api.Get("/core/enneagram/questionnaires", proxyService.GetQuestionnaires)

	// Rutas especiales para BFI-44
	api.Get("/core/bfi44/questionnaires/:id/responses", proxyService.GetBFI44QuestionnaireResponses)
	api.Get("/core/bfi44/questionnaires", proxyService.GetBFI44Questionnaires)

	// Rutas especiales para MBTI (Francis Psychological Type Scales)
	api.Get("/core/mbti/questionnaires/:id/responses", proxyService.GetMBTIQuestionnaireResponses)
	api.Get("/core/mbti/questionnaires", proxyService.GetMBTIQuestionnaires)

	// Ruta para verificar si un estudiante ya respondió un test específico
	api.Get("/core/plugin/:pluginName/tests/:testId/responses/student/:studentId", proxyService.GetStudentTestResponse)

	// Ruta para exportar evaluación de pares a Excel
	api.Get("/surveys/:surveyId/export/excel", proxyService.ExportPeerEvaluationExcel)

	// Ruta especial para obtener el grupo del estudiante con información completa
	api.Get("/core/groups/section/:sectionId/my-group-complete", proxyService.GetMyGroupComplete)

	// Ruta especial para obtener resumen de respuestas agregadas por estudiante y grupo
	api.Get("/sections/:section_id/responses/summary", proxyService.GetSectionResponsesSummary)

	// Ruta especial para obtener respuestas de evaluación de pares por sección y grupo
	api.Get("/sections/:section_id/responses/:group_id", proxyService.GetSectionGroupResponses)

	// Ruta especial para obtener todas las respuestas de evaluación de pares por sección
	api.Get("/sections/:section_id/responses", proxyService.GetSectionResponses)

	// Ruta explícita para actualizar scores de grupos (PATCH necesita registro explícito)
	api.Patch("/core/groups/section/:sectionId/scores", proxyService.ProxyToCore)

	// Rutas para ms-core (debe ir DESPUÉS de las rutas específicas)
	core := app.Group("/api/v1/core")
	core.All("/*", proxyService.ProxyToCore)

	// Rutas para ms-notis (notificaciones)
	notifications := api.Group("/notifications")
	notifications.All("/*", proxyService.ProxyToNotifications)

	// Health check del API Gateway
	app.Get("/health", func(c *fiber.Ctx) error {
		services := fiber.Map{
			"api-gateway": "ok",
		}
		status := "ok"
		message := "API Gateway is running and all services are healthy"

		// Verificar el estado del microservicio de registro curricular
		if err := proxyService.HealthCheck(); err != nil {
			services["registro-curricular"] = "unavailable"
			services["courses"] = "unavailable"
			status = "degraded"
			message = "API Gateway is running but some services are unavailable"
		} else {
			services["registro-curricular"] = "ok"
			services["courses"] = "ok"
		}

		// Verificar ms-core si está configurado
		if proxyService.CoreURL != "" {
			// Aquí se podría agregar verificación específica de ms-core
			services["ms-core"] = "ok"
		}

		var statusCode int
		if status == "ok" {
			statusCode = fiber.StatusOK
		} else {
			statusCode = fiber.StatusServiceUnavailable
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"status":   status,
			"message":  message,
			"services": services,
		})
	})

	// Rutas de información del gateway
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"name":        "TeamEdu API Gateway",
			"version":     "1.0.0",
			"description": "API Gateway for TeamEdu microservices",
			"endpoints": fiber.Map{
				"health":              "/health",
				"auth":                "/api/v1/auth/*",
				"users":               "/api/v1/users/*",
				"roles":               "/api/v1/roles/*",
				"permissions":         "/api/v1/permissions/*",
				"courses":             "/api/v1/courses/*",
				"semesters":           "/api/v1/semesters/*",
				"course-offerings":    "/api/v1/course-offerings/*",
				"sections":            "/api/v1/sections/*",
				"student-sections":    "/api/v1/student-sections/*",
				"students":            "/api/v1/students/*",
				"invitation-codes":    "/api/v1/invitation-codes/*",
				"questionnaires":      "/api/v1/core/enneagram/*",
				"surveys":             "/api/v1/core/surveys/*",
				"teams":               "/api/v1/core/teams/*",
				"plugins":             "/api/v1/core/plugins",
				"notifications":       "/api/v1/notifications/*",
				"registro-curricular": "/api/v1/registro/*",
			},
		})
	})

	// Rutas de autenticación del microservicio de registro curricular
	// Todas las rutas que empiecen con /api/v1/auth/* serán redirigidas
	auth := api.Group("/auth")
	auth.All("/*", proxyService.ProxyToRegistroCurricular)

	// Rutas de usuarios del microservicio de registro curricular
	// Todas las rutas que empiecen con /api/v1/users/* serán redirigidas
	users := api.Group("/users")
	users.All("/*", proxyService.ProxyToRegistroCurricular)

	// Rutas de roles del microservicio de registro curricular
	// Todas las rutas que empiecen con /api/v1/roles/* serán redirigidas
	roles := api.Group("/roles")
	roles.All("/*", proxyService.ProxyToRegistroCurricular)

	// Rutas de permisos del microservicio de registro curricular
	// Todas las rutas que empiecen con /api/v1/permissions/* serán redirigidas
	permissions := api.Group("/permissions")
	permissions.All("/*", proxyService.ProxyToRegistroCurricular)

	// Rutas de cursos del microservicio de cursos
	// Todas las rutas que empiecen con /api/v1/courses/* serán redirigidas
	courses := api.Group("/courses")
	courses.All("/*", proxyService.ProxyToCourses)

	// Rutas de semestres del microservicio de cursos
	semesters := api.Group("/semesters")
	semesters.All("/*", proxyService.ProxyToCourses)

	// Rutas de ofertas de cursos del microservicio de cursos
	courseOfferings := api.Group("/course-offerings")
	courseOfferings.All("/*", proxyService.ProxyToCourses)

	// Rutas de secciones del microservicio de cursos
	sections := api.Group("/sections")
	// Ruta especial para obtener estudiantes con toda su información en una sola llamada
	sections.Get("/:id/students-complete", proxyService.GetSectionStudentsComplete)
	// Ruta especial para obtener información completa de una sección en una sola llamada
	sections.Get("/:id/complete", proxyService.GetSectionComplete)
	// Ruta especial para obtener todos los datos necesarios para la sección de grupos (consolidado en ms-core)
	sections.Get("/:section_id/complete-groups-data", proxyService.GetSectionGroupsData)
	sections.Get("/:section_id/group-data", proxyService.GetSectionGroupData)
	sections.Get("/:section_id/survey-data", proxyService.GetSectionSurveyData)
	sections.All("/*", proxyService.ProxyToCourses)

	// Rutas de student-sections del microservicio de cursos
	studentSections := api.Group("/student-sections")
	studentSections.All("/*", proxyService.ProxyToCourses)

	// Rutas de invitation-codes del microservicio de cursos
	invitationCodes := api.Group("/invitation-codes")
	invitationCodes.All("/*", proxyService.ProxyToCourses)

	// Rutas de estudiantes del microservicio de cursos
	students := api.Group("/students")
	students.All("/*", proxyService.ProxyToCourses)

	// Rutas del microservicio de registro curricular (legacy)
	// Todas las rutas que empiecen con /api/v1/registro/* serán redirigidas
	registro := api.Group("/registro")
	registro.All("/*", proxyService.ProxyToRegistroCurricular)

	// Ruta catch-all para rutas no encontradas
	app.Use("*", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Route not found",
			"message": "The requested endpoint does not exist",
			"available_services": []string{
				"/api/v1/auth/*",
				"/api/v1/users/*",
				"/api/v1/roles/*",
				"/api/v1/permissions/*",
				"/api/v1/courses/*",
				"/api/v1/semesters/*",
				"/api/v1/course-offerings/*",
				"/api/v1/sections/*",
				"/api/v1/student-sections/*",
				"/api/v1/students/*",
				"/api/v1/invitation-codes/*",
				"/api/v1/core/*",
				"/api/v1/registro/*",
			},
		})
	})
}
