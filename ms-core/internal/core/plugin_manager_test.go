package core

import (
	"context"
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// mock Plugin implementation for tests
// ---------------------------------------------------------------------------

type mockPlugin struct {
	name    string
	version string
	execFn  func(ctx context.Context, data interface{}) (interface{}, error)
	initErr error
}

func (m *mockPlugin) Name() string    { return m.name }
func (m *mockPlugin) Version() string { return m.version }
func (m *mockPlugin) Initialize() error {
	if m.initErr != nil {
		return m.initErr
	}
	return nil
}
func (m *mockPlugin) Execute(ctx context.Context, data interface{}) (interface{}, error) {
	if m.execFn != nil {
		return m.execFn(ctx, data)
	}
	return map[string]interface{}{"ok": true}, nil
}
func (m *mockPlugin) Shutdown() error { return nil }

func newMock(name string) *mockPlugin {
	return &mockPlugin{name: name, version: "1.0.0"}
}

// ---------------------------------------------------------------------------
// PluginManager tests
// ---------------------------------------------------------------------------

func TestPluginManager_RegisterAndExecute(t *testing.T) {
	pm := NewPluginManager()
	p := &mockPlugin{
		name:    "test_plugin",
		version: "1.0.0",
		execFn: func(_ context.Context, data interface{}) (interface{}, error) {
			return "hello", nil
		},
	}
	if err := pm.RegisterPlugin(p); err != nil {
		t.Fatalf("RegisterPlugin: %v", err)
	}

	result, err := pm.ExecutePlugin("test_plugin", nil)
	if err != nil {
		t.Fatalf("ExecutePlugin: %v", err)
	}
	if result != "hello" {
		t.Errorf("expected 'hello', got %v", result)
	}
}

func TestPluginManager_RegisterNil(t *testing.T) {
	pm := NewPluginManager()
	if err := pm.RegisterPlugin(nil); err == nil {
		t.Error("expected error when registering nil plugin")
	}
}

func TestPluginManager_RegisterEmptyName(t *testing.T) {
	pm := NewPluginManager()
	p := &mockPlugin{name: "", version: "1.0.0"}
	if err := pm.RegisterPlugin(p); err == nil {
		t.Error("expected error when registering plugin with empty name")
	}
}

func TestPluginManager_RegisterDuplicate(t *testing.T) {
	pm := NewPluginManager()
	if err := pm.RegisterPlugin(newMock("dup")); err != nil {
		t.Fatalf("first register: %v", err)
	}
	if err := pm.RegisterPlugin(newMock("dup")); err == nil {
		t.Error("expected error on duplicate registration")
	}
}

func TestPluginManager_RegisterInitError(t *testing.T) {
	pm := NewPluginManager()
	p := &mockPlugin{name: "bad_init", version: "1.0.0", initErr: errors.New("init failed")}
	if err := pm.RegisterPlugin(p); err == nil {
		t.Error("expected error when Initialize() fails")
	}
}

func TestPluginManager_GetPluginNotFound(t *testing.T) {
	pm := NewPluginManager()
	_, err := pm.GetPlugin("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent plugin")
	}
}

func TestPluginManager_ExecutePluginNotFound(t *testing.T) {
	pm := NewPluginManager()
	_, err := pm.ExecutePlugin("nonexistent", nil)
	if err == nil {
		t.Error("expected error for nonexistent plugin")
	}
}

func TestPluginManager_EnableDisable(t *testing.T) {
	pm := NewPluginManager()
	if err := pm.RegisterPlugin(newMock("toggle")); err != nil {
		t.Fatalf("register: %v", err)
	}

	// starts enabled
	enabled, err := pm.IsEnabled("toggle")
	if err != nil || !enabled {
		t.Error("plugin should start enabled")
	}

	// disable
	if err := pm.DisablePlugin("toggle"); err != nil {
		t.Fatalf("disable: %v", err)
	}
	enabled, _ = pm.IsEnabled("toggle")
	if enabled {
		t.Error("plugin should be disabled")
	}

	// execute on disabled plugin should return ErrPluginDisabled
	_, execErr := pm.ExecutePlugin("toggle", nil)
	if !errors.Is(execErr, ErrPluginDisabled) {
		t.Errorf("expected ErrPluginDisabled, got %v", execErr)
	}

	// re-enable
	if err := pm.EnablePlugin("toggle"); err != nil {
		t.Fatalf("enable: %v", err)
	}
	enabled, _ = pm.IsEnabled("toggle")
	if !enabled {
		t.Error("plugin should be enabled again")
	}

	// execute works again
	if _, err := pm.ExecutePlugin("toggle", nil); err != nil {
		t.Errorf("execute after re-enable: %v", err)
	}
}

func TestPluginManager_DisableAlreadyDisabled(t *testing.T) {
	pm := NewPluginManager()
	if err := pm.RegisterPlugin(newMock("idempotent")); err != nil {
		t.Fatal(err)
	}
	pm.DisablePlugin("idempotent") //nolint
	// Calling disable again should not error
	if err := pm.DisablePlugin("idempotent"); err != nil {
		t.Errorf("expected no error on double-disable, got %v", err)
	}
}

func TestPluginManager_EnableAlreadyEnabled(t *testing.T) {
	pm := NewPluginManager()
	if err := pm.RegisterPlugin(newMock("always_on")); err != nil {
		t.Fatal(err)
	}
	// Already enabled; calling again should not error
	if err := pm.EnablePlugin("always_on"); err != nil {
		t.Errorf("expected no error on double-enable, got %v", err)
	}
}

