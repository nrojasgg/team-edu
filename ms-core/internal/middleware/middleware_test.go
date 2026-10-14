package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// ─── helper: build a signed JWT ──────────────────────────────────────────────

const testSecret = "test-secret"

func makeToken(secret string, claims jwt.MapClaims, expired bool) string {
	if expired {
		claims["exp"] = time.Now().Add(-time.Hour).Unix()
	} else {
		claims["exp"] = time.Now().Add(time.Hour).Unix()
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := tok.SignedString([]byte(secret))
	return signed
}

func doGet(app *fiber.App, path, authHeader string) *http.Response {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	resp, _ := app.Test(req, -1)
	return resp
}

// ─── extractBearerToken ──────────────────────────────────────────────────────

func TestExtractBearerToken_Valid(t *testing.T) {
	token, ok := extractBearerToken("Bearer mytoken123")
	if !ok || token != "mytoken123" {
		t.Fatalf("expected mytoken123 / true, got %q / %v", token, ok)
	}
}

func TestExtractBearerToken_MissingBearer(t *testing.T) {
	_, ok := extractBearerToken("Token abc")
	if ok {
		t.Fatal("expected false for non-Bearer scheme")
	}
}

func TestExtractBearerToken_EmptyString(t *testing.T) {
	_, ok := extractBearerToken("")
	if ok {
		t.Fatal("expected false for empty string")
	}
}

func TestExtractBearerToken_NoSpace(t *testing.T) {
	_, ok := extractBearerToken("Bearermytoken")
	if ok {
		t.Fatal("expected false for no space")
	}
}

func TestExtractBearerToken_ThreeParts(t *testing.T) {
	_, ok := extractBearerToken("Bearer a b")
	if ok {
		t.Fatal("expected false for three parts")
	}
}

// ─── AuthMiddleware ───────────────────────────────────────────────────────────

func TestAuthMiddleware_NoHeader(t *testing.T) {
	app := fiber.New()
	app.Use(AuthMiddleware(testSecret))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })

	resp := doGet(app, "/", "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	app := fiber.New()
	app.Use(AuthMiddleware(testSecret))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })

	resp := doGet(app, "/", "NotBearer token")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	app := fiber.New()
	app.Use(AuthMiddleware(testSecret))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })

	resp := doGet(app, "/", "Bearer notavalidtoken")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	app := fiber.New()
	app.Use(AuthMiddleware(testSecret))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })

	tok := makeToken(testSecret, jwt.MapClaims{"id": 1.0}, true)
	resp := doGet(app, "/", "Bearer "+tok)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	app := fiber.New()
	app.Use(AuthMiddleware(testSecret))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })

	tok := makeToken("wrong-secret", jwt.MapClaims{"id": 1.0}, false)
	resp := doGet(app, "/", "Bearer "+tok)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_ValidToken_WithID(t *testing.T) {
	app := fiber.New()
	app.Use(AuthMiddleware(testSecret))
	app.Get("/", func(c *fiber.Ctx) error {
		uid := c.Locals("userID")
		if uid == nil {
			return c.Status(500).SendString("no userID")
		}
		return c.SendString("ok")
	})

	tok := makeToken(testSecret, jwt.MapClaims{"id": float64(42), "role": "student"}, false)
	resp := doGet(app, "/", "Bearer "+tok)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_ValidToken_WithUserID(t *testing.T) {
	// Uses "user_id" claim instead of "id"
	app := fiber.New()
	app.Use(AuthMiddleware(testSecret))
	app.Get("/", func(c *fiber.Ctx) error {
		uid := c.Locals("userID")
		if uid == nil {
			return c.Status(500).SendString("no userID")
		}
		return c.SendString("ok")
	})

	tok := makeToken(testSecret, jwt.MapClaims{"user_id": float64(99)}, false)
	resp := doGet(app, "/", "Bearer "+tok)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_ValidToken_SetsRole(t *testing.T) {
	app := fiber.New()
	app.Use(AuthMiddleware(testSecret))
	app.Get("/", func(c *fiber.Ctx) error {
		r := c.Locals("userRole")
		if r != "teacher" {
			return c.Status(500).SendString("wrong role")
		}
		return c.SendString("ok")
	})

	tok := makeToken(testSecret, jwt.MapClaims{"id": float64(1), "role": "teacher"}, false)
	resp := doGet(app, "/", "Bearer "+tok)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

// ─── OptionalAuthMiddleware ───────────────────────────────────────────────────

func TestOptionalAuth_NoHeader(t *testing.T) {
	app := fiber.New()
	app.Use(OptionalAuthMiddleware(testSecret))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })

	resp := doGet(app, "/", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 (no auth required), got %d", resp.StatusCode)
	}
}

func TestOptionalAuth_InvalidFormat(t *testing.T) {
	// Should still pass through (optional)
	app := fiber.New()
	app.Use(OptionalAuthMiddleware(testSecret))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })

	resp := doGet(app, "/", "NotBearer xyz")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestOptionalAuth_InvalidToken(t *testing.T) {
	// Invalid token → passes through, no claims set
	app := fiber.New()
	app.Use(OptionalAuthMiddleware(testSecret))
	app.Get("/", func(c *fiber.Ctx) error {
		uid := c.Locals("userID")
		if uid != nil {
			return c.Status(500).SendString("should not have userID")
		}
		return c.SendString("ok")
	})

	resp := doGet(app, "/", "Bearer invalid.token.here")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestOptionalAuth_ValidToken(t *testing.T) {
	app := fiber.New()
	app.Use(OptionalAuthMiddleware(testSecret))
	app.Get("/", func(c *fiber.Ctx) error {
		uid := c.Locals("userID")
		if uid == nil {
			return c.Status(500).SendString("no userID")
		}
		return c.SendString("ok")
	})

	tok := makeToken(testSecret, jwt.MapClaims{"id": float64(7)}, false)
	resp := doGet(app, "/", "Bearer "+tok)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

// ─── RoleMiddleware ───────────────────────────────────────────────────────────

func TestRoleMiddleware_NoRole(t *testing.T) {
	app := fiber.New()
	app.Use(RoleMiddleware("admin"))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })

	resp := doGet(app, "/", "")
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}

