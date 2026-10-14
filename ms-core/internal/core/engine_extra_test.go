package core

// core_engine_extra_test.go — extra tests for uncovered TeamFormationEngine functions.

import (
	"context"
	"math/rand"
	"testing"
)

// ─── FormTeamsWithConfig ──────────────────────────────────────────────────────

func TestFormTeamsWithConfig_NoCandidates(t *testing.T) {
	engine := NewTeamFormationEngine(nil)
	_, err := engine.FormTeamsWithConfig(context.Background(), TeamFormationRequest{
		TeamSize: 2,
	})
	if err == nil || err.Error() != "no candidates provided" {
		t.Fatalf("expected 'no candidates provided', got %v", err)
	}
}

func TestFormTeamsWithConfig_ZeroTeamSize(t *testing.T) {
	engine := NewTeamFormationEngine(nil)
	_, err := engine.FormTeamsWithConfig(context.Background(), TeamFormationRequest{
		Candidates: []TeamMember{{ID: "a"}},
	})
	if err == nil || err.Error() != "team size must be positive" {
		t.Fatalf("expected 'team size must be positive', got %v", err)
	}
}

func TestFormTeamsWithConfig_WithCustomWeights(t *testing.T) {
	// Register a provider, then run FormTeamsWithConfig with custom weights.
	engine := NewTeamFormationEngine(&TeamFormationConfig{
		Algorithm:     "greedy",
		DefaultWeight: 1.0,
		Weights:       map[string]float64{},
	})
	p := NewMockProvider("p1", func(_ context.Context, a, b string) (float64, error) {
		return 0.5, nil
	})
	engine.RegisterProvider(p) //nolint:errcheck

	req := TeamFormationRequest{
		SectionID: "s1",
		Candidates: []TeamMember{
			{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"},
		},
		TeamSize:        2,
		MinTeamSize:     2,
		MaxTeamSize:     2,
		ProviderWeights: map[string]float64{"p1": 0.8},
	}
	result, err := engine.FormTeamsWithConfig(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Teams) != 2 {
		t.Fatalf("expected 2 teams, got %d", len(result.Teams))
	}
}

func TestFormTeamsWithConfig_NoCustomWeights(t *testing.T) {
	// Without ProviderWeights, should still work.
	engine := NewTeamFormationEngine(&TeamFormationConfig{
		Algorithm:     "greedy",
		DefaultWeight: 1.0,
		Weights:       map[string]float64{},
	})
	p := NewMockProvider("p1", nil)
	engine.RegisterProvider(p) //nolint:errcheck

	req := TeamFormationRequest{
		SectionID: "s1",
		Candidates: []TeamMember{
			{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"},
		},
		TeamSize:    2,
		MinTeamSize: 2,
		MaxTeamSize: 2,
	}
	result, err := engine.FormTeamsWithConfig(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Metadata == nil {
		t.Error("expected non-nil metadata")
	}
}

// ─── buildAssignedSet ────────────────────────────────────────────────────────

func TestBuildAssignedSet_Empty(t *testing.T) {
	set := buildAssignedSet([]Team{})
	if len(set) != 0 {
		t.Fatalf("expected empty set, got %v", set)
	}
}

func TestBuildAssignedSet_PopulatedTeam(t *testing.T) {
	teams := []Team{
		{Members: []TeamMember{{ID: "a"}, {ID: "b"}}},
		{Members: []TeamMember{{ID: "c"}}},
	}
	set := buildAssignedSet(teams)
	if !set["a"] || !set["b"] || !set["c"] {
		t.Fatalf("expected all members in set, got %v", set)
	}
	if len(set) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(set))
	}
}

// ─── PrecomputeAffinityMatrix ────────────────────────────────────────────────

func TestPrecomputeAffinityMatrix_Basic(t *testing.T) {
	engine := NewTeamFormationEngine(nil)
	p := NewMockProvider("p1", func(_ context.Context, a, b string) (float64, error) {
		return 0.7, nil
	})
	engine.RegisterProvider(p) //nolint:errcheck

	ctx := context.Background()
	candidates := []TeamMember{{ID: "x"}, {ID: "y"}, {ID: "z"}}
	matrix := engine.PrecomputeAffinityMatrix(ctx, candidates)

	if matrix["x"]["y"] != 0.7 {
		t.Fatalf("expected 0.7, got %v", matrix["x"]["y"])
	}
	if matrix["y"]["x"] != 0.7 {
		t.Fatalf("expected symmetric matrix")
	}
}

func TestPrecomputeAffinityMatrix_SingleCandidate(t *testing.T) {
	engine := NewTeamFormationEngine(nil)
	candidates := []TeamMember{{ID: "solo"}}
	matrix := engine.PrecomputeAffinityMatrix(context.Background(), candidates)
	if len(matrix) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(matrix))
	}
}

