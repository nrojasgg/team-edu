package routes

import (
	"ms-courses/internal/handlers"
	"ms-courses/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

type Handlers struct {
	Course         *handlers.CourseHandler
	Semester       *handlers.SemesterHandler
	CourseOffering *handlers.CourseOfferingHandler
	Section        *handlers.SectionHandler
	StudentSection *handlers.StudentSectionHandler
	InvitationCode *handlers.InvitationCodeHandler
}

func SetupRoutes(app *fiber.App, h *Handlers, jwtSecret string) {
	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "MS Courses API is running",
		})
	})

	// Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)

	// API version 1
	api := app.Group("/api/v1")

	// Todos los endpoints requieren autenticación
	authMW := middleware.AuthMiddleware(jwtSecret)
	adminMW := middleware.RoleMiddleware("admin")

	// ─────────────────────────────────────────
	// Courses
	// ─────────────────────────────────────────
	courses := api.Group("/courses", authMW)
	// Lectura: quienes ven todos los cursos o solo los suyos
	courses.Get("/", middleware.PermissionMiddleware("Ver todos los cursos", "Acceder a cursos vistos"), h.Course.GetAllCourses)
	courses.Get("/:id", middleware.PermissionMiddleware("Ver todos los cursos", "Acceder a cursos vistos"), h.Course.GetCourseByID)
	// Escritura
	courses.Post("/", middleware.PermissionMiddleware("Crear curso"), h.Course.CreateCourse)
	courses.Put("/:id", middleware.PermissionMiddleware("Editar curso"), h.Course.UpdateCourse)
	courses.Delete("/:id", adminMW, h.Course.DeleteCourse)

	// ─────────────────────────────────────────
	// Semesters
	// ─────────────────────────────────────────
	semesters := api.Group("/semesters", authMW)
	// Lectura: cualquier autenticado puede consultar semestres
	semesters.Get("/", h.Semester.GetAllSemesters)
	semesters.Get("/current", h.Semester.GetCurrentSemester)
	semesters.Get("/:id", h.Semester.GetSemesterByID)
	// Escritura: requiere permiso de edición de semestre
	semesters.Post("/", middleware.PermissionMiddleware("Editar semestre actual"), h.Semester.CreateSemester)
	semesters.Post("/create-or-activate", middleware.PermissionMiddleware("Editar semestre actual"), h.Semester.CreateOrActivateSemester)
	semesters.Put("/:id", middleware.PermissionMiddleware("Editar semestre actual"), h.Semester.UpdateSemester)
	semesters.Put("/:id/activate", middleware.PermissionMiddleware("Editar semestre actual"), h.Semester.SetActiveSemester)
	semesters.Delete("/:id", adminMW, h.Semester.DeleteSemester)

	// ─────────────────────────────────────────
	// Course Offerings
	// ─────────────────────────────────────────
	courseOfferings := api.Group("/course-offerings", authMW)
	courseOfferings.Get("/", middleware.PermissionMiddleware("Ver todos los cursos", "Acceder a cursos vistos"), h.CourseOffering.GetAllCourseOfferings)
	courseOfferings.Get("/:id", middleware.PermissionMiddleware("Ver todos los cursos", "Acceder a cursos vistos"), h.CourseOffering.GetCourseOfferingByID)
	courseOfferings.Post("/", middleware.PermissionMiddleware("Crear curso"), h.CourseOffering.CreateCourseOffering)
	courseOfferings.Delete("/:id", adminMW, h.CourseOffering.DeleteCourseOffering)

	// ─────────────────────────────────────────
	// Sections
	// ─────────────────────────────────────────
	sections := api.Group("/sections", authMW)
	// Lectura: quien ve todas las secciones o solo las propias
	sections.Get("/", middleware.PermissionMiddleware("Ver todas las secciones", "Ver secciones propias"), h.Section.GetAllSections)
	sections.Get("/:id", middleware.PermissionMiddleware("Ver todas las secciones", "Ver secciones propias"), h.Section.GetSectionByID)
	sections.Get("/:id/students", middleware.PermissionMiddleware("Ver todas las secciones", "Ver secciones propias"), h.StudentSection.GetStudentsBySectionID)
	sections.Get("/course-offering/:courseOfferingId", middleware.PermissionMiddleware("Ver todas las secciones", "Ver secciones propias"), h.Section.GetSectionsByCourseOffering)
	// Escritura
	sections.Post("/", middleware.PermissionMiddleware("Crear sección"), h.Section.CreateSection)
	sections.Put("/:id", middleware.PermissionMiddleware("Editar sección"), h.Section.UpdateSection)
	sections.Put("/:id/staff", middleware.PermissionMiddleware("Editar sección"), h.Section.UpdateSectionStaff)
	sections.Delete("/:id", adminMW, h.Section.DeleteSection)
	// Gestión de estudiantes en sección
	sections.Delete("/:sectionId/students/:studentId", middleware.PermissionMiddleware("Agregar estudiante a la sección"), h.StudentSection.RemoveStudentFromSectionByParams)

	// ─────────────────────────────────────────
	// Student Sections
	// ─────────────────────────────────────────
	studentSections := api.Group("/student-sections", authMW)
	studentSections.Post("/", middleware.PermissionMiddleware("Agregar estudiante a la sección"), h.StudentSection.AddStudentToSection)
	studentSections.Delete("/", middleware.PermissionMiddleware("Agregar estudiante a la sección"), h.StudentSection.RemoveStudentFromSection)

	// ─────────────────────────────────────────
	// Students
	// ─────────────────────────────────────────
	students := api.Group("/students", authMW)
	students.Get("/:id/sections", middleware.PermissionMiddleware("Ver secciones propias", "Ver todas las secciones"), h.StudentSection.GetSectionsByStudentID)

	// ─────────────────────────────────────────
	// Invitation Codes
	// ─────────────────────────────────────────
	invitationCodes := api.Group("/invitation-codes", authMW)
	invitationCodes.Post("/", middleware.PermissionMiddleware("Crear código de invitación"), h.InvitationCode.CreateInvitationCode)
	invitationCodes.Get("/section/:id", middleware.PermissionMiddleware("Ver todas las secciones", "Ver secciones propias"), h.InvitationCode.GetSectionInvitationCodes)
	invitationCodes.Delete("/:id", middleware.RoleMiddleware("admin", "coordinador", "profesor", "ayudante"), h.InvitationCode.DeactivateInvitationCode)
	invitationCodes.Post("/use", middleware.PermissionMiddleware("Usar código de invitación"), h.InvitationCode.UseInvitationCode)
	invitationCodes.Get("/:id/students", middleware.PermissionMiddleware("Ver todas las secciones", "Ver secciones propias"), h.InvitationCode.GetInvitationCodeUses)
}
