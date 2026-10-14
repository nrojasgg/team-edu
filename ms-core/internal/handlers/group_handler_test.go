package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

// ─── resolveUserID ────────────────────────────────────────────────────────────

func TestResolveUserID_WithUserIDLocal_Float64(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		c.Locals("userID", float64(42))
		id, err := resolveUserID(c)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		if id != 42 {
			return c.Status(500).SendString("wrong id")
		}
		return c.SendString("ok")
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestResolveUserID_WithUserIDLocal_Int64(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		c.Locals("userID", int64(99))
		id, err := resolveUserID(c)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		if id != 99 {
			return c.Status(500).SendString("wrong id")
		}
		return c.SendString("ok")
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestResolveUserID_WithUserIDLocal_Int(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		c.Locals("userID", int(7))
		id, err := resolveUserID(c)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		if id != 7 {
			return c.Status(500).SendString("wrong id")
		}
		return c.SendString("ok")
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestResolveUserID_WithClaims(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		// No userID local, but claims map present
		c.Locals("claims", map[string]interface{}{"id": float64(55)})
		id, err := resolveUserID(c)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		if id != 55 {
			return c.Status(500).SendString("wrong id")
		}
		return c.SendString("ok")
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestResolveUserID_NilUserID(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		_, err := resolveUserID(c)
		if err == nil {
			return c.Status(500).SendString("expected error")
		}
		return c.SendString("ok")
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestResolveUserID_InvalidType(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		c.Locals("userID", "not-a-number")
		_, err := resolveUserID(c)
		if err == nil {
			return c.Status(500).SendString("expected error")
		}
		return c.SendString("ok")
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

// ─── toInt ────────────────────────────────────────────────────────────────────

func TestToInt_Float64(t *testing.T) {
	v, ok := toInt(float64(3))
	if !ok || v != 3 {
		t.Fatalf("expected 3/true, got %d/%v", v, ok)
	}
}

func TestToInt_String(t *testing.T) {
	v, ok := toInt("7")
	if !ok || v != 7 {
		t.Fatalf("expected 7/true, got %d/%v", v, ok)
	}
}

func TestToInt_StringInvalid(t *testing.T) {
	_, ok := toInt("abc")
	if ok {
		t.Fatal("expected false for non-numeric string")
	}
}

func TestToInt_Int(t *testing.T) {
	v, ok := toInt(int(5))
	if !ok || v != 5 {
		t.Fatalf("expected 5/true, got %d/%v", v, ok)
	}
}

func TestToInt_Unsupported(t *testing.T) {
	_, ok := toInt(struct{}{})
	if ok {
		t.Fatal("expected false for unsupported type")
	}
}

// ─── classifyStudentsByEnneagram ────────────────────────────────────────────

func TestClassifyStudentsByEnneagram_AllTyped(t *testing.T) {
	students := []bson.M{
		{"id": int64(1), "eneatype": "1"},
		{"id": int64(2), "eneatype": "2"},
		{"id": int64(3), "eneatype": "1"},
	}
	groups, untyped := classifyStudentsByEnneagram(students)
	if len(untyped) != 0 {
		t.Fatalf("expected 0 untyped, got %d", len(untyped))
	}
	if len(groups["1"]) != 2 {
		t.Fatalf("expected 2 in group '1', got %d", len(groups["1"]))
	}
	if len(groups["2"]) != 1 {
		t.Fatalf("expected 1 in group '2', got %d", len(groups["2"]))
	}
}

func TestClassifyStudentsByEnneagram_NoType(t *testing.T) {
	students := []bson.M{
		{"id": int64(1)},
		{"id": int64(2), "eneatype": ""},
	}
	groups, untyped := classifyStudentsByEnneagram(students)
	if len(untyped) != 2 {
		t.Fatalf("expected 2 untyped, got %d", len(untyped))
	}
	if len(groups) != 0 {
		t.Fatalf("expected 0 typed groups, got %d", len(groups))
	}
}

func TestClassifyStudentsByEnneagram_Mixed(t *testing.T) {
	students := []bson.M{
		{"id": int64(1), "eneatype": "3"},
		{"id": int64(2)},
	}
	groups, untyped := classifyStudentsByEnneagram(students)
	if len(untyped) != 1 {
		t.Fatalf("expected 1 untyped, got %d", len(untyped))
	}
	if len(groups["3"]) != 1 {
		t.Fatalf("expected 1 in group '3', got %d", len(groups["3"]))
	}
}

// ─── buildOCEANFromEneatype ──────────────────────────────────────────────────

func TestBuildOCEANFromEneatype_NilPluginManager(t *testing.T) {
	h := &GroupHandler{pluginManager: nil}
	ocean := h.buildOCEANFromEneatype(float64(1))
	// Should return zero-valued vector without panicking
	if ocean.Openness != 0 || ocean.Conscientiousness != 0 {
		t.Fatal("expected zero OCEAN for nil pm")
	}
}

func TestBuildOCEANFromEneatype_NilEneatype(t *testing.T) {
	h := &GroupHandler{pluginManager: nil}
	ocean := h.buildOCEANFromEneatype(nil)
	if ocean.Openness != 0 {
		t.Fatal("expected zero OCEAN for nil eneatype")
	}
}

func TestBuildOCEANFromEneatype_InvalidType(t *testing.T) {
	h := &GroupHandler{pluginManager: nil}
	ocean := h.buildOCEANFromEneatype("not-convertible")
	if ocean.Openness != 0 {
		t.Fatal("expected zero OCEAN for unconvertible string")
	}
}

func TestBuildOCEANFromEneatype_WithPluginError(t *testing.T) {
	// Plugin that returns error
	errPlugin := &testPlugin{
		name: "enneagram",
		execFn: func(_ context.Context, __ interface{}) (interface{}, error) {
			return nil, fiber.NewError(500, "plugin error")
		},
	}
	pm := buildPM(errPlugin)
	h := &GroupHandler{pluginManager: pm}
	ocean := h.buildOCEANFromEneatype(float64(1))
	// Plugin error → zero vector, no panic
	if ocean.Openness != 0 {
		t.Fatal("expected zero OCEAN when plugin errors")
	}
}

func TestBuildOCEANFromEneatype_WithValidOCEAN(t *testing.T) {
	// Plugin that returns a valid OCEAN map
	oceanPlugin := &testPlugin{
		name: "enneagram",
		execFn: func(_ context.Context, __ interface{}) (interface{}, error) {
			return map[string]interface{}{
				"openness":          0.8,
				"conscientiousness": 0.7,
				"extraversion":      0.6,
				"agreeableness":     0.5,
				"neuroticism":       0.4,
			}, nil
		},
	}
	pm := buildPM(oceanPlugin)
	h := &GroupHandler{pluginManager: pm}
	ocean := h.buildOCEANFromEneatype(float64(1))
	if ocean.Openness != 0.8 {
		t.Fatalf("expected Openness=0.8, got %v", ocean.Openness)
	}
}

// ─── HTTP handler validation paths (nil DB → panic before DB call if JSON invalid) ──

func makeGroupApp(h *GroupHandler) *fiber.App {
	app := fiber.New()
	app.Get("/sections/:sectionId/groups", h.GetGroupsBySectionID)
	app.Delete("/sections/:sectionId/groups", h.DeleteGroupsBySection)
	app.Delete("/sections/:sectionId/students/:studentId", h.RemoveStudentFromGroup)
	app.Get("/sections/:sectionId/student-groups", h.GetStudentGroupMates)
	app.Post("/sections/:sectionId/groups/random", h.FormGroupsByEnneagram)
	app.Post("/sections/:sectionId/groups/form-random", h.FormRandomGroups)
	app.Post("/sections/:sectionId/groups/export", h.ExportGroupsToExcel)
	return app
}

func TestGroupHandler_GetGroupsBySectionID_BadID(t *testing.T) {
	h := NewGroupHandler(nil, nil)
	app := makeGroupApp(h)
	req := httptest.NewRequest(http.MethodGet, "/sections/notanumber/groups", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGroupHandler_DeleteGroupsBySection_BadID(t *testing.T) {
	h := NewGroupHandler(nil, nil)
	app := makeGroupApp(h)
	req := httptest.NewRequest(http.MethodDelete, "/sections/badid/groups", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGroupHandler_RemoveStudentFromGroup_BadSectionID(t *testing.T) {
	h := NewGroupHandler(nil, nil)
	app := makeGroupApp(h)
	req := httptest.NewRequest(http.MethodDelete, "/sections/bad/students/1", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGroupHandler_RemoveStudentFromGroup_BadStudentID(t *testing.T) {
	h := NewGroupHandler(nil, nil)
	app := makeGroupApp(h)
	req := httptest.NewRequest(http.MethodDelete, "/sections/1/students/badid", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGroupHandler_GetStudentGroupMates_BadSectionID(t *testing.T) {
	h := NewGroupHandler(nil, nil)
	app := makeGroupApp(h)
	req := httptest.NewRequest(http.MethodGet, "/sections/bad/student-groups", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGroupHandler_GetStudentGroupMates_NoAuth(t *testing.T) {
	h := NewGroupHandler(nil, nil)
	app := makeGroupApp(h)
	// Valid sectionID but no userID set → 401
	req := httptest.NewRequest(http.MethodGet, "/sections/1/student-groups", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestGroupHandler_FormGroupsByEnneagram_BadID(t *testing.T) {
	h := NewGroupHandler(nil, nil)
	app := makeGroupApp(h)
	req := httptest.NewRequest(http.MethodPost, "/sections/bad/groups/random", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGroupHandler_FormRandomGroups_BadID(t *testing.T) {
	h := NewGroupHandler(nil, nil)
	app := makeGroupApp(h)
	req := httptest.NewRequest(http.MethodPost, "/sections/bad/groups/form-random", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGroupHandler_FormRandomGroups_NoAuth(t *testing.T) {
	h := NewGroupHandler(nil, nil)
	app := makeGroupApp(h)
	req := httptest.NewRequest(http.MethodPost, "/sections/1/groups/form-random", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestGroupHandler_ExportGroupsToExcel_BadID(t *testing.T) {
	h := NewGroupHandler(nil, nil)
	app := makeGroupApp(h)
	req := httptest.NewRequest(http.MethodPost, "/sections/bad/groups/export", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGroupHandler_ExportGroupsToExcel_NoAuth(t *testing.T) {
	h := NewGroupHandler(nil, nil)
	app := makeGroupApp(h)
	req := httptest.NewRequest(http.MethodPost, "/sections/1/groups/export", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

// ─── NewGroupHandler ─────────────────────────────────────────────────────────

func TestNewGroupHandler(t *testing.T) {
	pm := buildPM(echoPlugin("x"))
	h := NewGroupHandler(nil, pm)
	if h == nil {
		t.Fatal("expected non-nil GroupHandler")
	}
	if h.pluginManager != pm {
		t.Fatal("plugin manager not stored")
	}
}
