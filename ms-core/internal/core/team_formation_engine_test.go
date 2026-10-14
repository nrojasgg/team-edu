package core

import (
	"context"
	"testing"
)

// MockAffinityProvider implementa AffinityProvider para tests
type MockAffinityProvider struct {
	id           string
	name         string
	affinityFunc func(ctx context.Context, a, b string) (float64, error)
	supportsCtx  bool
}

func NewMockProvider(id string, affinityFunc func(ctx context.Context, a, b string) (float64, error)) *MockAffinityProvider {
	return &MockAffinityProvider{
		id:           id,
		name:         "Mock " + id,
		affinityFunc: affinityFunc,
		supportsCtx:  true,
	}
}

func (m *MockAffinityProvider) ProviderID() string                       { return m.id }
func (m *MockAffinityProvider) ProviderName() string                     { return m.name }
func (m *MockAffinityProvider) SupportsContext(ctx context.Context) bool { return m.supportsCtx }
func (m *MockAffinityProvider) GetMetadata() ProviderMetadata {
	return ProviderMetadata{
		ID:       m.id,
		Name:     m.name,
		Category: "test",
		Version:  "1.0.0",
	}
}
func (m *MockAffinityProvider) CalculatePairwiseAffinity(ctx context.Context, personA, personB string) (float64, error) {
	if m.affinityFunc != nil {
		return m.affinityFunc(ctx, personA, personB)
	}
	return 0.5, nil
}

func TestTeamFormationEngine_RegisterProvider(t *testing.T) {
	engine := NewTeamFormationEngine(nil)

	provider := NewMockProvider("test", nil)
	err := engine.RegisterProvider(provider)
	if err != nil {
		t.Fatalf("Failed to register provider: %v", err)
	}

	// Intentar registrar el mismo provider de nuevo
	err = engine.RegisterProvider(provider)
	if err == nil {
		t.Error("Expected error when registering duplicate provider")
	}
}

func TestTeamFormationEngine_CalculateCombinedAffinity(t *testing.T) {
	config := &TeamFormationConfig{
		Weights: map[string]float64{
			"provider_a": 0.7,
			"provider_b": 0.3,
		},
		DefaultWeight: 1.0,
	}
	engine := NewTeamFormationEngine(config)

	// Provider A siempre retorna 1.0 (máxima afinidad)
	providerA := NewMockProvider("provider_a", func(ctx context.Context, a, b string) (float64, error) {
		return 1.0, nil
	})

	// Provider B siempre retorna -1.0 (máximo rechazo)
	providerB := NewMockProvider("provider_b", func(ctx context.Context, a, b string) (float64, error) {
		return -1.0, nil
	})

	engine.RegisterProvider(providerA)
	engine.RegisterProvider(providerB)

	ctx := context.WithValue(context.Background(), ContextKeySectionID, "test-section")

	score, providerScores, err := engine.CalculateCombinedAffinity(ctx, "user1", "user2")
	if err != nil {
		t.Fatalf("Failed to calculate affinity: %v", err)
	}

	// Score esperado: (1.0 * 0.7 + -1.0 * 0.3) / (0.7 + 0.3) = 0.4
	expectedScore := 0.4
	tolerance := 0.001
	if score < expectedScore-tolerance || score > expectedScore+tolerance {
		t.Errorf("Expected score %.2f, got %.2f", expectedScore, score)
	}

	t.Logf("Combined score: %.2f", score)

	if providerScores["provider_a"] != 1.0 {
		t.Errorf("Expected provider_a score 1.0, got %.2f", providerScores["provider_a"])
	}

	if providerScores["provider_b"] != -1.0 {
		t.Errorf("Expected provider_b score -1.0, got %.2f", providerScores["provider_b"])
	}
}

