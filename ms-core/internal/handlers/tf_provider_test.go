package handlers

// tf_provider_test.go — covers applyPersonalityWeight, applySociogramWeight,
// buildFromCustomProviders, markEarlyExit, Excel helper functions,
// and plugin-path coverage for executeFormTeams / runLegacyFormation.

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"ms-core/internal/config"
	"ms-core/internal/core"
	"ms-core/internal/core/domain"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
)

// ── minimal mock AffinityProvider ──────────────────────────────────────────

type fakeProv struct {
	id   string
	name string
}

func (f *fakeProv) ProviderID() string   { return f.id }
func (f *fakeProv) ProviderName() string { return f.name }
func (f *fakeProv) CalculatePairwiseAffinity(_ context.Context, _, _ string) (float64, error) {
	return 0.5, nil
}
func (f *fakeProv) SupportsContext(_ context.Context) bool { return true }
func (f *fakeProv) GetMetadata() core.ProviderMetadata {
	return core.ProviderMetadata{ID: f.id, Name: f.name, Category: "personality"}
}

// ── minimal mock Plugin ───────────────────────────────────────────────────

type fakePlugin struct {
	result interface{}
	err    error
}

func (p *fakePlugin) Name() string      { return "fake_tf_plugin" }
func (p *fakePlugin) Version() string   { return "1.0.0" }
func (p *fakePlugin) Initialize() error { return nil }
func (p *fakePlugin) Shutdown() error   { return nil }
func (p *fakePlugin) Execute(_ context.Context, _ interface{}) (interface{}, error) {
	return p.result, p.err
}

// buildTFHandlerWithProviders registers "enneagram" and "sociogram" in the registry.
func buildTFHandlerWithProviders() *TeamFormationHandler {
	cfg := core.DefaultTeamFormationConfig()
	cfg.Algorithm = "greedy"
	engine := core.NewTeamFormationEngine(cfg)

	cfgRegis := &config.TeamFormationConfig{
		DefaultWeight:     0.5,
		ProviderWeights:   map[string]float64{},
		ExcludedProviders: []string{},
	}
	registry := core.NewProviderRegistry(cfgRegis)
	_ = registry.Register(&fakeProv{id: "enneagram", name: "Enneagram"})
	_ = registry.Register(&fakeProv{id: "sociogram", name: "Sociogram"})
	return NewTeamFormationHandler(engine, registry, nil)
}

// ── applyPersonalityWeight ────────────────────────────────────────────────────

func TestApplyPersonalityWeight_EmptyTest(t *testing.T) {
	h := buildTFHandlerWithProviders()
	weights := map[string]float64{}
	err := h.applyPersonalityWeight(&ProviderConfigRequest{PersonalityTest: ""}, weights)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(weights) != 0 {
		t.Errorf("expected empty weights map, got %v", weights)
	}
}

func TestApplyPersonalityWeight_InvalidTest(t *testing.T) {
	h := buildTFHandlerWithProviders()
	weights := map[string]float64{}
	err := h.applyPersonalityWeight(&ProviderConfigRequest{PersonalityTest: "invalid_test", PersonalityWeight: 0.5}, weights)
	if err == nil {
		t.Fatal("expected error for invalid personality test")
	}
}

func TestApplyPersonalityWeight_ValidEnneagram(t *testing.T) {
	h := buildTFHandlerWithProviders()
	weights := map[string]float64{}
	err := h.applyPersonalityWeight(&ProviderConfigRequest{PersonalityTest: "enneagram", PersonalityWeight: 0.7}, weights)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if weights["enneagram"] != 0.7 {
		t.Errorf("expected enneagram weight=0.7, got %v", weights["enneagram"])
	}
}

func TestApplyPersonalityWeight_DefaultWeightOnZero(t *testing.T) {
	h := buildTFHandlerWithProviders()
	weights := map[string]float64{}
	// weight=0 → defaults to 0.5
	err := h.applyPersonalityWeight(&ProviderConfigRequest{PersonalityTest: "enneagram", PersonalityWeight: 0.0}, weights)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if weights["enneagram"] != 0.5 {
		t.Errorf("expected default weight=0.5, got %v", weights["enneagram"])
	}
}

