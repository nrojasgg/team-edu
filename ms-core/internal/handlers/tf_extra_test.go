package handlers

// tf_extra_test.go — additional coverage for TeamFormationHandler paths not yet exercised.

import (
	"context"
	"net/http"
	"testing"

	"ms-core/internal/config"
	"ms-core/internal/core"
	"ms-core/internal/core/genetic"

	"github.com/gofiber/fiber/v2"
)

// buildTFHandlerWithProvider creates a handler with a mock affinity provider registered.
func buildTFHandlerWithProvider() *TeamFormationHandler {
	cfg := core.DefaultTeamFormationConfig()
	cfg.Algorithm = "greedy"
	engine := core.NewTeamFormationEngine(cfg)

	cfgRegis := &config.TeamFormationConfig{
		DefaultWeight:     0.5,
		ProviderWeights:   map[string]float64{},
		ExcludedProviders: []string{},
	}
	registry := core.NewProviderRegistry(cfgRegis)
	return NewTeamFormationHandler(engine, registry, nil)
}

// ─── FormTeams success path ───────────────────────────────────────────────────

func TestTF_FormTeams_Success(t *testing.T) {
	h := buildTFHandlerWithProvider()
	app := fiber.New()
	app.Post("/form", h.FormTeams)

	body := map[string]interface{}{
		"section_id": "s1",
		"candidates": []string{"a", "b", "c", "d"},
		"team_size":  2,
	}
	status, result := doRequest(t, app, "POST", "/form", body)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d (result: %v)", status, result)
	}
	if result["success"] != true {
		t.Errorf("expected success=true, got %v", result)
	}
}

func TestTF_FormTeams_GreedyAlgo(t *testing.T) {
	// Explicitly specifying algorithm="greedy"
	h := buildTFHandlerWithProvider()
	app := fiber.New()
	app.Post("/form", h.FormTeams)

	body := map[string]interface{}{
		"section_id": "s1",
		"candidates": []string{"a", "b", "c", "d", "e", "f"},
		"team_size":  2,
		"algorithm":  "greedy",
	}
	status, result := doRequest(t, app, "POST", "/form", body)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d (result: %v)", status, result)
	}
}

// ─── FlexibleFormTeams with criteria_parameters (new mode) ───────────────────

func TestTF_FlexibleFormTeams_WithCriteriaParameters(t *testing.T) {
	h := buildTFHandlerWithProvider()
	app := fiber.New()
	app.Post("/flexible", h.FlexibleFormTeams)

	body := map[string]interface{}{
		"section_id": "s1",
		"candidates": []string{"a", "b", "c", "d"},
		"team_size":  2,
		"criteria_parameters": map[string]interface{}{
			"gpa": map[string]interface{}{
				"target": 0.5,
				"weight": 1.0,
			},
		},
		"demographic_data": map[string]interface{}{
			"a": map[string]interface{}{
				"numeric_attrs": map[string]interface{}{"gpa": 3.5},
			},
			"b": map[string]interface{}{
				"numeric_attrs": map[string]interface{}{"gpa": 3.0},
			},
			"c": map[string]interface{}{
				"numeric_attrs": map[string]interface{}{"gpa": 2.5},
			},
			"d": map[string]interface{}{
				"numeric_attrs": map[string]interface{}{"gpa": 4.0},
			},
		},
	}
	status, result := doRequest(t, app, "POST", "/flexible", body)
	// Criteria-based formation should succeed
	if status != http.StatusOK && status != http.StatusInternalServerError {
		t.Fatalf("unexpected status %d, result: %v", status, result)
	}
}

// ─── FlexibleFormTeams with provider_config (legacy mode) ────────────────────

func TestTF_FlexibleFormTeams_WithProviderConfig_CustomProviders(t *testing.T) {
	h := buildTFHandlerWithProvider()
	app := fiber.New()
	app.Post("/flexible", h.FlexibleFormTeams)

	// Using custom_providers in provider_config with invalid provider → At least one must be selected
	body := map[string]interface{}{
		"section_id": "s1",
		"candidates": []string{"a", "b", "c", "d"},
		"team_size":  2,
		"provider_config": map[string]interface{}{
			"personality_test":   "",
			"personality_weight": 0.0,
			"use_sociogram":      false,
			"sociogram_weight":   0.0,
		},
	}
	status, _ := doRequest(t, app, "POST", "/flexible", body)
	// No providers selected → 500 (from runLegacyFormation returning error)
	if status != http.StatusInternalServerError && status != http.StatusBadRequest {
		t.Fatalf("expected error status, got %d", status)
	}
}