func TestPluginManager_EnableNonExistent(t *testing.T) {
	pm := NewPluginManager()
	if err := pm.EnablePlugin("ghost"); err == nil {
		t.Error("expected error enabling nonexistent plugin")
	}
}

func TestPluginManager_DisableNonExistent(t *testing.T) {
	pm := NewPluginManager()
	if err := pm.DisablePlugin("ghost"); err == nil {
		t.Error("expected error disabling nonexistent plugin")
	}
}

func TestPluginManager_IsEnabledNonExistent(t *testing.T) {
	pm := NewPluginManager()
	_, err := pm.IsEnabled("ghost")
	if err == nil {
		t.Error("expected error for nonexistent plugin")
	}
}

func TestPluginManager_GetLoadedPlugins(t *testing.T) {
	pm := NewPluginManager()
	pm.RegisterPlugin(newMock("alpha")) //nolint
	pm.RegisterPlugin(newMock("beta"))  //nolint
	pm.DisablePlugin("beta")            //nolint

	plugins := pm.GetLoadedPlugins()
	if len(plugins) != 2 {
		t.Errorf("expected 2 plugins, got %d", len(plugins))
	}

	byName := make(map[string]PluginInfo)
	for _, pi := range plugins {
		byName[pi.Name] = pi
	}
	if !byName["alpha"].Enabled {
		t.Error("alpha should be enabled")
	}
	if byName["beta"].Enabled {
		t.Error("beta should be disabled")
	}
}

func TestPluginManager_GetLoadedPluginNames(t *testing.T) {
	pm := NewPluginManager()
	pm.RegisterPlugin(newMock("x")) //nolint
	pm.RegisterPlugin(newMock("y")) //nolint

	names := pm.GetLoadedPluginNames()
	if len(names) != 2 {
		t.Errorf("expected 2 names, got %d", len(names))
	}
}

func TestPluginManager_Shutdown(t *testing.T) {
	pm := NewPluginManager()
	pm.RegisterPlugin(newMock("shutdown_test")) //nolint
	if err := pm.Shutdown(); err != nil {
		t.Errorf("Shutdown returned error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// FitnessResult.CombinedScore tests
// ---------------------------------------------------------------------------

func TestFitnessResult_CombinedScore_Feasible(t *testing.T) {
	f := FitnessResult{
		PsychometricScore: 0.8,
		SociometricScore:  10.0,
		IsFeasible:        true,
	}
	score := f.CombinedScore(0.5, 0.5)
	if score <= 0 {
		t.Errorf("expected positive combined score, got %f", score)
	}
}

func TestFitnessResult_CombinedScore_Infeasible(t *testing.T) {
	f := FitnessResult{
		PsychometricScore: 0.9,
		SociometricScore:  -1000.0,
		IsFeasible:        false,
	}
	score := f.CombinedScore(0.5, 0.5)
	if score != -1000.0 {
		t.Errorf("expected -1000 for infeasible, got %f", score)
	}
}

func TestFitnessResult_CombinedScore_ZeroWeights(t *testing.T) {
	f := FitnessResult{
		PsychometricScore: 0.5,
		SociometricScore:  5.0,
		IsFeasible:        true,
	}
	score := f.CombinedScore(0, 0)
	if score != 0.0 {
		t.Errorf("expected 0 with zero weights, got %f", score)
	}
}

// ---------------------------------------------------------------------------
// CalculateSociometricFitnessNormalized tests
// ---------------------------------------------------------------------------

func TestCalculateSociometricFitnessNormalized_PerfectCohesion(t *testing.T) {
	team := []TeamMember{{ID: "a"}, {ID: "b"}, {ID: "c"}}
	adj := map[string]map[string]float64{
		"a": {"b": 3, "c": 3},
		"b": {"a": 3, "c": 3},
		"c": {"a": 3, "b": 3},
	}
	score := CalculateSociometricFitnessNormalized(team, adj)
	if score < 0.5 || score > 1.0 {
		t.Errorf("expected [0.5, 1.0] for perfect cohesion, got %f", score)
	}
}

func TestCalculateSociometricFitnessNormalized_VetoPenalty(t *testing.T) {
	team := []TeamMember{{ID: "a"}, {ID: "b"}}
	adj := map[string]map[string]float64{
		"a": {"b": -1},
	}
	score := CalculateSociometricFitnessNormalized(team, adj)
	if score != 0.0 {
		t.Errorf("expected 0 for veto, got %f", score)
	}
}

func TestCalculateSociometricFitnessNormalized_SingleMember(t *testing.T) {
	team := []TeamMember{{ID: "a"}}
	score := CalculateSociometricFitnessNormalized(team, nil)
	if score != 0.5 {
		t.Errorf("expected 0.5 for single member, got %f", score)
	}
}

func TestCalculateTeamSociometricFitness(t *testing.T) {
	team := []TeamMember{
		{ID: "a", Sociometry: map[string]float64{"b": 2.0}},
		{ID: "b", Sociometry: map[string]float64{"a": 2.0}},
	}
	score := CalculateTeamSociometricFitness(team)
	// Two mutual votes: 2 + 2 = 4, no isolates
	if score != 4.0 {
		t.Errorf("expected 4.0, got %f", score)
	}
}
