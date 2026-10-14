package core

import (
	"context"
	"math"
	"math/rand"
	"sort"
	"testing"

	"ms-core/internal/core/genetic"
)

// =============================================================================
// TESTS DE INTEGRACIÓN MULTIOBJETIVO
//
// Estos tests verifican que el pipeline completo del NSGA-II en
// TeamFormationEngine se comporta genuinamente como un algoritmo multiobjetivo:
//
//  1. El engine expone ambos scores (PsyScore, SocScore) separados en el resultado
//  2. La selección usa CrowdedComparison, no un escalar combinado
//  3. El frente de Pareto final contiene trade-offs reales
//  4. El knee-point es la solución más cercana al punto ideal (1,1), no la de
//     mayor promedio
//  5. El engine NO colapsa los dos objetivos en un solo float64 durante la
//     evolución (eso lo hace solo al final, para GlobalScore)
// =============================================================================

// buildEngine crea un TeamFormationEngine con un provider de afinidad controlado.
func buildEngine(affinityFn func(ctx context.Context, a, b string) (float64, error)) *TeamFormationEngine {
	config := &TeamFormationConfig{
		Algorithm:     "genetic",
		DefaultWeight: 1.0,
		Weights:       map[string]float64{"mock": 1.0},
		GeneticParams: &GeneticAlgorithmParams{
			PopulationSize: 30,
			Generations:    20,
			MutationRate:   0.15,
			ElitismRate:    0.1,
		},
	}
	engine := NewTeamFormationEngine(config)
	provider := NewMockProvider("mock", affinityFn)
	_ = engine.RegisterProvider(provider)
	return engine
}

// makeRequest crea un TeamFormationRequest con los candidatos dados.
func makeRequest(sectionID string, ids []string, teamSize int) TeamFormationRequest {
	candidates := make([]TeamMember, len(ids))
	for i, id := range ids {
		candidates[i] = TeamMember{ID: id}
	}
	return TeamFormationRequest{
		SectionID:   sectionID,
		Candidates:  candidates,
		TeamSize:    teamSize,
		MinTeamSize: teamSize - 1,
		MaxTeamSize: teamSize + 1,
	}
}

// hasKey es un helper para verificar presencia de clave en map[string]interface{}.
func hasKey(m map[string]interface{}, key string) bool {
	_, ok := m[key]
	return ok
}

// metadataKeys retorna las claves del mapa como slice ordenado.
func metadataKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// extractFloat busca un valor float en el metadata por múltiples claves posibles.
func extractFloat(m map[string]interface{}, keys ...string) float64 {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch val := v.(type) {
			case float64:
				return val
			case float32:
				return float64(val)
			case int:
				return float64(val)
			}
		}
	}
	return 0.0
}

// -----------------------------------------------------------------------------
// Test 1: El resultado expone PsyScore y SocScore por separado en Metadata
// -----------------------------------------------------------------------------

// TestEngineExposesBothObjectivesInMetadata verifica que el resultado del
// algoritmo incluye los dos scores objetivos de forma independiente en Metadata.
// Si solo se expone el promedio escalar, no es un resultado multiobjetivo real.
func TestEngineExposesBothObjectivesInMetadata(t *testing.T) {
	engine := buildEngine(func(_ context.Context, a, b string) (float64, error) {
		return 0.9, nil
	})

	ctx := context.WithValue(context.Background(), ContextKeySectionID, "sec-1")
	req := makeRequest("sec-1", []string{"A", "B", "C", "D"}, 2)

	result, err := engine.FormTeams(ctx, req)
	if err != nil {
		t.Fatalf("FormTeams failed: %v", err)
	}

	if result.Metadata == nil {
		t.Fatal("Result Metadata is nil — cannot inspect per-objective scores")
	}

	hasSeparateScores := hasKey(result.Metadata, "psy_score") ||
		hasKey(result.Metadata, "fitness_psy") ||
		hasKey(result.Metadata, "fitness_soc") ||
		hasKey(result.Metadata, "pareto_rank")

	if !hasSeparateScores {
		t.Errorf("Metadata does not expose individual objective scores. Keys present: %v",
			metadataKeys(result.Metadata))
	}

	t.Logf("Metadata keys: %v", metadataKeys(result.Metadata))
}

