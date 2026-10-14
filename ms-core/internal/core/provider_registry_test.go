package core

import (
	"context"
	"testing"

	"ms-core/internal/config"
	"ms-core/internal/core/domain"
)

// ---------------------------------------------------------------------------
// minimal mock AffinityProvider
// ---------------------------------------------------------------------------

type mockAffinityProvider struct {
	id       string
	name     string
	category string
}

func (m *mockAffinityProvider) ProviderID() string   { return m.id }
func (m *mockAffinityProvider) ProviderName() string { return m.name }
func (m *mockAffinityProvider) CalculatePairwiseAffinity(_ context.Context, _, _ string) (float64, error) {
	return 0.5, nil
}
func (m *mockAffinityProvider) SupportsContext(_ context.Context) bool { return true }
func (m *mockAffinityProvider) GetMetadata() ProviderMetadata {
	return ProviderMetadata{ID: m.id, Name: m.name, Category: m.category}
}

func makeRegistry(excluded ...string) *ProviderRegistry {
	cfg := &config.TeamFormationConfig{
		DefaultWeight:     0.5,
		ProviderWeights:   map[string]float64{},
		ExcludedProviders: excluded,
	}
	return NewProviderRegistry(cfg)
}

// ---------------------------------------------------------------------------
// ProviderRegistry.Register
// ---------------------------------------------------------------------------