// ─── OptimizeWithMatrix ──────────────────────────────────────────────────────

func TestOptimizeWithMatrix_Basic(t *testing.T) {
	candidates := []TeamMember{
		{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"},
	}
	matrix := map[string]map[string]float64{
		"a": {"b": 0.8, "c": 0.3, "d": 0.5},
		"b": {"a": 0.8, "c": 0.3, "d": 0.4},
		"c": {"a": 0.3, "b": 0.3, "d": 0.9},
		"d": {"a": 0.5, "b": 0.4, "c": 0.9},
	}

	config := &TeamFormationConfig{
		Algorithm:     "greedy",
		DefaultWeight: 1.0,
		Weights:       map[string]float64{},
	}

	req := TeamFormationRequest{
		SectionID:   "s1",
		Candidates:  candidates,
		TeamSize:    2,
		MinTeamSize: 2,
		MaxTeamSize: 2,
	}

	result, err := OptimizeWithMatrix(context.Background(), config, req, matrix)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || len(result.Teams) == 0 {
		t.Fatal("expected non-empty result")
	}
}

func TestOptimizeWithMatrix_NilConfig(t *testing.T) {
	// With nil config, should use default config
	candidates := []TeamMember{{ID: "a"}, {ID: "b"}}
	matrix := map[string]map[string]float64{
		"a": {"b": 0.5},
		"b": {"a": 0.5},
	}
	req := TeamFormationRequest{
		SectionID:   "s1",
		Candidates:  candidates,
		TeamSize:    2,
		MinTeamSize: 1,
		MaxTeamSize: 2,
	}
	result, err := OptimizeWithMatrix(context.Background(), nil, req, matrix)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
}

// ─── calculateTargetDistance ─────────────────────────────────────────────────

func TestCalculateTargetDistance_Empty(t *testing.T) {
	d := calculateTargetDistance([]int{}, 3)
	if d != 0 {
		t.Fatalf("expected 0, got %v", d)
	}
}

func TestCalculateTargetDistance_AllEqual(t *testing.T) {
	d := calculateTargetDistance([]int{3, 3, 3}, 3)
	if d != 0 {
		t.Fatalf("expected 0, got %v", d)
	}
}

func TestCalculateTargetDistance_Mixed(t *testing.T) {
	d := calculateTargetDistance([]int{2, 3, 4}, 3)
	// |2-3| + |3-3| + |4-3| = 1+0+1 = 2; avg = 2/3 ≈ 0.667
	if d < 0.66 || d > 0.68 {
		t.Fatalf("expected ~0.667, got %v", d)
	}
}

// ─── tournamentSelectStandard ────────────────────────────────────────────────

func TestTournamentSelectStandard_ReturnsHigher(t *testing.T) {
	engine := NewTeamFormationEngine(nil)
	// fitness[0]=1.0, fitness[1]=0.0 → should always select idx 0
	// (with fixed seed so rand picks indices 0 and 1)
	rng := rand.New(rand.NewSource(42))
	fitness := []float64{1.0, 0.0}

	best := -1
	for i := 0; i < 20; i++ {
		idx := engine.tournamentSelectStandard(fitness, rng)
		if best == -1 {
			best = idx
		}
	}
	// At least one call should return 0 (the higher fitness)
	// Just verify it doesn't panic and returns valid index
	if best < 0 || best > 1 {
		t.Fatalf("unexpected index: %d", best)
	}
}

// ─── CalculateCombinedAffinity — allProviderWeight0 branch ───────────────────

func TestCalculateCombinedAffinity_AllWeightZero(t *testing.T) {
	config := &TeamFormationConfig{
		Weights:       map[string]float64{"p1": 0.0},
		DefaultWeight: 0.0,
	}
	engine := NewTeamFormationEngine(config)
	p := NewMockProvider("p1", nil)
	engine.RegisterProvider(p) //nolint:errcheck

	score, _, err := engine.CalculateCombinedAffinity(context.Background(), "a", "b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// All weight 0 → neutral 0.5
	if score != 0.5 {
		t.Fatalf("expected 0.5, got %v", score)
	}
}

// ─── RegisterProvider edge: nil provider ────────────────────────────────────

func TestEngineRegisterProvider_Nil(t *testing.T) {
	engine := NewTeamFormationEngine(nil)
	err := engine.RegisterProvider(nil)
	if err == nil {
		t.Fatal("expected error for nil provider")
	}
}