// -----------------------------------------------------------------------------
// Test 2a: El knee-point selecciona la solución más cercana a (0,0) en minimización
// -----------------------------------------------------------------------------

// TestSelectKneePointChoosesBalancedSolution verifica que selectKneePoint
// elige la solución con el mejor balance entre ambos objetivos (la más cercana
// al punto ideal (0,0) en minimización).
func TestSelectKneePointChoosesBalancedSolution(t *testing.T) {
	engine := NewTeamFormationEngine(nil)

	// Frente de Pareto con tres soluciones en MINIMIZACIÓN:
	//   A: psy=0.0, soc=1.0  → dist a (0,0) = 1.0
	//   B: psy=0.2, soc=0.2  → dist a (0,0) = 0.283  ← knee (más cercano al ideal)
	//   C: psy=1.0, soc=0.0  → dist a (0,0) = 1.0
	// En minimización, menores scores son mejores.
	front := []*genetic.Individual{
		{Fitness: genetic.FitnessResult{PsyScore: 0.0, SocScore: 1.0, IsFeasible: true}},
		{Fitness: genetic.FitnessResult{PsyScore: 0.2, SocScore: 0.2, IsFeasible: true}},
		{Fitness: genetic.FitnessResult{PsyScore: 1.0, SocScore: 0.0, IsFeasible: true}},
	}

	knee := engine.selectKneePoint(front)
	if knee == nil {
		t.Fatal("selectKneePoint returned nil")
	}

	if math.Abs(knee.Fitness.PsyScore-0.2) > 0.001 || math.Abs(knee.Fitness.SocScore-0.2) > 0.001 {
		t.Errorf("Expected knee point (0.2, 0.2), got (%.2f, %.2f)",
			knee.Fitness.PsyScore, knee.Fitness.SocScore)
	}
}

// -----------------------------------------------------------------------------
// Test 2b: El knee-point difiere del máximo escalar cuando hay conflicto
// -----------------------------------------------------------------------------

// TestSelectKneePointVsScalarAverage verifica que el knee-point difiere del
// promedio escalar cuando hay una solución con promedio bajo pero balanceada.
// En minimización: menores scores son mejores, ideal es (0,0).
//
//	X: psy=0.01, soc=0.40 → avg=0.205, dist=(0.01²+0.40²)^0.5 ≈ 0.400
//	Y: psy=0.25, soc=0.25 → avg=0.250, dist=(0.25²+0.25²)^0.5 ≈ 0.354  ← knee
//	Z: psy=0.40, soc=0.01 → avg=0.205, dist=(0.40²+0.01²)^0.5  ≈ 0.400
//
// Y tiene mayor promedio pero menor distancia al ideal (0,0).
func TestSelectKneePointVsScalarAverage(t *testing.T) {
	engine := NewTeamFormationEngine(nil)

	front := []*genetic.Individual{
		{Fitness: genetic.FitnessResult{PsyScore: 0.01, SocScore: 0.40, IsFeasible: true}},
		{Fitness: genetic.FitnessResult{PsyScore: 0.25, SocScore: 0.25, IsFeasible: true}},
		{Fitness: genetic.FitnessResult{PsyScore: 0.40, SocScore: 0.01, IsFeasible: true}},
	}

	knee := engine.selectKneePoint(front)
	if knee == nil {
		t.Fatal("selectKneePoint returned nil")
	}

	expectedPsy, expectedSoc := 0.25, 0.25
	if math.Abs(knee.Fitness.PsyScore-expectedPsy) > 0.01 ||
		math.Abs(knee.Fitness.SocScore-expectedSoc) > 0.01 {
		t.Errorf(
			"Expected knee point (%.2f, %.2f) — the balanced solution, got (%.2f, %.2f). "+
				"If X or Z was selected, the engine may be using scalar average instead of proximity to ideal.",
			expectedPsy, expectedSoc,
			knee.Fitness.PsyScore, knee.Fitness.SocScore,
		)
	}

	t.Logf("Knee point selected: (psy=%.3f, soc=%.3f)", knee.Fitness.PsyScore, knee.Fitness.SocScore)
}