func TestProviderRegistry_Register_Success(t *testing.T) {
	r := makeRegistry()
	p := &mockAffinityProvider{id: "enneagram", name: "Enneagram", category: "personality"}
	if err := r.Register(p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProviderRegistry_Register_Nil(t *testing.T) {
	r := makeRegistry()
	if err := r.Register(nil); err == nil {
		t.Error("expected error for nil provider")
	}
}

func TestProviderRegistry_Register_EmptyID(t *testing.T) {
	r := makeRegistry()
	p := &mockAffinityProvider{id: ""}
	if err := r.Register(p); err == nil {
		t.Error("expected error for empty provider ID")
	}
}

func TestProviderRegistry_Register_Duplicate(t *testing.T) {
	r := makeRegistry()
	p := &mockAffinityProvider{id: "enneagram"}
	r.Register(p) //nolint:errcheck
	if err := r.Register(p); err == nil {
		t.Error("expected error for duplicate provider")
	}
}

// ---------------------------------------------------------------------------
// ProviderRegistry.Get
// ---------------------------------------------------------------------------

func TestProviderRegistry_Get_Found(t *testing.T) {
	r := makeRegistry()
	p := &mockAffinityProvider{id: "enneagram"}
	r.Register(p) //nolint:errcheck
	got, ok := r.Get("enneagram")
	if !ok || got == nil {
		t.Error("expected to find registered provider")
	}
}

func TestProviderRegistry_Get_NotFound(t *testing.T) {
	r := makeRegistry()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("expected not found for unregistered provider")
	}
}

// ---------------------------------------------------------------------------
// ProviderRegistry.HasProvider
// ---------------------------------------------------------------------------

func TestProviderRegistry_HasProvider_True(t *testing.T) {
	r := makeRegistry()
	r.Register(&mockAffinityProvider{id: "mbti"}) //nolint:errcheck
	if !r.HasProvider("mbti") {
		t.Error("expected HasProvider=true")
	}
}

func TestProviderRegistry_HasProvider_False(t *testing.T) {
	r := makeRegistry()
	if r.HasProvider("sociogram") {
		t.Error("expected HasProvider=false for unregistered provider")
	}
}

// ---------------------------------------------------------------------------
// ProviderRegistry.IsExcluded
// ---------------------------------------------------------------------------

func TestProviderRegistry_IsExcluded_Excluded(t *testing.T) {
	r := makeRegistry("peer_evaluation")
	if !r.IsExcluded("peer_evaluation") {
		t.Error("expected IsExcluded=true for 'peer_evaluation'")
	}
}

func TestProviderRegistry_IsExcluded_NotExcluded(t *testing.T) {
	r := makeRegistry()
	if r.IsExcluded("enneagram") {
		t.Error("expected IsExcluded=false")
	}
}

func TestProviderRegistry_IsExcluded_NilConfig(t *testing.T) {
	r := NewProviderRegistry(nil)
	if r.IsExcluded("anything") {
		t.Error("expected false when config is nil")
	}
}

// ---------------------------------------------------------------------------
// ProviderRegistry.GetEligibleProviders
// ---------------------------------------------------------------------------

func TestProviderRegistry_GetEligibleProviders(t *testing.T) {
	r := makeRegistry("peer_evaluation")
	r.Register(&mockAffinityProvider{id: "enneagram"})       //nolint:errcheck
	r.Register(&mockAffinityProvider{id: "peer_evaluation"}) //nolint:errcheck

	eligible := r.GetEligibleProviders()
	for _, p := range eligible {
		if p.ProviderID() == "peer_evaluation" {
			t.Error("peer_evaluation should be excluded from eligible providers")
		}
	}
	if len(eligible) != 1 {
		t.Errorf("expected 1 eligible provider, got %d", len(eligible))
	}
}

// ---------------------------------------------------------------------------
// ProviderRegistry.GetAllProviders
// ---------------------------------------------------------------------------

func TestProviderRegistry_GetAllProviders(t *testing.T) {
	r := makeRegistry()
	r.Register(&mockAffinityProvider{id: "a"}) //nolint:errcheck
	r.Register(&mockAffinityProvider{id: "b"}) //nolint:errcheck

	all := r.GetAllProviders()
	if len(all) != 2 {
		t.Errorf("expected 2 providers, got %d", len(all))
	}
}

// ---------------------------------------------------------------------------
// ProviderRegistry.GetProviderIDs
// ---------------------------------------------------------------------------

func TestProviderRegistry_GetProviderIDs(t *testing.T) {
	r := makeRegistry()
	r.Register(&mockAffinityProvider{id: "x"}) //nolint:errcheck
	ids := r.GetProviderIDs()
	if len(ids) != 1 || ids[0] != "x" {
		t.Errorf("expected ['x'], got %v", ids)
	}
}

// ---------------------------------------------------------------------------
// ProviderRegistry.GetWeightedEligibleProviders
// ---------------------------------------------------------------------------

func TestProviderRegistry_GetWeightedEligibleProviders(t *testing.T) {
	r := makeRegistry("peer_evaluation")
	r.Register(&mockAffinityProvider{id: "enneagram"})       //nolint:errcheck
	r.Register(&mockAffinityProvider{id: "peer_evaluation"}) //nolint:errcheck

	weighted := r.GetWeightedEligibleProviders()
	if len(weighted) != 1 {
		t.Errorf("expected 1 weighted eligible provider, got %d", len(weighted))
	}
	if weighted[0].Provider.ProviderID() != "enneagram" {
		t.Errorf("expected enneagram provider, got %s", weighted[0].Provider.ProviderID())
	}
}

// ---------------------------------------------------------------------------
// ProviderRegistry.Summary
// ---------------------------------------------------------------------------

func TestProviderRegistry_Summary_Keys(t *testing.T) {
	r := makeRegistry()
	r.Register(&mockAffinityProvider{id: "enneagram", name: "Ennea", category: "personality"}) //nolint:errcheck

	summary := r.Summary()

	if _, ok := summary["total_providers"]; !ok {
		t.Error("expected 'total_providers' key in summary")
	}
	if _, ok := summary["eligible_providers"]; !ok {
		t.Error("expected 'eligible_providers' key")
	}
	if summary["total_providers"].(int) != 1 {
		t.Errorf("expected total_providers=1, got %v", summary["total_providers"])
	}
}

// ---------------------------------------------------------------------------
// TeamFormationEngine.EvaluateSolution
// ---------------------------------------------------------------------------

func TestEvaluateSolution_NoOCEAN_NoSocial(t *testing.T) {
	engine := NewTeamFormationEngine(nil)
	team := []TeamMember{
		{ID: "A"},
		{ID: "B"},
	}
	result := engine.EvaluateSolution(team, nil)
	// No OCEAN → psyScore default 0.5; no adjacency → soc 0
	if result.PsychometricScore != 0.5 {
		t.Errorf("expected PsychometricScore=0.5, got %f", result.PsychometricScore)
	}
	if !result.IsFeasible {
		t.Error("expected feasible with no veto")
	}
}

func TestEvaluateSolution_WithVeto(t *testing.T) {
	engine := NewTeamFormationEngine(nil)
	team := []TeamMember{
		{ID: "A"},
		{ID: "B"},
	}
	// B vetoes A (strong negative)
	adjacency := map[string]map[string]float64{
		"B": {"A": -3.0},
	}
	result := engine.EvaluateSolution(team, adjacency)
	if result.IsFeasible {
		t.Error("expected infeasible due to veto")
	}
	if result.SociometricScore > -500 {
		t.Errorf("expected sociometric score <= -500 (veto), got %f", result.SociometricScore)
	}
}

func TestEvaluateSolution_WithOCEAN(t *testing.T) {
	engine := NewTeamFormationEngine(nil)
	ocean := &domain.OCEANVector{
		Openness: 0.7, Conscientiousness: 0.8, Extraversion: 0.5,
		Agreeableness: 0.9, Neuroticism: 0.2,
	}
	team := []TeamMember{
		{ID: "A", Psychometrics: ocean},
		{ID: "B", Psychometrics: ocean},
	}
	result := engine.EvaluateSolution(team, nil)
	if result.PsychometricScore == 0.5 {
		t.Error("expected non-default psychometric score when OCEAN data is present")
	}
	if result.PsychometricScore < 0 || result.PsychometricScore > 1 {
		t.Errorf("expected score in [0,1], got %f", result.PsychometricScore)
	}
}

// ---------------------------------------------------------------------------
// TeamFormationEngine.UpdateWeights (extra case)
// ---------------------------------------------------------------------------

func TestTeamFormationEngine_UpdateWeights_Extra(t *testing.T) {
	engine := NewTeamFormationEngine(nil)
	engine.UpdateWeights(map[string]float64{"enneagram": 0.7, "sociogram": 0.3})
	// Verify via GetProviderInfo (just ensure no panic)
	_ = engine
}

// ---------------------------------------------------------------------------
// TeamFormationEngine.GetProviderInfo
// ---------------------------------------------------------------------------

func TestTeamFormationEngine_GetProviderInfo(t *testing.T) {
	engine := NewTeamFormationEngine(nil)
	info := engine.GetProviderInfo()
	// info is []WeightedProvider; initially empty
	if len(info) != 0 {
		t.Errorf("expected empty provider info for fresh engine, got %d", len(info))
	}
}