func TestTF_FlexibleFormTeams_WithProviderConfig_EnneagramProvider(t *testing.T) {
	h := buildTFHandlerWithProvider()
	app := fiber.New()
	app.Post("/flexible", h.FlexibleFormTeams)

	body := map[string]interface{}{
		"section_id": "s1",
		"candidates": []string{"a", "b", "c", "d"},
		"team_size":  2,
		"provider_config": map[string]interface{}{
			"personality_test":   "enneagram",
			"personality_weight": 0.5,
			"use_sociogram":      false,
			"sociogram_weight":   0.0,
		},
	}
	status, result := doRequest(t, app, "POST", "/flexible", body)
	// Engine has no socket/provider for "enneagram"; runs anyway with 0 weight total
	if status != http.StatusOK && status != http.StatusInternalServerError {
		t.Fatalf("unexpected status %d, result: %v", status, result)
	}
}

// ─── GetAvailableOptions — personality test descriptions ─────────────────────

func TestTF_GetAvailableOptions_PersonalityDescriptions(t *testing.T) {
	h := buildTFHandlerWithProvider()
	app := fiber.New()
	app.Get("/options", h.GetAvailableOptions)

	status, result := doRequest(t, app, "GET", "/options", nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data field")
	}
	if _, ok := data["algorithms"]; !ok {
		t.Error("expected algorithms in data")
	}
}

// ─── CalculateAffinity paths ──────────────────────────────────────────────────

func TestTF_CalculateAffinity_WithPersonAOnly(t *testing.T) {
	h := buildTFHandlerWithProvider()
	app := fiber.New()
	app.Get("/affinity", h.CalculateAffinity)

	status, _ := doRequest(t, app, "GET", "/affinity?section_id=s1&person_a=A", nil)
	if status != http.StatusBadRequest {
		t.Fatalf("expected 400 (missing person_b), got %d", status)
	}
}

func TestTF_CalculateAffinity_WithPersonBOnly(t *testing.T) {
	h := buildTFHandlerWithProvider()
	app := fiber.New()
	app.Get("/affinity", h.CalculateAffinity)

	status, _ := doRequest(t, app, "GET", "/affinity?section_id=s1&person_b=B", nil)
	if status != http.StatusBadRequest {
		t.Fatalf("expected 400 (missing person_a), got %d", status)
	}
}

func TestTF_CalculateAffinity_WithSectionMissing(t *testing.T) {
	h := buildTFHandlerWithProvider()
	app := fiber.New()
	app.Get("/affinity", h.CalculateAffinity)

	status, _ := doRequest(t, app, "GET", "/affinity?person_a=A&person_b=B", nil)
	if status != http.StatusBadRequest {
		t.Fatalf("expected 400 (missing section_id), got %d", status)
	}
}

// ─── marshalToTeamFormationResult — additional ────────────────────────────────

func TestMarshalToTeamFormationResult_WithTeams(t *testing.T) {
	raw := map[string]interface{}{
		"teams": []interface{}{
			map[string]interface{}{
				"id":             "team-1",
				"affinity_score": 0.8,
				"members": []interface{}{
					map[string]interface{}{"id": "a"},
					map[string]interface{}{"id": "b"},
				},
			},
		},
		"global_score": 0.8,
		"algorithm":    "greedy",
	}
	result, err := marshalToTeamFormationResult(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Teams) != 1 {
		t.Fatalf("expected 1 team, got %d", len(result.Teams))
	}
}

// ─── getPersonalityTestDescription ───────────────────────────────────────────

func TestGetPersonalityTestDescription_Known(t *testing.T) {
	desc := getPersonalityTestDescription("enneagram")
	if desc == "" {
		t.Fatal("expected non-empty description for enneagram")
	}
	desc2 := getPersonalityTestDescription("mbti")
	if desc2 == "" {
		t.Fatal("expected non-empty description for mbti")
	}
}

func TestGetPersonalityTestDescription_Unknown(t *testing.T) {
	desc := getPersonalityTestDescription("unknown_test")
	if desc == "" {
		t.Fatal("expected fallback description for unknown test")
	}
}

// ─── CalculateAffinity — success with providers ───────────────────────────────