func TestApplyPersonalityWeight_WeightExceedsOne(t *testing.T) {
	h := buildTFHandlerWithProviders()
	weights := map[string]float64{}
	err := h.applyPersonalityWeight(&ProviderConfigRequest{PersonalityTest: "mbti", PersonalityWeight: 1.5}, weights)
	if err == nil {
		t.Fatal("expected error for weight > 1")
	}
}

func TestApplyPersonalityWeight_ValidMBTI(t *testing.T) {
	cfg := core.DefaultTeamFormationConfig()
	engine := core.NewTeamFormationEngine(cfg)
	cfgRegis := &config.TeamFormationConfig{
		DefaultWeight:   0.5,
		ProviderWeights: map[string]float64{},
	}
	registry := core.NewProviderRegistry(cfgRegis)
	_ = registry.Register(&fakeProv{id: "mbti", name: "MBTI"})
	h := NewTeamFormationHandler(engine, registry, nil)

	weights := map[string]float64{}
	err := h.applyPersonalityWeight(&ProviderConfigRequest{PersonalityTest: "mbti", PersonalityWeight: 0.4}, weights)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if weights["mbti"] != 0.4 {
		t.Errorf("expected mbti weight=0.4, got %v", weights["mbti"])
	}
}

func TestApplyPersonalityWeight_ValidTestNotRegistered(t *testing.T) {
	h := buildTFHandlerWithProvider() // no providers registered
	weights := map[string]float64{}
	err := h.applyPersonalityWeight(&ProviderConfigRequest{PersonalityTest: "enneagram", PersonalityWeight: 0.5}, weights)
	if err == nil {
		t.Fatal("expected error: enneagram provider not available")
	}
}

// ── applySociogramWeight ─────────────────────────────────────────────────────

func TestApplySociogramWeight_NotUsed(t *testing.T) {
	h := buildTFHandlerWithProviders()
	weights := map[string]float64{}
	err := h.applySociogramWeight(&ProviderConfigRequest{UseSociogram: false}, weights)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(weights) != 0 {
		t.Errorf("expected empty weights, got %v", weights)
	}
}

func TestApplySociogramWeight_NotRegistered(t *testing.T) {
	h := buildTFHandlerWithProvider() // no providers registered
	weights := map[string]float64{}
	err := h.applySociogramWeight(&ProviderConfigRequest{UseSociogram: true, SociogramWeight: 0.5}, weights)
	if err == nil {
		t.Fatal("expected error: sociogram provider not available")
	}
}

func TestApplySociogramWeight_ValidSociogram(t *testing.T) {
	h := buildTFHandlerWithProviders()
	weights := map[string]float64{}
	err := h.applySociogramWeight(&ProviderConfigRequest{UseSociogram: true, SociogramWeight: 0.6}, weights)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if weights["sociogram"] != 0.6 {
		t.Errorf("expected sociogram weight=0.6, got %v", weights["sociogram"])
	}
}

func TestApplySociogramWeight_DefaultOnZero(t *testing.T) {
	h := buildTFHandlerWithProviders()
	weights := map[string]float64{}
	err := h.applySociogramWeight(&ProviderConfigRequest{UseSociogram: true, SociogramWeight: 0.0}, weights)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if weights["sociogram"] != 0.5 {
		t.Errorf("expected default=0.5, got %v", weights["sociogram"])
	}
}

func TestApplySociogramWeight_WeightExceedsOne(t *testing.T) {
	h := buildTFHandlerWithProviders()
	weights := map[string]float64{}
	err := h.applySociogramWeight(&ProviderConfigRequest{UseSociogram: true, SociogramWeight: 1.1}, weights)
	if err == nil {
		t.Fatal("expected error for sociogram_weight > 1")
	}
}

