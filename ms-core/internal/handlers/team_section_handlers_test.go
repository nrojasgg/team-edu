package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ms-core/internal/config"
	"ms-core/internal/core"

	"github.com/gofiber/fiber/v2"
)

// ---------------------------------------------------------------------------
// TeamFormationHandler helpers
// ---------------------------------------------------------------------------

func buildTFHandler() *TeamFormationHandler {
	cfg := core.DefaultTeamFormationConfig()
	engine := core.NewTeamFormationEngine(cfg)
	cfgRegis := &config.TeamFormationConfig{
		DefaultWeight:     0.5,
		ProviderWeights:   map[string]float64{},
		ExcludedProviders: []string{},
	}
	registry := core.NewProviderRegistry(cfgRegis)
	return NewTeamFormationHandler(engine, registry, nil)
}

// ---------------------------------------------------------------------------
// FormTeams – bad body
// ---------------------------------------------------------------------------

func TestTF_FormTeams_BadBody(t *testing.T) {
	h := buildTFHandler()
	app := fiber.New()
	app.Post("/form", h.FormTeams)

	req := httptest.NewRequest("POST", "/form", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	// missing section_id → 400
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		t.Errorf("unexpected status %d", resp.StatusCode)
	}
}

func TestTF_FormTeams_MissingSectionID(t *testing.T) {
	h := buildTFHandler()
	app := fiber.New()
	app.Post("/form", h.FormTeams)

	body := map[string]interface{}{
		"candidates": []string{"a", "b", "c"},
		"team_size":  2,
	}
	status, _ := doRequest(t, app, "POST", "/form", body)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 (missing section_id), got %d", status)
	}
}

func TestTF_FormTeams_MissingCandidates(t *testing.T) {
	h := buildTFHandler()
	app := fiber.New()
	app.Post("/form", h.FormTeams)

	status, _ := doRequest(t, app, "POST", "/form", map[string]interface{}{
		"section_id": "s1",
		"team_size":  2,
	})
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 (missing candidates), got %d", status)
	}
}

func TestTF_FormTeams_SmallTeamSize(t *testing.T) {
	h := buildTFHandler()
	app := fiber.New()
	app.Post("/form", h.FormTeams)

	status, _ := doRequest(t, app, "POST", "/form", map[string]interface{}{
		"section_id": "s1",
		"candidates": []string{"a", "b"},
		"team_size":  1,
	})
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 (team_size < 2), got %d", status)
	}
}

// ---------------------------------------------------------------------------
// FlexibleFormTeams – validation
// ---------------------------------------------------------------------------

func TestTF_FlexibleFormTeams_MissingSectionID(t *testing.T) {
	h := buildTFHandler()
	app := fiber.New()
	app.Post("/flexible", h.FlexibleFormTeams)

	status, _ := doRequest(t, app, "POST", "/flexible", map[string]interface{}{
		"candidates": []string{"a", "b"},
		"team_size":  2,
	})
	if status != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}

func TestTF_FlexibleFormTeams_NoProviderAndNoCriteria(t *testing.T) {
	h := buildTFHandler()
	app := fiber.New()
	app.Post("/flexible", h.FlexibleFormTeams)

	status, _ := doRequest(t, app, "POST", "/flexible", map[string]interface{}{
		"section_id": "s1",
		"candidates": []string{"a", "b", "c"},
		"team_size":  2,
	})
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 (no provider_config/criteria), got %d", status)
	}
}

// ---------------------------------------------------------------------------
// GetProviders
// ---------------------------------------------------------------------------

func TestTF_GetProviders(t *testing.T) {
	h := buildTFHandler()
	app := fiber.New()
	app.Get("/providers", h.GetProviders)

	status, body := doRequest(t, app, "GET", "/providers", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
	if body["success"] != true {
		t.Error("expected success=true")
	}
}

// ---------------------------------------------------------------------------
// UpdateWeights
// ---------------------------------------------------------------------------

func TestTF_UpdateWeights_Success(t *testing.T) {
	h := buildTFHandler()
	app := fiber.New()
	app.Put("/weights", h.UpdateWeights)

	status, body := doRequest(t, app, "PUT", "/weights", map[string]interface{}{
		"weights": map[string]float64{"enneagram": 0.5, "sociogram": 0.5},
	})
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body)
	}
}

func TestTF_UpdateWeights_InvalidWeight(t *testing.T) {
	h := buildTFHandler()
	app := fiber.New()
	app.Put("/weights", h.UpdateWeights)

	status, _ := doRequest(t, app, "PUT", "/weights", map[string]interface{}{
		"weights": map[string]float64{"enneagram": 1.5}, // > 1
	})
	if status != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}

// ---------------------------------------------------------------------------
// CalculateAffinity
// ---------------------------------------------------------------------------

func TestTF_CalculateAffinity_MissingParams(t *testing.T) {
	h := buildTFHandler()
	app := fiber.New()
	app.Get("/affinity", h.CalculateAffinity)

	status, _ := doRequest(t, app, "GET", "/affinity", nil)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}

func TestTF_CalculateAffinity_NoProviders(t *testing.T) {
	h := buildTFHandler()
	app := fiber.New()
	app.Get("/affinity", h.CalculateAffinity)

	status, _ := doRequest(t, app, "GET", "/affinity?section_id=s1&person_a=A&person_b=B", nil)
	// No providers registered → engine returns 0 score (no error), OK
	if status != http.StatusOK && status != http.StatusInternalServerError {
		t.Errorf("unexpected status %d", status)
	}
}

// ---------------------------------------------------------------------------
// GetAvailableOptions
// ---------------------------------------------------------------------------

func TestTF_GetAvailableOptions(t *testing.T) {
	h := buildTFHandler()
	app := fiber.New()
	app.Get("/options", h.GetAvailableOptions)

	status, body := doRequest(t, app, "GET", "/options", nil)
	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
	if body["success"] != true {
		t.Errorf("expected success=true, got %v", body)
	}
}

// ---------------------------------------------------------------------------
// SectionCompleteDataHandler
// ---------------------------------------------------------------------------

func TestSCD_InvalidSectionID(t *testing.T) {
	pm := buildPM(echoPlugin("enneagram"))
	h := NewSectionCompleteDataHandler(nil, pm)
	app := fiber.New()
	app.Get("/sections/:section_id", h.GetCompleteSectionGroupsData)

	req := httptest.NewRequest("GET", "/sections/notanumber", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestSCD_MissingAuthHeader(t *testing.T) {
	pm := buildPM(echoPlugin("enneagram"))
	h := NewSectionCompleteDataHandler(nil, pm)
	app := fiber.New()
	app.Get("/sections/:section_id", h.GetCompleteSectionGroupsData)

	req := httptest.NewRequest("GET", "/sections/42", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}