func TestTeamFormationEngine_FormTeamsGreedy(t *testing.T) {
	config := &TeamFormationConfig{
		Algorithm:     "greedy",
		DefaultWeight: 1.0,
		Weights: map[string]float64{
			"test": 1.0,
		},
	}
	engine := NewTeamFormationEngine(config)

	// Provider que da alta afinidad a pares específicos
	affinityMap := map[string]map[string]float64{
		"user1": {"user2": 0.9, "user3": 0.1, "user4": 0.2, "user5": 0.3, "user6": 0.1},
		"user2": {"user1": 0.9, "user3": 0.2, "user4": 0.1, "user5": 0.1, "user6": 0.2},
		"user3": {"user1": 0.1, "user2": 0.2, "user4": 0.8, "user5": 0.2, "user6": 0.3},
		"user4": {"user1": 0.2, "user2": 0.1, "user3": 0.8, "user5": 0.3, "user6": 0.2},
		"user5": {"user1": 0.3, "user2": 0.1, "user3": 0.2, "user4": 0.3, "user6": 0.7},
		"user6": {"user1": 0.1, "user2": 0.2, "user3": 0.3, "user4": 0.2, "user5": 0.7},
	}

	provider := NewMockProvider("test", func(ctx context.Context, a, b string) (float64, error) {
		if scores, ok := affinityMap[a]; ok {
			if score, ok := scores[b]; ok {
				return score, nil
			}
		}
		return 0, nil
	})

	engine.RegisterProvider(provider)

	ctx := context.WithValue(context.Background(), ContextKeySectionID, "test-section")

	request := TeamFormationRequest{
		SectionID: "test-section",
		Candidates: []TeamMember{
			{ID: "user1"},
			{ID: "user2"},
			{ID: "user3"},
			{ID: "user4"},
			{ID: "user5"},
			{ID: "user6"},
		},
		TeamSize:    2,
		MinTeamSize: 2,
		MaxTeamSize: 2,
	}

	result, err := engine.FormTeams(ctx, request)
	if err != nil {
		t.Fatalf("Failed to form teams: %v", err)
	}

	// Debería formar 3 equipos de 2
	if len(result.Teams) != 3 {
		t.Errorf("Expected 3 teams, got %d", len(result.Teams))
	}

	for _, team := range result.Teams {
		if len(team.Members) != 2 {
			t.Errorf("Expected team size 2, got %d", len(team.Members))
		}
	}

	if result.Algorithm != "greedy" {
		t.Errorf("Expected algorithm 'greedy', got '%s'", result.Algorithm)
	}

	t.Logf("Teams formed: %+v", result.Teams)
	t.Logf("Global score: %.2f", result.GlobalScore)
	t.Logf("Execution time: %dms", result.ExecutionTimeMs)
}

func TestTeamFormationEngine_FormTeamsGenetic(t *testing.T) {
	config := &TeamFormationConfig{
		Algorithm:     "genetic",
		DefaultWeight: 1.0,
		Weights: map[string]float64{
			"test": 1.0,
		},
		GeneticParams: &GeneticAlgorithmParams{
			PopulationSize: 20,
			Generations:    10,
			MutationRate:   0.1,
			ElitismRate:    0.1,
		},
	}
	engine := NewTeamFormationEngine(config)

	// Provider simple para test
	provider := NewMockProvider("test", func(ctx context.Context, a, b string) (float64, error) {
		// Afinidad basada en "cercanía" de IDs
		if a < b {
			return 0.8, nil
		}
		return 0.3, nil
	})

	engine.RegisterProvider(provider)

	ctx := context.WithValue(context.Background(), ContextKeySectionID, "test-section")

	request := TeamFormationRequest{
		SectionID: "test-section",
		Candidates: []TeamMember{
			{ID: "user1"}, {ID: "user2"}, {ID: "user3"}, {ID: "user4"},
		},
		TeamSize:    2,
		MinTeamSize: 2,
		MaxTeamSize: 2,
	}

	result, err := engine.FormTeams(ctx, request)
	if err != nil {
		t.Fatalf("Failed to form teams: %v", err)
	}

	if len(result.Teams) != 2 {
		t.Errorf("Expected 2 teams, got %d", len(result.Teams))
	}

	if result.Algorithm != "nsga-2" {
		t.Errorf("Expected algorithm 'nsga-2', got '%s'", result.Algorithm)
	}

	t.Logf("Genetic algorithm result: %+v", result)
}