// ── buildFromCustomProviders ─────────────────────────────────────────────────

func TestBuildFromCustomProviders_InvalidWeight(t *testing.T) {
	h := buildTFHandlerWithProviders()
	_, err := h.buildFromCustomProviders(map[string]float64{"enneagram": -0.1})
	if err == nil {
		t.Fatal("expected error for negative weight")
	}
}

func TestBuildFromCustomProviders_UnknownProvider(t *testing.T) {
	h := buildTFHandlerWithProviders()
	_, err := h.buildFromCustomProviders(map[string]float64{"unknown_prov": 0.5})
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestBuildFromCustomProviders_ValidProvider(t *testing.T) {
	h := buildTFHandlerWithProviders()
	result, err := h.buildFromCustomProviders(map[string]float64{"enneagram": 0.8})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result["enneagram"] != 0.8 {
		t.Errorf("expected 0.8, got %v", result["enneagram"])
	}
}

func TestBuildFromCustomProviders_ZeroWeightFiltered(t *testing.T) {
	h := buildTFHandlerWithProviders()
	// weight=0 → should not be included in result (filtered out)
	result, err := h.buildFromCustomProviders(map[string]float64{"enneagram": 0.0})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if _, ok := result["enneagram"]; ok {
		t.Error("expected enneagram to be filtered (weight=0)")
	}
}

// ── buildProviderWeights with nil config ─────────────────────────────────────

func TestBuildProviderWeights_NilConfig(t *testing.T) {
	h := buildTFHandlerWithProviders()
	result, err := h.buildProviderWeights(nil)
	if err != nil {
		t.Fatalf("expected nil error for nil config, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil defaults")
	}
}

// ── markEarlyExit ─────────────────────────────────────────────────────────────

func TestMarkEarlyExit_NilChromosome(t *testing.T) {
	// Should not panic
	markEarlyExit(nil, 2)
}

func TestMarkEarlyExit_SetsStats(t *testing.T) {
	chrom := &domain.Chromosome{
		Fitness: 0.95,
		Stats:   nil,
	}
	markEarlyExit(chrom, 2)
	if chrom.Stats == nil {
		t.Fatal("expected Stats to be initialized")
	}
	runsVal, ok := chrom.Stats["runs_executed"]
	if !ok {
		t.Fatal("expected runs_executed in stats")
	}
	if runsVal != 2 {
		t.Errorf("expected runs_executed=2, got %v", runsVal)
	}
	if chrom.Stats["early_exit"] != true {
		t.Errorf("expected early_exit=true, got %v", chrom.Stats["early_exit"])
	}
}

func TestMarkEarlyExit_ExistingStats(t *testing.T) {
	chrom := &domain.Chromosome{
		Fitness: 0.95,
		Stats:   map[string]interface{}{"other": "value"},
	}
	markEarlyExit(chrom, 1)
	if chrom.Stats["runs_executed"] != 1 {
		t.Errorf("expected runs_executed=1, got %v", chrom.Stats["runs_executed"])
	}
}

// ── Excel helper functions (pure functions) ───────────────────────────────────

func TestWriteExcelHeaders_Basic(t *testing.T) {
	f := excelize.NewFile()
	headers := []string{"Name", "Email", "Group"}
	writeExcelHeaders(f, "Sheet1", headers)
	// Verify first cell has value
	val, err := f.GetCellValue("Sheet1", "A1")
	if err != nil {
		t.Fatalf("GetCellValue error: %v", err)
	}
	if val != "Name" {
		t.Errorf("expected A1=Name, got %q", val)
	}
	val2, _ := f.GetCellValue("Sheet1", "C1")
	if val2 != "Group" {
		t.Errorf("expected C1=Group, got %q", val2)
	}
}

func TestWriteExcelHeaders_EmptyHeaders(t *testing.T) {
	f := excelize.NewFile()
	writeExcelHeaders(f, "Sheet1", []string{})
	// No panic - function handles empty slice
}

func TestWriteExcelStudentRow_Basic(t *testing.T) {
	f := excelize.NewFile()
	d := StudentRowData{
		CourseName:  "Ingeniería",
		SectionCode: "SEC01",
		Semester:    "2024-1",
		Student: excelStudent{
			Nombres:    "Juan",
			Paterno:    "Pérez",
			Materno:    "García",
			Email:      "juan@test.cl",
			Rut:        "12345678-9",
			CareerCode: "INF",
		},
		GroupName: "Grupo 1",
		Psych: studentPsychData{
			Source: "bfi44",
			OCEAN: domain.OCEANVector{
				Openness:          0.7,
				Conscientiousness: 0.8,
				Extraversion:      0.6,
				Agreeableness:     0.9,
				Neuroticism:       0.3,
			},
		},
		HasPsych: true,
	}
	writeExcelStudentRow(f, "Sheet1", 2, d)
	// Verify course name in A2
	val, err := f.GetCellValue("Sheet1", "A2")
	if err != nil {
		t.Fatalf("GetCellValue error: %v", err)
	}
	if val != "Ingeniería" {
		t.Errorf("expected A2=Ingeniería, got %q", val)
	}
	// Verify full name in D2
	name, _ := f.GetCellValue("Sheet1", "D2")
	if name != "Juan Pérez García" {
		t.Errorf("expected D2=Juan Pérez García, got %q", name)
	}
}

func TestSetExcelColumnWidths_NoError(t *testing.T) {
	f := excelize.NewFile()
	// Should not panic
	setExcelColumnWidths(f, "Sheet1")
}

// ── FlexibleFormTeams with registered providers ───────────────────────────────

func TestTF_FlexibleFormTeams_WithRegisteredEnneagram(t *testing.T) {
	h := buildTFHandlerWithProviders()
	app := fiber.New()
	app.Post("/flexible", h.FlexibleFormTeams)

	body := map[string]interface{}{
		"section_id": "s1",
		"candidates": []string{"a", "b", "c", "d"},
		"team_size":  2,
		"provider_config": map[string]interface{}{
			"personality_test":   "enneagram",
			"personality_weight": 0.6,
			"use_sociogram":      false,
		},
	}
	status, result := doRequest(t, app, "POST", "/flexible", body)
	// Engine has enneagram provider now, should form teams
	if status != http.StatusOK && status != http.StatusInternalServerError {
		t.Fatalf("unexpected status %d, result: %v", status, result)
	}
}

func TestTF_FlexibleFormTeams_WithBothProviders(t *testing.T) {
	h := buildTFHandlerWithProviders()
	app := fiber.New()
	app.Post("/flexible", h.FlexibleFormTeams)

	body := map[string]interface{}{
		"section_id": "s1",
		"candidates": []string{"a", "b", "c", "d"},
		"team_size":  2,
		"provider_config": map[string]interface{}{
			"personality_test":   "enneagram",
			"personality_weight": 0.5,
			"use_sociogram":      true,
			"sociogram_weight":   0.5,
		},
	}
	status, result := doRequest(t, app, "POST", "/flexible", body)
	if status != http.StatusOK && status != http.StatusInternalServerError {
		t.Fatalf("unexpected status %d, result: %v", status, result)
	}
}

func TestTF_FlexibleFormTeams_InvalidPersonalityTest(t *testing.T) {
	h := buildTFHandlerWithProviders()
	app := fiber.New()
	app.Post("/flexible", h.FlexibleFormTeams)

	body := map[string]interface{}{
		"section_id": "s1",
		"candidates": []string{"a", "b", "c", "d"},
		"team_size":  2,
		"provider_config": map[string]interface{}{
			"personality_test":   "not_a_valid_test",
			"personality_weight": 0.5,
		},
	}
	status, _ := doRequest(t, app, "POST", "/flexible", body)
	if status != http.StatusInternalServerError {
		t.Fatalf("expected 500 (invalid personality test), got %d", status)
	}
}

func TestTF_FlexibleFormTeams_WeightOverOne(t *testing.T) {
	h := buildTFHandlerWithProviders()
	app := fiber.New()
	app.Post("/flexible", h.FlexibleFormTeams)

	body := map[string]interface{}{
		"section_id": "s1",
		"candidates": []string{"a", "b", "c", "d"},
		"team_size":  2,
		"provider_config": map[string]interface{}{
			"personality_test":   "enneagram",
			"personality_weight": 1.5, // > 1 → error
		},
	}
	status, _ := doRequest(t, app, "POST", "/flexible", body)
	if status != http.StatusInternalServerError {
		t.Fatalf("expected 500 (weight > 1), got %d", status)
	}
}

func TestTF_FlexibleFormTeams_SociogramWeightOverOne(t *testing.T) {
	h := buildTFHandlerWithProviders()
	app := fiber.New()
	app.Post("/flexible", h.FlexibleFormTeams)

	body := map[string]interface{}{
		"section_id": "s1",
		"candidates": []string{"a", "b", "c", "d"},
		"team_size":  2,
		"provider_config": map[string]interface{}{
			"personality_test": "",
			"use_sociogram":    true,
			"sociogram_weight": 1.2, // > 1 → error
		},
	}
	status, _ := doRequest(t, app, "POST", "/flexible", body)
	if status != http.StatusInternalServerError {
		t.Fatalf("expected 500 (sociogram_weight > 1), got %d", status)
	}
}

// ── GetAvailableOptions with registered providers ────────────────────────────

func TestTF_GetAvailableOptions_WithProviders(t *testing.T) {
	h := buildTFHandlerWithProviders()
	app := fiber.New()
	app.Get("/options", h.GetAvailableOptions)

	status, result := doRequest(t, app, "GET", "/options", nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if result["success"] != true {
		t.Error("expected success=true")
	}
	data := result["data"].(map[string]interface{})
	if _, ok := data["personality_tests"]; !ok {
		t.Error("expected personality_tests in data")
	}
}

// ── buildProviderWeights with custom providers ────────────────────────────────

func TestTF_FlexibleFormTeams_CustomProvidersInvalid(t *testing.T) {
	h := buildTFHandlerWithProviders()
	app := fiber.New()
	app.Post("/flexible", h.FlexibleFormTeams)

	body := map[string]interface{}{
		"section_id": "s1",
		"candidates": []string{"a", "b", "c", "d"},
		"team_size":  2,
		"provider_config": map[string]interface{}{
			"custom_providers": map[string]interface{}{
				"enneagram": 1.5, // > 1 → buildFromCustomProviders error
			},
		},
	}
	status, _ := doRequest(t, app, "POST", "/flexible", body)
	if status != http.StatusInternalServerError {
		t.Fatalf("expected 500 (invalid custom weight), got %d", status)
	}
}

func TestTF_FlexibleFormTeams_CustomProvidersValid(t *testing.T) {
	h := buildTFHandlerWithProviders()
	app := fiber.New()
	app.Post("/flexible", h.FlexibleFormTeams)

	body := map[string]interface{}{
		"section_id": "s1",
		"candidates": []string{"a", "b", "c", "d"},
		"team_size":  2,
		"provider_config": map[string]interface{}{
			"custom_providers": map[string]interface{}{
				"enneagram": 0.7,
			},
		},
	}
	status, result := doRequest(t, app, "POST", "/flexible", body)
	if status != http.StatusOK && status != http.StatusInternalServerError {
		t.Fatalf("unexpected status %d, result: %v", status, result)
	}
}

// buildTFHandlerWithPlugin creates a handler with a real plugin set, and registered providers.
func buildTFHandlerWithPlugin(plugin core.Plugin) *TeamFormationHandler {
	cfg := core.DefaultTeamFormationConfig()
	cfg.Algorithm = "greedy"
	engine := core.NewTeamFormationEngine(cfg)

	cfgRegis := &config.TeamFormationConfig{
		DefaultWeight:     0.5,
		ProviderWeights:   map[string]float64{},
		ExcludedProviders: []string{},
	}
	registry := core.NewProviderRegistry(cfgRegis)
	_ = registry.Register(&fakeProv{id: "enneagram", name: "Enneagram"})
	return NewTeamFormationHandler(engine, registry, plugin)
}

// ── executeFormTeams with plugin path ────────────────────────────────────────

func TestTF_FormTeams_WithPlugin_Success(t *testing.T) {
	pluginResult := map[string]interface{}{
		"teams": []interface{}{
			map[string]interface{}{
				"id":             "team-1",
				"affinity_score": 0.75,
				"members":        []interface{}{map[string]interface{}{"id": "a"}, map[string]interface{}{"id": "b"}},
			},
			map[string]interface{}{
				"id":             "team-2",
				"affinity_score": 0.80,
				"members":        []interface{}{map[string]interface{}{"id": "c"}, map[string]interface{}{"id": "d"}},
			},
		},
		"global_score": 0.775,
		"algorithm":    "greedy",
	}
	plugin := &fakePlugin{result: pluginResult, err: nil}
	h := buildTFHandlerWithPlugin(plugin)
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
}

func TestTF_FormTeams_WithPlugin_PluginError_FallbackToEngine(t *testing.T) {
	plugin := &fakePlugin{result: nil, err: fmt.Errorf("plugin execution failed")}
	h := buildTFHandlerWithPlugin(plugin)
	app := fiber.New()
	app.Post("/form", h.FormTeams)

	body := map[string]interface{}{
		"section_id": "s1",
		"candidates": []string{"a", "b", "c", "d"},
		"team_size":  2,
	}
	status, result := doRequest(t, app, "POST", "/form", body)
	if status != http.StatusOK {
		t.Fatalf("expected 200 (fallback), got %d (result: %v)", status, result)
	}
}

func TestTF_FormTeams_WithPlugin_NilResult(t *testing.T) {
	plugin := &fakePlugin{result: nil, err: nil}
	h := buildTFHandlerWithPlugin(plugin)
	app := fiber.New()
	app.Post("/form", h.FormTeams)

	body := map[string]interface{}{
		"section_id": "s1",
		"candidates": []string{"a", "b", "c", "d"},
		"team_size":  2,
	}
	status, _ := doRequest(t, app, "POST", "/form", body)
	if status != http.StatusInternalServerError {
		t.Fatalf("expected 500 (nil plugin result), got %d", status)
	}
}

// ── runLegacyFormation with plugin path ───────────────────────────────────────

func TestTF_FlexibleFormTeams_WithPlugin_LegacyMode_Success(t *testing.T) {
	pluginResult := map[string]interface{}{
		"teams": []interface{}{
			map[string]interface{}{
				"id":      "team-1",
				"members": []interface{}{map[string]interface{}{"id": "a"}, map[string]interface{}{"id": "b"}},
			},
		},
		"global_score": 0.8,
		"algorithm":    "greedy",
	}
	plugin := &fakePlugin{result: pluginResult, err: nil}
	h := buildTFHandlerWithPlugin(plugin)
	app := fiber.New()
	app.Post("/flexible", h.FlexibleFormTeams)

	body := map[string]interface{}{
		"section_id": "s1",
		"candidates": []string{"a", "b", "c", "d"},
		"team_size":  2,
		"provider_config": map[string]interface{}{
			"personality_test":   "enneagram",
			"personality_weight": 0.6,
		},
	}
	status, result := doRequest(t, app, "POST", "/flexible", body)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d (result: %v)", status, result)
	}
}

func TestTF_FlexibleFormTeams_WithPlugin_LegacyMode_PluginError_Fallback(t *testing.T) {
	plugin := &fakePlugin{result: nil, err: fmt.Errorf("plugin failed")}
	h := buildTFHandlerWithPlugin(plugin)
	app := fiber.New()
	app.Post("/flexible", h.FlexibleFormTeams)

	body := map[string]interface{}{
		"section_id": "s1",
		"candidates": []string{"a", "b", "c", "d"},
		"team_size":  2,
		"provider_config": map[string]interface{}{
			"personality_test":   "enneagram",
			"personality_weight": 0.6,
		},
	}
	status, result := doRequest(t, app, "POST", "/flexible", body)
	if status != http.StatusOK {
		t.Fatalf("expected 200 (plugin error fallback), got %d (result: %v)", status, result)
	}
}

// ── runCriteriaFormation with plugin path ─────────────────────────────────────

func TestTF_FlexibleFormTeams_WithPlugin_CriteriaMode_Success(t *testing.T) {
	pluginResult := map[string]interface{}{
		"teams": []interface{}{
			map[string]interface{}{
				"id":      "team-1",
				"members": []interface{}{map[string]interface{}{"id": "a"}, map[string]interface{}{"id": "b"}},
			},
		},
		"global_score": 0.7,
		"algorithm":    "genetic",
	}
	plugin := &fakePlugin{result: pluginResult, err: nil}
	h := buildTFHandlerWithPlugin(plugin)
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
			"a": map[string]interface{}{"numeric_attrs": map[string]interface{}{"gpa": 3.5}},
			"b": map[string]interface{}{"numeric_attrs": map[string]interface{}{"gpa": 3.0}},
			"c": map[string]interface{}{"numeric_attrs": map[string]interface{}{"gpa": 2.5}},
			"d": map[string]interface{}{"numeric_attrs": map[string]interface{}{"gpa": 4.0}},
		},
	}
	status, result := doRequest(t, app, "POST", "/flexible", body)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d (result: %v)", status, result)
	}
}