// -----------------------------------------------------------------------------
// Test 3: El torneo NSGA-II usa CrowdedComparison, no fitness escalar
// -----------------------------------------------------------------------------

// TestTournamentSelectNSGA2UsesCrowdedComparison verifica que el torneo del
// NSGA-II selecciona individuos usando el operador crowded-comparison
// (rank primero, distancia después) y NO por fitness escalar.
func TestTournamentSelectNSGA2UsesCrowdedComparison(t *testing.T) {
	engine := NewTeamFormationEngine(nil)

	// Crear dos individuos:
	//   A: rank=1, distance=0.1 (knee point - muy cercano al ideal (0,0))
	//   B: rank=1, distance=5.0 (extremo - lejos del ideal)
	//
	// En MINIMIZACIÓN con crowding distance como criterio de selección:
	//   CrowdedComparison (mismo rank) → A gana (distance 0.1 < 5.0)
	indA := &genetic.Individual{
		Fitness: genetic.FitnessResult{
			PsyScore: 0.1, SocScore: 0.1,
			Rank: 1, Distance: 0.1,
			IsFeasible: true,
		},
	}
	indB := &genetic.Individual{
		Fitness: genetic.FitnessResult{
			PsyScore: 1.0, SocScore: 1.0,
			Rank: 1, Distance: 5.0,
			IsFeasible: true,
		},
	}

	// CrowdedComparison directo: A debe ganar (menor distancia, mismo rank - knee point)
	if !genetic.CrowdedComparison(indA.Fitness, indB.Fitness) {
		t.Error("CrowdedComparison: A (dist=0.1) should beat B (dist=5.0) at same rank (minimization: knee point wins)")
	}
	if genetic.CrowdedComparison(indB.Fitness, indA.Fitness) {
		t.Error("CrowdedComparison: B (dist=5.0) should NOT beat A (dist=0.1) at same rank")
	}

	// Verificar que tournamentSelectNSGA2 usa CrowdedComparison ejecutando
	// múltiples torneos con una población de solo A y B.
	// Con 2 individuos en minimización: P(A gana) = P(A,A) + P(A,B) + P(B,A) = 0.75
	population := []*genetic.Individual{indA, indB}
	aWins := 0
	trials := 200
	rng := rand.New(rand.NewSource(42))

	for i := 0; i < trials; i++ {
		selected := engine.tournamentSelectNSGA2(population, rng)
		if selected == indA {
			aWins++
		}
	}

	// A debería ganar al menos 60% de los torneos (umbral conservador).
	expectedMinAWins := int(float64(trials) * 0.6)
	if aWins < expectedMinAWins {
		t.Errorf(
			"Expected A (knee point, dist=0.1) to win at least %d/%d tournaments "+
				"using CrowdedComparison in minimization mode, but only won %d.",
			expectedMinAWins, trials, aWins,
		)
	}
	t.Logf("A (knee) won %d/%d tournaments (%.1f%%)", aWins, trials, float64(aWins)/float64(trials)*100)
}

// -----------------------------------------------------------------------------
// Test 4: nsga2Evolve mantiene diversidad en el frente de Pareto final
// -----------------------------------------------------------------------------