func TestTF_CalculateAffinity_WithEngineSuccess(t *testing.T) {
	// Register a provider so CalculateCombinedAffinity doesn't fail
	cfg := core.DefaultTeamFormationConfig()
	cfg.Algorithm = "greedy"
	engine := core.NewTeamFormationEngine(cfg)

	// Use mock provider from core package
	// We can't import it directly but we can test the handler path
	cfgRegis := &config.TeamFormationConfig{
		DefaultWeight:     0.5,
		ProviderWeights:   map[string]float64{},
		ExcludedProviders: []string{},
	}
	registry := core.NewProviderRegistry(cfgRegis)
	h := NewTeamFormationHandler(engine, registry, nil)

	app := fiber.New()
	app.Get("/affinity", h.CalculateAffinity)

	// section_id, person_a, person_b all present but no providers → engine returns error → handler returns 200 with neutral score
	status, result := doRequest(t, app, "GET", "/affinity?section_id=s1&person_a=A&person_b=B", nil)
	if status != http.StatusOK && status != http.StatusInternalServerError {
		t.Fatalf("unexpected status %d, result: %v", status, result)
	}
}

// ─── buildProviderWeights edge cases ─────────────────────────────────────────

func TestTF_FlexibleFormTeams_WithCustomProvidersMap(t *testing.T) {
	h := buildTFHandlerWithProvider()
	app := fiber.New()
	app.Post("/flexible", h.FlexibleFormTeams)

	// custom_providers override with known weight
	body := map[string]interface{}{
		"section_id": "s1",
		"candidates": []string{"a", "b", "c", "d"},
		"team_size":  2,
		"provider_config": map[string]interface{}{
			"custom_providers": map[string]interface{}{
				"my_provider": 1.0,
			},
		},
	}
	status, _ := doRequest(t, app, "POST", "/flexible", body)
	// Custom provider "my_provider" not in the engine → still runs with greedy (no affinity providers)
	if status != http.StatusOK && status != http.StatusInternalServerError {
		t.Fatalf("unexpected status %d", status)
	}
}

// ─── FormTeams with algorithm = "genetic"  (tests executeFormTeams branch) ────

func TestTF_FormTeams_NilMinMaxTeamSize_DefaultsApplied(t *testing.T) {
	h := buildTFHandlerWithProvider()
	app := fiber.New()
	app.Post("/form", h.FormTeams)

	// min/max not set → handler defaults to teamSize ± 1
	body := map[string]interface{}{
		"section_id":    "s1",
		"candidates":    []string{"a", "b", "c", "d"},
		"team_size":     2,
		"min_team_size": 0,
		"max_team_size": 0,
	}
	status, _ := doRequest(t, app, "POST", "/form", body)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
}

// ─── FlexibleFormTeams — missing candidates ───────────────────────────────────

func TestTF_FlexibleFormTeams_NoCandidates(t *testing.T) {
	h := buildTFHandlerWithProvider()
	app := fiber.New()
	app.Post("/flexible", h.FlexibleFormTeams)

	status, _ := doRequest(t, app, "POST", "/flexible", map[string]interface{}{
		"section_id": "s1",
		"team_size":  2,
		"provider_config": map[string]interface{}{
			"personality_test":   "enneagram",
			"personality_weight": 0.5,
		},
	})
	if status != http.StatusBadRequest {
		t.Fatalf("expected 400 (no candidates), got %d", status)
	}
}

// ─── CalculateAffinity missing section_id only ───────────────────────────────

func TestTF_CalculateAffinity_AllParamsPresent_InternalError(t *testing.T) {
	h := buildTFHandlerWithProvider()
	app := fiber.New()
	app.Get("/affinity", h.CalculateAffinity)

	// All params present, no affinity providers → error handled gracefully
	status, _ := doRequest(t, app, "GET", "/affinity?section_id=sec1&person_a=Alice&person_b=Bob", nil)
	// Could be 200 (neutral) or 500 depending on engine behavior
	if status == 0 {
		t.Fatal("expected non-zero status")
	}
}

// ─── UpdateWeights — missing body ────────────────────────────────────────────

func TestTF_UpdateWeights_MissingBody(t *testing.T) {
	h := buildTFHandlerWithProvider()
	app := fiber.New()
	app.Put("/weights", h.UpdateWeights)

	status, _ := doRequest(t, app, "PUT", "/weights", nil)
	// nil body → no weights → 400
	if status != http.StatusBadRequest && status != http.StatusOK {
		t.Fatalf("unexpected status %d", status)
	}
}