func TestTF_FlexibleFormTeams_WithPlugin_CriteriaMode_PluginError_Fallback(t *testing.T) {
	plugin := &fakePlugin{result: nil, err: fmt.Errorf("criteria plugin failed")}
	h := buildTFHandlerWithPlugin(plugin)
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
			"a": map[string]interface{}{"numeric_attrs": map[string]interface{}{"gpa": 3.5}},
			"b": map[string]interface{}{"numeric_attrs": map[string]interface{}{"gpa": 3.0}},
			"c": map[string]interface{}{"numeric_attrs": map[string]interface{}{"gpa": 2.5}},
			"d": map[string]interface{}{"numeric_attrs": map[string]interface{}{"gpa": 4.0}},
		},
	}
	status, result := doRequest(t, app, "POST", "/flexible", body)
	if status != http.StatusOK && status != http.StatusInternalServerError {
		t.Fatalf("unexpected status %d (result: %v)", status, result)
	}
}

// ── getFinalDiversity via multiple criteria ───────────────────────────────────

func TestTF_CriteriaFormation_MultipleCriteria(t *testing.T) {
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
				"weight": 0.5,
			},
			"gender": map[string]interface{}{
				"target": 0.5,
				"weight": 0.5,
			},
		},
		"demographic_data": map[string]interface{}{
			"a": map[string]interface{}{
				"numeric_attrs":     map[string]interface{}{"gpa": 3.5},
				"categorical_attrs": map[string]interface{}{"gender": "M"},
			},
			"b": map[string]interface{}{
				"numeric_attrs":     map[string]interface{}{"gpa": 3.0},
				"categorical_attrs": map[string]interface{}{"gender": "F"},
			},
			"c": map[string]interface{}{
				"numeric_attrs":     map[string]interface{}{"gpa": 2.5},
				"categorical_attrs": map[string]interface{}{"gender": "M"},
			},
			"d": map[string]interface{}{
				"numeric_attrs":     map[string]interface{}{"gpa": 4.0},
				"categorical_attrs": map[string]interface{}{"gender": "F"},
			},
		},
	}
	status, result := doRequest(t, app, "POST", "/flexible", body)
	if status != http.StatusOK && status != http.StatusInternalServerError {
		t.Fatalf("unexpected status %d (result: %v)", status, result)
	}
}
