package middleware

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// extractBearerToken parses "Bearer <token>" from an Authorization header value.
// Returns the token string and true on success, or empty string and false otherwise.
func extractBearerToken(authHeader string) (string, bool) {
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", false
	}
	return parts[1], true
}

// jwtKeyFunc builds a jwt.Keyfunc that validates the signing method and returns the secret.
func jwtKeyFunc(secret string) jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.NewError(401, "Invalid signing method")
		}
		return []byte(secret), nil
	}
}

// applyClaimsToContext copies JWT claims (user ID, role, raw claims) into the Fiber context.
func applyClaimsToContext(c *fiber.Ctx, token *jwt.Token) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return
	}
	if id, exists := claims["id"]; exists {
		c.Locals("userID", id)
	} else if userID, exists := claims["user_id"]; exists {
		c.Locals("userID", userID)
	}
	if role, exists := claims["role"]; exists {
		c.Locals("userRole", role)
	}
	c.Locals("claims", claims)

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

// AuthMiddleware valida el token JWT
func AuthMiddleware(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{
				"error":   true,
				"message": "Authorization header is required",
			})
		}

		tokenString, ok := extractBearerToken(authHeader)
		if !ok {
			return c.Status(401).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid authorization header format",
			})
		}

		token, err := jwt.Parse(tokenString, jwtKeyFunc(jwtSecret))
		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid token",
			})
		}

		if !token.Valid {
			return c.Status(401).JSON(fiber.Map{
				"error":   true,
				"message": "Token is not valid",
			})
		}

		applyClaimsToContext(c, token)
		return c.Next()
	}
}

// OptionalAuthMiddleware valida el token JWT pero no requiere autenticación
func OptionalAuthMiddleware(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		tokenString, ok := extractBearerToken(authHeader)
		if !ok {
			return c.Next()
		}

		token, err := jwt.Parse(tokenString, jwtKeyFunc(jwtSecret))
		if err == nil && token.Valid {
			applyClaimsToContext(c, token)
		}

		return c.Next()
	}
}

// RoleMiddleware verifica que el usuario tenga el rol requerido
func RoleMiddleware(requiredRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := c.Locals("userRole")
		if userRole == nil {
			return c.Status(403).JSON(fiber.Map{
				"error":   true,
				"message": "Access denied: no role found",
			})
		}

		roleStr, _ := userRole.(string)
		for _, role := range requiredRoles {
			if roleStr == role {
				return c.Next()
			}
		}

		return c.Status(403).JSON(fiber.Map{
			"error":   true,
			"message": fmt.Sprintf("Access denied: requires role %v", requiredRoles),
		})
	}
}

// PermissionMiddleware verifica que el usuario tenga al menos uno de los permisos requeridos
func PermissionMiddleware(requiredPermissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		perms, _ := c.Locals("userPermissions").([]string)
		if len(perms) == 0 {
			return c.Status(403).JSON(fiber.Map{
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

		return c.Status(403).JSON(fiber.Map{
			"error":   true,
			"message": fmt.Sprintf("Access denied: requires one of %v", requiredPermissions),
		})
	}
}