func TestTeamFormationEngine_UpdateWeights(t *testing.T) {
	config := &TeamFormationConfig{
		Weights: map[string]float64{
			"provider_a": 0.5,
			"provider_b": 0.5,
		},
		DefaultWeight: 1.0,
	}
	engine := NewTeamFormationEngine(config)

	providerA := NewMockProvider("provider_a", func(ctx context.Context, a, b string) (float64, error) {
		return 1.0, nil
	})
	providerB := NewMockProvider("provider_b", func(ctx context.Context, a, b string) (float64, error) {
		return 0.0, nil
	})

	engine.RegisterProvider(providerA)
	engine.RegisterProvider(providerB)

	ctx := context.WithValue(context.Background(), ContextKeySectionID, "test-section")

	// Con pesos iguales: (1.0 * 0.5 + 0.0 * 0.5) / 1.0 = 0.5
	score1, _, _ := engine.CalculateCombinedAffinity(ctx, "a", "b")
	if score1 != 0.5 {
		t.Errorf("Initial score should be 0.5, got %.2f", score1)
	}

	// Actualizar pesos: provider_a = 0.9, provider_b = 0.1
	engine.UpdateWeights(map[string]float64{
		"provider_a": 0.9,
		"provider_b": 0.1,
	})

	// Nuevo score: (1.0 * 0.9 + 0.0 * 0.1) / 1.0 = 0.9
	score2, _, _ := engine.CalculateCombinedAffinity(ctx, "a", "b")
	if score2 != 0.9 {
		t.Errorf("Updated score should be 0.9, got %.2f", score2)
	}
}

func TestDefaultAlgorithmIsGenetic(t *testing.T) {
	config := DefaultTeamFormationConfig()
	if config.Algorithm != "genetic" {
		t.Errorf("Expected default algorithm to be 'genetic', got '%s'", config.Algorithm)
	}

	if config.GeneticParams == nil {
		t.Error("Expected genetic parameters to be initialized by default")
	}
}

func TestGenerateDistributions(t *testing.T) {
	// Test: 5 candidatos, tamaño deseado 3, min 1, max 4
	// Distribuciones esperadas: [3,2], [2,2,1], [4,1], [3,1,1], [2,1,1,1], [1,1,1,1,1]
	distributions := generateDistributions(5, 3, 1, 4)

	t.Logf("Distributions for 5 candidates (target=3, min=1, max=4):")
	for i, dist := range distributions {
		sum := 0
		for _, s := range dist {
			sum += s
		}
		t.Logf("  [%d] %v (sum=%d)", i, dist, sum)
	}

	// Verificar que todas las distribuciones suman 5
	for _, dist := range distributions {
		sum := 0
		for _, s := range dist {
			sum += s
		}
		if sum != 5 {
			t.Errorf("Distribution %v does not sum to 5 (got %d)", dist, sum)
		}
	}

	// Verificar que hay al menos las distribuciones básicas
	if len(distributions) < 2 {
		t.Errorf("Expected at least 2 distributions, got %d", len(distributions))
	}

	// Debe incluir [3,2]
	found32 := false
	for _, dist := range distributions {
		if len(dist) == 2 && dist[0] == 3 && dist[1] == 2 {
			found32 = true
			break
		}
	}
	if !found32 {
		t.Error("Expected to find distribution [3,2]")
	}

	// Debe incluir [2,2,1]
	found221 := false
	for _, dist := range distributions {
		if len(dist) == 3 && dist[0] == 2 && dist[1] == 2 && dist[2] == 1 {
			found221 = true
			break
		}
	}
	if !found221 {
		t.Error("Expected to find distribution [2,2,1]")
	}
}