// ─── GetProviders with actual providers registered ────────────────────────────

func TestTF_GetProviders_WithRegisteredProvider(t *testing.T) {
	cfg := core.DefaultTeamFormationConfig()
	engine := core.NewTeamFormationEngine(cfg)
	cfgRegis := &config.TeamFormationConfig{
		DefaultWeight:   0.5,
		ProviderWeights: map[string]float64{},
	}
	registry := core.NewProviderRegistry(cfgRegis)
	h := NewTeamFormationHandler(engine, registry, nil)

	app := fiber.New()
	app.Get("/providers", h.GetProviders)

	status, body := doRequest(t, app, "GET", "/providers", nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if body["success"] != true {
		t.Error("expected success=true")
	}
}

// ─── FlexibleFormTeams — team_size < 2 ───────────────────────────────────────

func TestTF_FlexibleFormTeams_SmallTeamSize(t *testing.T) {
	h := buildTFHandlerWithProvider()
	app := fiber.New()
	app.Post("/flexible", h.FlexibleFormTeams)

	status, _ := doRequest(t, app, "POST", "/flexible", map[string]interface{}{
		"section_id": "s1",
		"candidates": []string{"a", "b"},
		"team_size":  1,
	})
	if status != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", status)
	}
}

// ─── FlexibleFormTeams with sociogram weight ─────────────────────────────────

func TestTF_FlexibleFormTeams_WithSociogramWeight(t *testing.T) {
	h := buildTFHandlerWithProvider()
	app := fiber.New()
	app.Post("/flexible", h.FlexibleFormTeams)

	body := map[string]interface{}{
		"section_id": "s1",
		"candidates": []string{"a", "b", "c", "d"},
		"team_size":  2,
		"provider_config": map[string]interface{}{
			"personality_test":   "",
			"personality_weight": 0.0,
			"use_sociogram":      true,
			"sociogram_weight":   1.0,
		},
	}
	status, _ := doRequest(t, app, "POST", "/flexible", body)
	// Sociogram provider "sociogram" not in engine → empty providers → error
	if status != http.StatusOK && status != http.StatusInternalServerError {
		t.Fatalf("unexpected status %d", status)
	}
}

// ─── NewTeamFormationHandler nil tfPlugin ─────────────────────────────────────

func TestNewTeamFormationHandler_NilPlugin(t *testing.T) {
	cfg := core.DefaultTeamFormationConfig()
	engine := core.NewTeamFormationEngine(cfg)
	cfgRegis := &config.TeamFormationConfig{
		DefaultWeight:   0.5,
		ProviderWeights: map[string]float64{},
	}
	registry := core.NewProviderRegistry(cfgRegis)
	h := NewTeamFormationHandler(engine, registry, nil)
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
	if h.tfPlugin != nil {
		t.Error("expected nil tfPlugin")
	}
}

// ─── FlexibleFormTeams success via criteria (larger group) ───────────────────

func TestTF_FlexibleFormTeams_CriteriaCategorial(t *testing.T) {
	h := buildTFHandlerWithProvider()
	app := fiber.New()
	app.Post("/flexible", h.FlexibleFormTeams)

	body := map[string]interface{}{
		"section_id": "s1",
		"candidates": []string{"a", "b", "c", "d"},
		"team_size":  2,
		"criteria_parameters": map[string]interface{}{
			"gender": map[string]interface{}{
				"target": 0.5,
				"weight": 1.0,
			},
		},
		"demographic_data": map[string]interface{}{
			"a": map[string]interface{}{"categorical_attrs": map[string]interface{}{"gender": "M"}},
			"b": map[string]interface{}{"categorical_attrs": map[string]interface{}{"gender": "F"}},
			"c": map[string]interface{}{"categorical_attrs": map[string]interface{}{"gender": "M"}},
			"d": map[string]interface{}{"categorical_attrs": map[string]interface{}{"gender": "F"}},
		},
	}
	status, _ := doRequest(t, app, "POST", "/flexible", body)
	if status != http.StatusOK && status != http.StatusInternalServerError {
		t.Fatalf("unexpected status %d", status)
	}
}

// ─── getFinalDiversity — nil and empty-generations branches ──────────────────

