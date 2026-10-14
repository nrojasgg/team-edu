package middleware

import (
	"errors"
	"fmt"
	"ms-registro-curricular/internal/models"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

// AuthMiddleware validates JWT tokens and carga el usuario en c.Locals("user")
func AuthMiddleware(jwtSecret string, userService interface {
	GetByID(id int64) (*models.User, error)
}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Authorization header is required",
			})
		}

		// Extract token from Bearer scheme
		tokenString := extractBearerToken(authHeader)
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid authorization header format",
			})
		}

		// Decodificar y validar el JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid token",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid token claims",
			})
		}

		// Extraer el id de usuario del claim (acepta float64 o string)
		var userID int64
		switch v := claims["id"].(type) {
		case float64:
			userID = int64(v)
		case string:
			// Intentar convertir string a int64
			var err error
			userID, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"success": false,
					"message": "User ID in token is not a valid int64",
				})
			}
		default:
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "User ID not found in token",
			})
		}

		// Cargar el usuario desde la base de datos usando el servicio inyectado
		user, userErr := userService.GetByID(userID)
		if userErr != nil || user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "User not found",
			})
		}
		c.Locals("user", user)
		c.Locals("token", tokenString)

		// Extraer permisos del token JWT si existen
		if permissions, exists := claims["permissions"]; exists {
			if perms, ok := permissions.([]interface{}); ok {
				var permStrings []string
				for _, p := range perms {
					if permStr, ok := p.(string); ok {
						permStrings = append(permStrings, permStr)
					}
				}
				c.Locals("userPermissions", permStrings)
			}
		}

		return c.Next()
	}
}

func extractBearerToken(authHeader string) string {
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return parts[1]
}

// ErrorHandler handles application errors
func ErrorHandler(c *fiber.Ctx, err error) error {
	// Default error response
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	// Handle Fiber errors
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	// Handle custom errors
	switch err {
	case errors.New("user not found"):
		code = fiber.StatusNotFound
		message = "User not found"
	case errors.New("Correo o contraseña incorrectos"):
		code = fiber.StatusUnauthorized
		message = "Correo o contraseña incorrectos"
	case errors.New("user with this email already exists"):
		code = fiber.StatusConflict
		message = "User with this email already exists"
	}

	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"message": message,
		"error":   err.Error(),
	})
}

// PermissionMiddleware verifica que el usuario tenga al menos uno de los permisos requeridos
func PermissionMiddleware(requiredPermissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userPermissions := c.Locals("userPermissions")
		if userPermissions == nil {
			return c.Status(403).JSON(fiber.Map{
				"error":   true,
				"message": "Access denied: no permissions found",
			})
		}

		// Convertir a slice de strings
		var permissions []string
		if perms, ok := userPermissions.([]interface{}); ok {
			for _, p := range perms {
				if permStr, ok := p.(string); ok {
					permissions = append(permissions, permStr)
				}
			}
		} else if perms, ok := userPermissions.([]string); ok {
			permissions = perms
		}

		// Verificar si el usuario tiene al menos uno de los permisos requeridos
		for _, requiredPerm := range requiredPermissions {
			for _, userPerm := range permissions {
				if userPerm == requiredPerm {
					return c.Next()
				}
			}
		}

		return c.Status(403).JSON(fiber.Map{
			"error":   true,
			"message": fmt.Sprintf("Access denied: requires one of %v", requiredPermissions),
		})
	}
}

// RoleMiddleware verifica que el usuario autenticado tenga alguno de los roles requeridos.
// Depende de que AuthMiddleware haya cargado el usuario en c.Locals("user").
func RoleMiddleware(requiredRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(*models.User)
		if !ok || user == nil || user.Role == nil {
			return c.Status(403).JSON(fiber.Map{
				"error":   true,
				"message": "Access denied: no role found",
			})
		}

		for _, role := range requiredRoles {
			if user.Role.Name == role {
				return c.Next()
			}
		}

		return c.Status(403).JSON(fiber.Map{
			"error":   true,
			"message": fmt.Sprintf("Access denied: requires role %v", requiredRoles),
		})
	}
}