func TestTeamFormationWithIncompatibleEneatypes(t *testing.T) {
	config := &TeamFormationConfig{
		Algorithm:     "genetic",
		DefaultWeight: 1.0,
		Weights: map[string]float64{
			"test": 1.0,
		},
		GeneticParams: &GeneticAlgorithmParams{
			PopulationSize: 50,
			Generations:    30,
			MutationRate:   0.1,
			ElitismRate:    0.1,
		},
	}
	engine := NewTeamFormationEngine(config)

	// Simular compatibilidad de eneagrama:
	// A-B: compatible (1.0)
	// A-C: incompatible (0.0)
	// A-D: compatible (1.0)
	// A-E: compatible (1.0)
	// B-C: compatible (1.0)
	// B-D: incompatible (0.0)
	// B-E: compatible (1.0)
	// C-D: compatible (1.0)
	// C-E: incompatible (0.0)
	// D-E: compatible (1.0)
	affinityMap := map[string]map[string]float64{
		"A": {"B": 1.0, "C": 0.0, "D": 1.0, "E": 1.0},
		"B": {"A": 1.0, "C": 1.0, "D": 0.0, "E": 1.0},
		"C": {"A": 0.0, "B": 1.0, "D": 1.0, "E": 0.0},
		"D": {"A": 1.0, "B": 0.0, "C": 1.0, "E": 1.0},
		"E": {"A": 1.0, "B": 1.0, "C": 0.0, "D": 1.0},
	}

	provider := NewMockProvider("test", func(ctx context.Context, a, b string) (float64, error) {
		if scores, ok := affinityMap[a]; ok {
			if score, ok := scores[b]; ok {
				return score, nil
			}
		}
		if scores, ok := affinityMap[b]; ok {
			if score, ok := scores[a]; ok {
				return score, nil
			}
		}
		return 0.5, nil
	})

	engine.RegisterProvider(provider)

	ctx := context.WithValue(context.Background(), ContextKeySectionID, "test-section")

	request := TeamFormationRequest{
		SectionID: "test-section",
		Candidates: []TeamMember{
			{ID: "A"},
			{ID: "B"},
			{ID: "C"},
			{ID: "D"},
			{ID: "E"},
		},
		TeamSize:    3,
		MinTeamSize: 1,
		MaxTeamSize: 4,
	}

	result, err := engine.FormTeams(ctx, request)
	if err != nil {
		t.Fatalf("Failed to form teams: %v", err)
	}

	t.Logf("Result algorithm: %s", result.Algorithm)
	t.Logf("Global score: %.4f", result.GlobalScore)
	t.Logf("Number of teams: %d", len(result.Teams))
	t.Logf("Unassigned members: %d", len(result.UnassignedMembers))

	// Log cada equipo
	totalAssigned := 0
	for i, team := range result.Teams {
		memberIDs := make([]string, len(team.Members))
		for j, m := range team.Members {
			memberIDs[j] = m.ID
		}
		t.Logf("Team %d: %v (score: %.4f)", i+1, memberIDs, team.AffinityScore)
		totalAssigned += len(team.Members)

		// Verificar compatibilidades en el equipo
		for x := 0; x < len(team.Members); x++ {
			for y := x + 1; y < len(team.Members); y++ {
				a, b := team.Members[x].ID, team.Members[y].ID
				compat := affinityMap[a][b]
				if compat == 0 {
					compat = affinityMap[b][a]
				}
				t.Logf("  Pair (%s, %s): %.1f", a, b, compat)
			}
		}
	}

	// VERIFICACIÓN CRÍTICA: Todos los candidatos deben estar asignados
	if len(result.UnassignedMembers) > 0 {
		unassignedIDs := make([]string, len(result.UnassignedMembers))
		for i, m := range result.UnassignedMembers {
			unassignedIDs[i] = m.ID
		}
		t.Errorf("CRITICAL: %d members were left unassigned: %v", len(result.UnassignedMembers), unassignedIDs)
	}

	if totalAssigned != 5 {
		t.Errorf("Expected 5 members assigned, got %d", totalAssigned)
	}

	// Log metadata si existe
	if result.Metadata != nil {
		if dist, ok := result.Metadata["chosen_distribution"]; ok {
			t.Logf("Chosen distribution: %v", dist)
		}
	}
}