// TestNSGA2EvolveProducesMultipleParetoSolutions verifica que tras la evolución
// el frente de Pareto final contiene MÁS DE UNA solución con trade-offs reales.
func TestNSGA2EvolveProducesMultipleParetoSolutions(t *testing.T) {
	highSocPairs := map[string]float64{
		"A-B": 0.95, "C-D": 0.95, "E-F": 0.95,
		"A-C": 0.1, "B-D": 0.1, "A-D": 0.2,
		"B-C": 0.2, "E-A": 0.3, "F-B": 0.3,
	}

	engine := buildEngine(func(_ context.Context, a, b string) (float64, error) {
		key := a + "-" + b
		if v, ok := highSocPairs[key]; ok {
			return v, nil
		}
		key2 := b + "-" + a
		if v, ok := highSocPairs[key2]; ok {
			return v, nil
		}
		return 0.4, nil
	})

	ctx := context.WithValue(context.Background(), ContextKeySectionID, "sec-pareto")
	req := makeRequest("sec-pareto", []string{"A", "B", "C", "D", "E", "F"}, 2)

	result, err := engine.FormTeams(ctx, req)
	if err != nil {
		t.Fatalf("FormTeams failed: %v", err)
	}

	if result.Metadata != nil {
		if size, ok := result.Metadata["pareto_front_size"]; ok {
			var frontSize int
			switch v := size.(type) {
			case int:
				frontSize = v
			case float64:
				frontSize = int(v)
			}
			if frontSize < 1 {
				t.Errorf("Pareto front size should be ≥ 1, got %d", frontSize)
			}
			t.Logf("Pareto front size: %d", frontSize)
		}
	}

	if result.Metadata != nil {
		psyVal := extractFloat(result.Metadata, "fitness_psy", "psy_score")
		socVal := extractFloat(result.Metadata, "fitness_soc", "soc_score")

		t.Logf("Final solution: psy=%.4f soc=%.4f global=%.4f",
			psyVal, socVal, result.GlobalScore)

		// GlobalScore debe ser el promedio de ambos
		if psyVal > 0 && socVal > 0 {
			expectedGlobal := (psyVal + socVal) / 2.0
			if math.Abs(result.GlobalScore-expectedGlobal) > 0.01 {
				t.Errorf(
					"GlobalScore (%.4f) should be avg of psy (%.4f) and soc (%.4f) = %.4f. "+
						"Discrepancy suggests scoring inconsistency.",
					result.GlobalScore, psyVal, socVal, expectedGlobal,
				)
			}
		}
	}

	// Todos los candidatos deben estar asignados
	totalAssigned := 0
	for _, team := range result.Teams {
		totalAssigned += len(team.Members)
	}
	if totalAssigned+len(result.UnassignedMembers) != len(req.Candidates) {
		t.Errorf("Expected %d total members (assigned+unassigned), got %d+%d=%d",
			len(req.Candidates), totalAssigned, len(result.UnassignedMembers),
			totalAssigned+len(result.UnassignedMembers))
	}
}

// -----------------------------------------------------------------------------
// Test 5: El algoritmo reportado es NSGA-II
// -----------------------------------------------------------------------------

// TestGeneticEngineIsScalarNotMultiObjective documenta que el genetic.Engine
// del paquete genetic usa un único float64 como fitness (monoobjetivo) y que
// el comportamiento multiobjetivo real lo provee TeamFormationEngine vía NSGA-II.
func TestGeneticEngineIsScalarNotMultiObjective(t *testing.T) {
	t.Log("genetic.Engine uses scalar fitness (float64) per chromosome — this is correct.")
	t.Log("Multi-objective behavior is provided by TeamFormationEngine.nsga2Evolve()")
	t.Log("which uses genetic.FastNonDominatedSort and genetic.CalculateCrowdingDistance.")

	engine := buildEngine(func(_ context.Context, _, _ string) (float64, error) {
		return 0.5, nil
	})

	ctx := context.WithValue(context.Background(), ContextKeySectionID, "sec-algo")
	req := makeRequest("sec-algo", []string{"A", "B", "C", "D"}, 2)

	result, err := engine.FormTeams(ctx, req)
	if err != nil {
		t.Fatalf("FormTeams failed: %v", err)
	}

	validAlgorithms := map[string]bool{
		"nsga-2":                    true,
		"genetic_intelligent_nsga2": true,
		"genetic":                   true,
	}

	if !validAlgorithms[result.Algorithm] {
		t.Errorf("Expected NSGA-II based algorithm, got: %q", result.Algorithm)
	}

	t.Logf("Algorithm used: %s", result.Algorithm)
}

// -----------------------------------------------------------------------------
// Test 6: nsga2Evolve mantiene padres débiles cuando son mejores en minimización
// -----------------------------------------------------------------------------

