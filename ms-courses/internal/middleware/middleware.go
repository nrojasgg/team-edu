package middleware

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware verifica el token JWT y carga user_id, email, role y permissions en el contexto
func AuthMiddleware(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		// Verificar formato Bearer
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		tokenString := parts[1]

		// Verificar token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("invalid signing method")
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Extraer claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			var uid interface{}
			if v, exists := claims["id"]; exists {
				uid = v
			} else if v, exists := claims["user_id"]; exists {
				uid = v
			}

			c.Locals("user_id", uid)
			c.Locals("email", claims["email"])
			c.Locals("role", claims["role"])

			// Extraer permisos del token
			if perms, exists := claims["permissions"]; exists {
				if permsSlice, ok := perms.([]interface{}); ok {
					permStrings := make([]string, 0, len(permsSlice))
					for _, p := range permsSlice {
						if ps, ok := p.(string); ok {
							permStrings = append(permStrings, ps)
						}
					}
					c.Locals("userPermissions", permStrings)
				}
			}
		}

		return c.Next()
	}
}

// PermissionMiddleware verifica que el usuario tenga al menos uno de los permisos requeridos
func PermissionMiddleware(requiredPermissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		perms, _ := c.Locals("userPermissions").([]string)
		if len(perms) == 0 {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   true,
				"message": "Access denied: no permissions found",
			})
		}

		for _, required := range requiredPermissions {
			for _, userPerm := range perms {
				if userPerm == required {
					return c.Next()
				}
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": fmt.Sprintf("Access denied: requires one of %v", requiredPermissions),
		})
	}
}

// RoleMiddleware verifica que el usuario tenga alguno de los roles requeridos
func RoleMiddleware(requiredRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("role").(string)
		if role == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   true,
				"message": "Access denied: no role found",
			})
		}

		for _, r := range requiredRoles {
			if role == r {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   true,
			"message": fmt.Sprintf("Access denied: requires role %v", requiredRoles),
		})
	}
}