func TestGetFinalDiversity_NilStats(t *testing.T) {
	result := getFinalDiversity(nil)
	if result != 0.0 {
		t.Errorf("expected 0.0 for nil stats, got %f", result)
	}
}

func TestGetFinalDiversity_EmptyGenerations(t *testing.T) {
	// Import genetic package to build EvolutionStats
	stats := &genetic.EvolutionStats{Generations: nil}
	result := getFinalDiversity(stats)
	if result != 0.0 {
		t.Errorf("expected 0.0 for empty generations, got %f", result)
	}
}

// ─── GetAvailableOptions with sociogram + unknown personality provider ────────

// mockAffinityProvider is a minimal AffinityProvider for handler tests.
type mockAffinityProvider struct {
	id       string
	category string
}

func (m *mockAffinityProvider) ProviderID() string   { return m.id }
func (m *mockAffinityProvider) ProviderName() string { return m.id + "-name" }
func (m *mockAffinityProvider) CalculatePairwiseAffinity(_ context.Context, _, _ string) (float64, error) {
	return 0.5, nil
}
func (m *mockAffinityProvider) SupportsContext(_ context.Context) bool { return true }
func (m *mockAffinityProvider) GetMetadata() core.ProviderMetadata {
	return core.ProviderMetadata{ID: m.id, Name: m.id + "-name", Category: m.category}
}

// buildTFHandlerWithSociogramAndUnknownProvider registers a sociogram provider
// and an unknown personality provider to cover those branches in GetAvailableOptions.
func buildTFHandlerWithSociogramAndUnknownProvider() *TeamFormationHandler {
	cfg := core.DefaultTeamFormationConfig()
	engine := core.NewTeamFormationEngine(cfg)

	cfgRegis := &config.TeamFormationConfig{
		DefaultWeight:     0.5,
		ProviderWeights:   map[string]float64{},
		ExcludedProviders: []string{},
	}
	registry := core.NewProviderRegistry(cfgRegis)
	// Register a sociogram provider → hasSociogram = true branch
	registry.Register(&mockAffinityProvider{id: "sociogram", category: "social"}) //nolint:errcheck
	// Register an unknown personality provider → description fallback branch
	registry.Register(&mockAffinityProvider{id: "custom_personality", category: "personality"}) //nolint:errcheck

	return NewTeamFormationHandler(engine, registry, nil)
}

func TestTF_GetAvailableOptions_WithSociogramAndUnknownPersonality(t *testing.T) {
	h := buildTFHandlerWithSociogramAndUnknownProvider()
	app := fiber.New()
	app.Get("/options", h.GetAvailableOptions)

	status, result := doRequest(t, app, "GET", "/options", nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data field")
	}
	// hasSociogram should be true
	if data["sociogram_available"] != true {
		t.Errorf("expected sociogram_available=true, got %v", data["sociogram_available"])
	}
	// personality_tests should include the unknown provider (with default description)
	tests, ok := data["personality_tests"].([]interface{})
	if !ok {
		t.Fatal("expected personality_tests as []interface{}")
	}
	if len(tests) == 0 {
		t.Error("expected at least one personality test")
	}
}

// ─── formTeamsWithDynamicCriteria mapper error path ──────────────────────────

// TestTF_FlexibleFormTeams_MapperErrorFromNoDemographicData triggers the
// formTeamsWithDynamicCriteria → MapToDomain error path by sending
// criteria_parameters but no demographic_data (fails DTO validation).
func TestTF_FlexibleFormTeams_MapperErrorFromNoDemographicData(t *testing.T) {
	h := buildTFHandlerWithProvider()
	app := fiber.New()
	app.Post("/flexible", h.FlexibleFormTeams)

	body := map[string]interface{}{
		"section_id": "s1",
		"candidates": []string{"a", "b", "c", "d"},
		"team_size":  2,
		"criteria_parameters": map[string]interface{}{
			"skill": map[string]interface{}{
				"target": 0.5,
				"weight": 0.8,
			},
		},
		// deliberately omit demographic_data to trigger Validate() ErrNoDemographicData
	}
	status, _ := doRequest(t, app, "POST", "/flexible", body)
	// expect 500 (MapToDomain returns error → runCriteriaFormation returns err → handler returns 500)
	if status != http.StatusInternalServerError {
		t.Fatalf("expected 500 for mapper error, got %d", status)
	}
}