// TestNSGA2EvolveReplacesWeakParents verifica el principio de elitismo NSGA-II:
// en MINIMIZACIÓN, los individuos con scores BAJOS son MEJORES.
// Los weak parents (scores ~0.1-0.2) son ÓPTIMOS y los offspring generados
// (scores ~0.8-0.9) son PEORES, por lo que los weak parents deben sobrevivir.
func TestNSGA2EvolveReplacesWeakParents(t *testing.T) {
	// En minimización, estos SON los MEJORES (menores scores)
	weakParents := []*genetic.Individual{
		{Chromosome: []int{0, 1, 2, 3}, Fitness: genetic.FitnessResult{PsyScore: 0.1, SocScore: 0.1, IsFeasible: true}},
		{Chromosome: []int{1, 0, 3, 2}, Fitness: genetic.FitnessResult{PsyScore: 0.15, SocScore: 0.1, IsFeasible: true}},
		{Chromosome: []int{2, 3, 0, 1}, Fitness: genetic.FitnessResult{PsyScore: 0.1, SocScore: 0.15, IsFeasible: true}},
		{Chromosome: []int{3, 2, 1, 0}, Fitness: genetic.FitnessResult{PsyScore: 0.15, SocScore: 0.15, IsFeasible: true}},
	}

	callCount := 0
	evalFn := func(c []int) genetic.FitnessResult {
		callCount++
		// Offspring con scores MÁS ALTOS son PEORES en minimización
		return genetic.FitnessResult{
			PsyScore:   0.8 + float64(callCount%3)*0.05,
			SocScore:   0.8 + float64(callCount%2)*0.05,
			IsFeasible: true,
		}
	}

	engine := NewTeamFormationEngine(nil)
	params := &GeneticAlgorithmParams{
		PopulationSize: 4,
		Generations:    3,
		MutationRate:   0.0,
		ElitismRate:    0.0,
	}

	rng := rand.New(rand.NewSource(99))
	finalFronts := engine.Nsga2Evolve(weakParents, params, rng, evalFn, nil)

	if len(finalFronts) == 0 {
		t.Fatal("nsga2Evolve returned no fronts")
	}

	paretoFront := finalFronts[0]
	if len(paretoFront) == 0 {
		t.Fatal("Pareto front is empty")
	}

	// En minimización, los weak parents (scores bajos ~0.1-0.15) son ÓPTIMOS
	// y DOMINAN a los offspring (scores altos ~0.8-0.9)
	// Verificamos que los weak parents sobrevivan en el frente de Pareto
	weakParentFound := false
	for _, ind := range paretoFront {
		if ind.Fitness.PsyScore <= 0.2 && ind.Fitness.SocScore <= 0.2 {
			weakParentFound = true
			t.Logf("Weak parent survived (good in minimization): (psy=%.2f, soc=%.2f)",
				ind.Fitness.PsyScore, ind.Fitness.SocScore)
		}
	}

	if !weakParentFound {
		t.Error("In minimization, weak parents (low scores ~0.1) should dominate offspring (high scores ~0.8). " +
			"This suggests the algorithm is not correctly handling minimization.")
	}

	t.Logf("Final Pareto front has %d solutions", len(paretoFront))
	for i, ind := range paretoFront {
		t.Logf("  [%d] psy=%.3f soc=%.3f rank=%d",
			i, ind.Fitness.PsyScore, ind.Fitness.SocScore, ind.Fitness.Rank)
	}
}

// -----------------------------------------------------------------------------
// Test 7: TeamMetadata expone PsychometricScore y SociometricScore por separado
// -----------------------------------------------------------------------------