func TestRoleMiddleware_WrongRole(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userRole", "student")
		return c.Next()
	})
	app.Use(RoleMiddleware("admin"))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })

	resp := doGet(app, "/", "")
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}

func TestRoleMiddleware_MatchingRole(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userRole", "admin")
		return c.Next()
	})
	app.Use(RoleMiddleware("teacher", "admin"))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })

	resp := doGet(app, "/", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRoleMiddleware_MultipleRoles(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("userRole", "teacher")
		return c.Next()
	})
	app.Use(RoleMiddleware("teacher", "admin"))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })

	resp := doGet(app, "/", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

// ─── applyClaimsToContext ─────────────────────────────────────────────────────

func TestApplyClaimsToContext_NoClaims(t *testing.T) {
	// When token has no MapClaims — should not panic
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		tok := &jwt.Token{Claims: jwt.RegisteredClaims{}}
		applyClaimsToContext(c, tok)
		return c.SendString("ok")
	})
	resp := doGet(app, "/", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestApplyClaimsToContext_WithClaims(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		claims := jwt.MapClaims{"id": float64(5), "role": "admin"}
		tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tok.Valid = true
		applyClaimsToContext(c, tok)
		if c.Locals("userID") == nil {
			return c.Status(500).SendString("no userID")
		}
		if c.Locals("userRole") != "admin" {
			return c.Status(500).SendString("wrong role")
		}
		return c.SendString("ok")
	})
	resp := doGet(app, "/", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

// ─── jwtKeyFunc ──────────────────────────────────────────────────────────────

func TestJwtKeyFunc_HMACMethod(t *testing.T) {
	kf := jwtKeyFunc("secret")
	tok := &jwt.Token{Method: jwt.SigningMethodHS256}
	key, err := kf(tok)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(key.([]byte)) != "secret" {
		t.Fatalf("unexpected key: %v", key)
	}
}

func TestJwtKeyFunc_NonHMACMethod(t *testing.T) {
	kf := jwtKeyFunc("secret")
	tok := &jwt.Token{Method: jwt.SigningMethodRS256}
	_, err := kf(tok)
	if err == nil {
		t.Fatal("expected error for non-HMAC method")
	}
}
