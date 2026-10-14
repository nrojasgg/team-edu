package routes

import (
	"ms-registro-curricular/internal/handlers"
	"ms-registro-curricular/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

// Aliases de permisos para legibilidad en las rutas
var (
	canCreateUser = []string{"Crear estudiante", "Crear profesor", "Crear coordinador", "Crear administrador"}
	canEditUser   = []string{"Editar estudiante", "Editar profesor", "Editar coordinador", "Editar administrador"}
)

type Handlers struct {
	Auth       *handlers.AuthHandler
	User       *handlers.UserHandler
	Role       *handlers.RoleHandler
	Permission *handlers.PermissionHandler
}

func SetupRoutes(app *fiber.App, h *Handlers, jwtSecret string) {
	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "MS Registro Curricular API is running",
		})
	})

	// Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)

	// API version 1
	api := app.Group("/api/v1")

	// Auth routes (no middleware required)
	auth := api.Group("/auth")
	auth.Post("/login", h.Auth.Login)                    // Deprecated
	auth.Post("/register", h.Auth.Register)              // Deprecated
	auth.Get("/google", h.Auth.GoogleAuth)               // Get Google OAuth URL
	auth.Post("/google/callback", h.Auth.GoogleCallback) // Google OAuth callback
	auth.Post("/forgot-password", h.Auth.ForgotPassword) // Send reset link
	auth.Post("/reset-password", h.Auth.ResetPassword)   // Change password

	// Protected groups
	authMW := middleware.AuthMiddleware(jwtSecret, h.User)
	adminMW := middleware.RoleMiddleware("admin")

	// User routes
	users := api.Group("/users", authMW)
	users.Get("/", h.User.ListUsers)
	users.Post("/", middleware.PermissionMiddleware(canCreateUser...), h.User.CreateUser)
	users.Get("/by-email", h.User.GetUserByEmail)
	users.Get("/by-rut", h.User.GetUserByRUT)
	users.Get(":id", h.User.GetUser)
	users.Put(":id", middleware.PermissionMiddleware(canEditUser...), h.User.UpdateUser)
	users.Delete(":id", middleware.PermissionMiddleware("Eliminar usuario"), h.User.DeleteUser)
	users.Post("/batch", h.User.GetUsersBatch)
	users.Post(":id/impersonate", adminMW, h.User.ImpersonateUser)
	users.Get(":id/permissions", h.User.GetUserPermissions)

	// Role routes — lectura: cualquier autenticado; escritura: solo admin
	roles := api.Group("/roles", authMW)
	roles.Get("/", h.Role.GetRoles)
	roles.Get("/:id", h.Role.GetRole)
	roles.Post("/", adminMW, h.Role.CreateRole)
	roles.Put("/:id", adminMW, h.Role.UpdateRole)
	roles.Delete("/:id", adminMW, h.Role.DeleteRole)

	// Permission routes — lectura: cualquier autenticado; escritura: solo admin
	permissions := api.Group("/permissions", authMW)
	permissions.Get("/", h.Permission.GetPermissions)
	permissions.Get("/:id", h.Permission.GetPermission)
	permissions.Post("/", adminMW, h.Permission.CreatePermission)
	permissions.Put("/:id", adminMW, h.Permission.UpdatePermission)
	permissions.Delete("/:id", adminMW, h.Permission.DeletePermission)

	// Role-permission management — solo admin
	roles.Get("/:id/permissions", h.Permission.GetRolePermissions)
	roles.Post("/:id/permissions", adminMW, h.Permission.AssignPermissionToRole)
	roles.Delete("/:roleId/permissions/:permissionId", adminMW, h.Permission.RemovePermissionFromRole)
}