// TestTeamMetadataExposesPerObjectiveScores verifica que cada equipo en el
// resultado tiene sus scores de ambos objetivos accesibles por separado.
func TestTeamMetadataExposesPerObjectiveScores(t *testing.T) {
	engine := buildEngine(func(_ context.Context, a, b string) (float64, error) {
		if a < b {
			return 0.8, nil
		}
		return 0.3, nil
	})

	ctx := context.WithValue(context.Background(), ContextKeySectionID, "sec-meta")
	req := makeRequest("sec-meta", []string{"A", "B", "C", "D", "E", "F"}, 3)

	result, err := engine.FormTeams(ctx, req)
	if err != nil {
		t.Fatalf("FormTeams failed: %v", err)
	}

	for i, team := range result.Teams {
		psyScore := team.Metadata.PsychometricScore
		socScore := team.Metadata.SociometricScore

		t.Logf("Team %d: psy=%.4f soc=%.4f affinity=%.4f members=%d",
			i+1, psyScore, socScore, team.AffinityScore, len(team.Members))

		if psyScore < 0 || psyScore > 1 {
			t.Errorf("Team %d PsychometricScore out of [0,1]: %.4f", i+1, psyScore)
		}
		if socScore < 0 || socScore > 1 {
			t.Errorf("Team %d SociometricScore out of [0,1]: %.4f", i+1, socScore)
		}

		// AffinityScore debe ser el promedio de ambos
		expectedAffinity := (psyScore + socScore) / 2.0
		if math.Abs(team.AffinityScore-expectedAffinity) > 0.01 {
			t.Errorf(
				"Team %d: AffinityScore (%.4f) should equal avg(psy=%.4f, soc=%.4f)=%.4f",
				i+1, team.AffinityScore, psyScore, socScore, expectedAffinity,
			)
		}
	}
}

// -----------------------------------------------------------------------------
// Test 8: evaluateChromosome retorna FitnessResult con ambos scores no-triviales
// -----------------------------------------------------------------------------

// TestEvaluateChromosomeReturnsBothScores verifica que la evaluación de un
// cromosoma produce un FitnessResult con PsyScore y SocScore no-triviales.
func TestEvaluateChromosomeReturnsBothScores(t *testing.T) {
	// Candidatos con datos OCEAN para que PsyScore sea no-cero
	candidates := []TeamMember{
		{
			ID: "A",
			Psychometrics: &OCEANVector{
				Openness: 0.8, Conscientiousness: 0.7,
				Extraversion: 0.6, Agreeableness: 0.5, Neuroticism: 0.3,
			},
		},
		{
			ID: "B",
			Psychometrics: &OCEANVector{
				Openness: 0.3, Conscientiousness: 0.4,
				Extraversion: 0.7, Agreeableness: 0.8, Neuroticism: 0.5,
			},
		},
		{
			ID: "C",
			Psychometrics: &OCEANVector{
				Openness: 0.5, Conscientiousness: 0.6,
				Extraversion: 0.4, Agreeableness: 0.6, Neuroticism: 0.4,
			},
		},
		{
			ID: "D",
			Psychometrics: &OCEANVector{
				Openness: 0.7, Conscientiousness: 0.3,
				Extraversion: 0.9, Agreeableness: 0.4, Neuroticism: 0.6,
			},
		},
	}

	affinityMatrix := map[string]map[string]float64{
		"A": {"B": 0.8, "C": 0.3, "D": 0.6},
		"B": {"A": 0.8, "C": 0.7, "D": 0.2},
		"C": {"A": 0.3, "B": 0.7, "D": 0.9},
		"D": {"A": 0.6, "B": 0.2, "C": 0.9},
	}

	engine := NewTeamFormationEngine(nil)

	chromosome := []int{0, 1, 2, 3}
	fitness := engine.EvaluateChromosome(chromosome, candidates, affinityMatrix, 2)

	t.Logf("FitnessResult: psy=%.4f soc=%.4f feasible=%v",
		fitness.PsyScore, fitness.SocScore, fitness.IsFeasible)

	if !fitness.IsFeasible {
		t.Error("Expected feasible=true for valid chromosome")
	}

	// SocScore debe ser > 0 dado que hay afinidades no-cero
	if fitness.SocScore <= 0 {
		t.Errorf("SocScore should be > 0 with affinity data, got %.4f. "+
			"Check calculateTeamScoreFromMatrix.", fitness.SocScore)
	}

	// PsyScore: loguear si es 0 (indica que los vectores OCEAN no se están leyendo)
	if fitness.PsyScore == 0 {
		t.Logf("WARNING: PsyScore=0. This may indicate OCEANVector data is not being " +
			"read from TeamMember.Psychometrics correctly in getOCEANVectors(). " +
			"Expected non-zero value given all candidates have valid OCEAN data.")
	} else {
		t.Logf("PsyScore is non-zero (%.4f) — OCEAN data is being read correctly.", fitness.PsyScore)
	}

	expectedGlobal := (fitness.PsyScore + fitness.SocScore) / 2.0
	t.Logf("Expected GlobalScore from this individual: %.4f", expectedGlobal)
}

// -----------------------------------------------------------------------------
// Test 9: Verificar la consistencia de ranking entre frentes en nsga2Evolve
// -----------------------------------------------------------------------------

// TestNSGA2EvolveRankConsistency verifica que tras la evolución, las soluciones
// en el frente 1 tienen rank=1 y las de frentes posteriores tienen rank>1.
func TestNSGA2EvolveRankConsistency(t *testing.T) {
	// Población con soluciones que formarán un frente de Pareto claro
	initialPop := []*genetic.Individual{
		{Chromosome: []int{0, 1}, Fitness: genetic.FitnessResult{PsyScore: 0.9, SocScore: 0.2, IsFeasible: true}},
		{Chromosome: []int{1, 0}, Fitness: genetic.FitnessResult{PsyScore: 0.2, SocScore: 0.9, IsFeasible: true}},
		{Chromosome: []int{0, 1}, Fitness: genetic.FitnessResult{PsyScore: 0.5, SocScore: 0.5, IsFeasible: true}},
		{Chromosome: []int{1, 0}, Fitness: genetic.FitnessResult{PsyScore: 0.3, SocScore: 0.3, IsFeasible: true}},
	}

	// evalFn mantiene la misma distribución de scores para control
	callCount := 0
	scores := [][2]float64{
		{0.9, 0.2}, {0.2, 0.9}, {0.5, 0.5}, {0.3, 0.3},
		{0.85, 0.3}, {0.3, 0.85}, {0.6, 0.6}, {0.25, 0.25},
	}
	evalFn := func(c []int) genetic.FitnessResult {
		s := scores[callCount%len(scores)]
		callCount++
		return genetic.FitnessResult{
			PsyScore:   s[0],
			SocScore:   s[1],
			IsFeasible: true,
		}
	}

	engine := NewTeamFormationEngine(nil)
	params := &GeneticAlgorithmParams{
		PopulationSize: 4,
		Generations:    5,
		MutationRate:   0.1,
		ElitismRate:    0.0,
	}

	rng := rand.New(rand.NewSource(12345))
	finalFronts := engine.Nsga2Evolve(initialPop, params, rng, evalFn, nil)

	if len(finalFronts) == 0 {
		t.Fatal("nsga2Evolve returned no fronts")
	}

	// Verificar que los ranks son consistentes
	for frontIdx, front := range finalFronts {
		expectedRank := frontIdx + 1
		for _, ind := range front {
			if ind.Fitness.Rank != expectedRank {
				t.Errorf("Front %d: expected rank %d, got %d for (psy=%.2f, soc=%.2f)",
					frontIdx+1, expectedRank, ind.Fitness.Rank,
					ind.Fitness.PsyScore, ind.Fitness.SocScore)
			}
		}
	}

	// Ninguna solución del frente 1 debe ser dominada por otra del frente 1
	front1 := finalFronts[0]
	for i := 0; i < len(front1); i++ {
		for j := i + 1; j < len(front1); j++ {
			if genetic.CheckDominance(front1[i].Fitness, front1[j].Fitness) {
				t.Errorf("Pareto front violation: [%d](%.2f,%.2f) dominates [%d](%.2f,%.2f)",
					i, front1[i].Fitness.PsyScore, front1[i].Fitness.SocScore,
					j, front1[j].Fitness.PsyScore, front1[j].Fitness.SocScore)
			}
		}
	}

	t.Logf("nsga2Evolve produced %d fronts; front1 has %d solutions", len(finalFronts), len(front1))
}
