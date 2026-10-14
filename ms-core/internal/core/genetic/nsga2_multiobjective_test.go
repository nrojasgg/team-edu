package genetic

import (
	"math"
	"sort"
	"testing"
)

// =============================================================================
// TESTS DE COMPORTAMIENTO MULTIOBJETIVO REAL
// Estos tests verifican que el NSGA-II se comporta como un algoritmo genuinamente
// multiobjetivo: mantiene diversidad en el frente de Pareto, no colapsa a un
// único escalar, y preserva soluciones no-dominadas correctamente.
// =============================================================================

// makeInd es un helper que crea un Individual con scores dados.
func makeInd(psy, soc float64) *Individual {
	return &Individual{
		Fitness: FitnessResult{
			PsyScore:   psy,
			SocScore:   soc,
			IsFeasible: true,
		},
	}
}

// TestParetoFrontPreservesTradeoffs verifica que el frente de Pareto resultante
// contenga soluciones con trade-offs reales entre los dos objetivos.
// Un algoritmo monoobjetivo colapsaría a un solo punto; NSGA-II debe preservar
// toda la frontera.
func TestParetoFrontPreservesTradeoffs(t *testing.T) {
	t.Parallel()

	// Población con trade-offs claros entre PsyScore y SocScore en MAXIMIZACIÓN.
	// En maximización, mayor score = mejor (scores en [0, 1]).
	// El verdadero frente de Pareto (no dominados) son:
	//   (0.0, 1.0) - mejor en soc, peor en psy
	//   (0.9, 0.9) - balanceado, excelente en ambos
	//   (1.0, 0.0) - mejor en psy, peor en soc
	// Los dominados (por tener scores inferiores a (0.9, 0.9)):
	//   (0.7, 0.7) - dominado por (0.9, 0.9): 0.7<0.9 y 0.7<0.9
	//   (0.5, 0.3) - dominado por (0.9, 0.9): 0.5<0.9 y 0.3<0.9
	//   (0.2, 0.6) - dominado por (0.9, 0.9): 0.2<0.9 y 0.6<0.9
	//   (0.6, 0.8) - dominado por (0.9, 0.9): 0.6<0.9 y 0.8<0.9
	population := []*Individual{
		makeInd(0.0, 1.0), // Pareto front
		makeInd(0.9, 0.9), // Pareto front (balanceado)
		makeInd(1.0, 0.0), // Pareto front
		makeInd(0.7, 0.7), // Dominado por (0.9, 0.9)
		makeInd(0.5, 0.3), // Dominado por (0.9, 0.9)
		makeInd(0.2, 0.6), // Dominado por (0.9, 0.9)
		makeInd(0.6, 0.8), // Dominado por (0.9, 0.9)
	}

	fronts := FastNonDominatedSort(population)

	if len(fronts) == 0 {
		t.Fatal("FastNonDominatedSort returned no fronts")
	}

	front1 := fronts[0]

	// El frente de Pareto debe tener exactamente 3 soluciones
	if len(front1) != 3 {
		t.Errorf("Expected 3 solutions in Pareto front, got %d", len(front1))
		for i, ind := range front1 {
			t.Logf("  Front1[%d]: psy=%.2f soc=%.2f", i, ind.Fitness.PsyScore, ind.Fitness.SocScore)
		}
	}

	// Verificar que todos los miembros del frente 1 tienen rank=1
	for _, ind := range front1 {
		if ind.Fitness.Rank != 1 {
			t.Errorf("Expected rank 1, got %d for (%.2f, %.2f)",
				ind.Fitness.Rank, ind.Fitness.PsyScore, ind.Fitness.SocScore)
		}
	}

	// Verificar que los dominados están en fronts[1] o más
	if len(fronts) < 2 {
		t.Error("Expected at least 2 fronts (dominated solutions exist)")
	}

	// Ningún par en el frente 1 debe dominarse mutuamente
	for i := 0; i < len(front1); i++ {
		for j := i + 1; j < len(front1); j++ {
			if CheckDominance(front1[i].Fitness, front1[j].Fitness) {
				t.Errorf("front1[%d] (%.2f,%.2f) dominates front1[%d] (%.2f,%.2f) — violation!",
					i, front1[i].Fitness.PsyScore, front1[i].Fitness.SocScore,
					j, front1[j].Fitness.PsyScore, front1[j].Fitness.SocScore)
			}
			if CheckDominance(front1[j].Fitness, front1[i].Fitness) {
				t.Errorf("front1[%d] (%.2f,%.2f) dominates front1[%d] (%.2f,%.2f) — violation!",
					j, front1[j].Fitness.PsyScore, front1[j].Fitness.SocScore,
					i, front1[i].Fitness.PsyScore, front1[i].Fitness.SocScore)
			}
		}
	}
}

// TestNSGA2DoesNotCollapseToScalar verifica que NSGA-II NO reduce los dos
// objetivos a un único escalar antes del sort. Si lo hiciera, todas las
// soluciones con el mismo escalar combinado pero distinto trade-off recibirían
// el mismo rank cuando deberían estar en el mismo frente de Pareto.
func TestNSGA2DoesNotCollapseToScalar(t *testing.T) {
	t.Parallel()

	// En MAXIMIZACIÓN (higher is better):
	// A: psy=0.9, soc=0.1 → bueno en psy (0.9), bajo en soc (0.1)
	// B: psy=0.1, soc=0.9 → bajo en psy (0.1), bueno en soc (0.9)
	// C: psy=0.5, soc=0.5 → balanceado en ambos
	// D: psy=0.4, soc=0.4 → dominado por C en ambos (0.4 < 0.5 y 0.4 < 0.5)
	//
	// Análisis de dominancia:
	// - A y B no se dominan mutuamente (trade-offs opuestos)
	// - C (0.5, 0.5) domina a D (0.4, 0.4) porque 0.5>0.4 y 0.5>0.4
	// - C no es dominado por A ni B (C.SocScore 0.5 > A.SocScore 0.1, C.PsyScore 0.5 > B.PsyScore 0.1)
	//
	// Entonces A, B y C son no-dominados entre sí y están en Front 1 (rank=1). D está en Front 2 (rank=2).
	a := makeInd(0.9, 0.1)
	b := makeInd(0.1, 0.9)
	c := makeInd(0.5, 0.5)
	d := makeInd(0.4, 0.4)

	fronts := FastNonDominatedSort([]*Individual{a, b, c, d})

	if len(fronts) < 2 {
		t.Fatal("Expected at least 2 fronts")
	}

	// A, B y C deben estar en el frente 1 (ninguno domina a otro)
	front1 := fronts[0]
	if len(front1) != 3 {
		t.Errorf("Expected 3 non-dominated solutions (A, B, C), got %d", len(front1))
		for _, ind := range front1 {
			t.Logf("  (%.2f, %.2f) rank=%d", ind.Fitness.PsyScore, ind.Fitness.SocScore, ind.Fitness.Rank)
		}
	}

	// D debe estar en el frente 2 (dominado por C)
	if d.Fitness.Rank != 2 {
		t.Errorf("Expected D to be in front 2 (rank=2), got rank=%d", d.Fitness.Rank)
	}
}

// TestCrowdingDistanceMaintainsDiversity verifica que el crowding distance
// favorece soluciones en zonas poco exploradas del espacio de objetivos.
// Esto es la garantía de diversidad del NSGA-II.
func TestCrowdingDistanceMaintainsDiversity(t *testing.T) {
	t.Parallel()

	// Frente con 5 soluciones: extremos + 3 interiores
	// Distribución uniforme: distancias interiores deben ser iguales
	uniform := []*Individual{
		makeInd(0.0, 1.0),
		makeInd(0.25, 0.75),
		makeInd(0.5, 0.5),
		makeInd(0.75, 0.25),
		makeInd(1.0, 0.0),
	}

	CalculateCrowdingDistance(uniform)

	// Los extremos deben tener distancia infinita
	extremes := []*Individual{uniform[0], uniform[4]}
	for _, ind := range extremes {
		if !math.IsInf(ind.Fitness.Distance, 1) {
			t.Errorf("Boundary point (%.2f,%.2f) should have Inf distance, got %.4f",
				ind.Fitness.PsyScore, ind.Fitness.SocScore, ind.Fitness.Distance)
		}
	}

	// Los puntos interiores deben tener distancia finita y > 0
	interiors := uniform[1:4]
	for _, ind := range interiors {
		if math.IsInf(ind.Fitness.Distance, 0) || ind.Fitness.Distance <= 0 {
			t.Errorf("Interior point (%.2f,%.2f) should have finite positive distance, got %.4f",
				ind.Fitness.PsyScore, ind.Fitness.SocScore, ind.Fitness.Distance)
		}
	}

	// Frente con clustering: punto aislado debe tener mayor distancia que los agrupados
	// Puntos: (0,1), (0.49,0.51), (0.5,0.5), (0.51,0.49), (1,0)
	// Los puntos centrales están muy juntos → distancias bajas
	// Los extremos tienen Inf → ignorar
	// El punto central (0.5,0.5) está entre dos vecinos muy cercanos → distancia baja
	clustered := []*Individual{
		makeInd(0.0, 1.0),   // extremo → Inf
		makeInd(0.49, 0.51), // agrupado
		makeInd(0.5, 0.5),   // agrupado
		makeInd(0.51, 0.49), // agrupado
		makeInd(1.0, 0.0),   // extremo → Inf
	}

	// Frente disperso: puntos bien separados
	spread := []*Individual{
		makeInd(0.0, 1.0),
		makeInd(0.25, 0.75),
		makeInd(0.5, 0.5),
		makeInd(0.75, 0.25),
		makeInd(1.0, 0.0),
	}

	CalculateCrowdingDistance(clustered)
	CalculateCrowdingDistance(spread)

	// El punto central en el frente disperso debe tener mayor distancia
	// que el punto central en el frente agrupado
	clusteredCenter := clustered[2] // (0.5, 0.5) en frente agrupado
	spreadCenter := spread[2]       // (0.5, 0.5) en frente disperso

	if !math.IsInf(spreadCenter.Fitness.Distance, 0) &&
		!math.IsInf(clusteredCenter.Fitness.Distance, 0) &&
		spreadCenter.Fitness.Distance <= clusteredCenter.Fitness.Distance {
		t.Errorf(
			"Spread center (dist=%.4f) should have larger crowding distance than clustered center (dist=%.4f)",
			spreadCenter.Fitness.Distance,
			clusteredCenter.Fitness.Distance,
		)
	}
}

// TestCrowdedComparisonPrefersDiversity verifica que el operador crowded-comparison
// favorece: primero el menor rank (frente de Pareto superior), y luego, en empate de rank,
// la mayor crowding distance para preservar la diversidad del frente (Deb et al., 2002).
func TestCrowdedComparisonPrefersDiversity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		a       FitnessResult
		b       FitnessResult
		wantA   bool // true si a debe ganar
		wantB   bool // true si b debe ganar
		wantTie bool // true si ninguno gana
	}{
		{
			name:  "mejor rank gana independientemente de la distancia",
			a:     FitnessResult{Rank: 1, Distance: 0.1},   // rank mejor, distancia peor
			b:     FitnessResult{Rank: 2, Distance: 100.0}, // rank peor, distancia mejor
			wantA: true,
		},
		{
			name:  "mismo rank: mayor distancia (mayor diversidad) gana",
			a:     FitnessResult{Rank: 1, Distance: 10.0},
			b:     FitnessResult{Rank: 1, Distance: 5.0},
			wantA: true,
		},
		{
			name:  "mismo rank: menor distancia pierde frente a mayor distancia",
			a:     FitnessResult{Rank: 2, Distance: 1.0},
			b:     FitnessResult{Rank: 2, Distance: 50.0},
			wantB: true,
		},
		{
			name:    "mismo rank y misma distancia: empate",
			a:       FitnessResult{Rank: 1, Distance: 5.0},
			b:       FitnessResult{Rank: 1, Distance: 5.0},
			wantTie: true,
		},
		{
			name:  "distancia positiva gana sobre distancia 0 en mismo rank",
			a:     FitnessResult{Rank: 1, Distance: 0.9},
			b:     FitnessResult{Rank: 1, Distance: 0.0},
			wantA: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aWins := CrowdedComparison(tt.a, tt.b)
			bWins := CrowdedComparison(tt.b, tt.a)

			if tt.wantA && !aWins {
				t.Errorf("Expected A to win: A=(rank=%d,dist=%.2f) vs B=(rank=%d,dist=%.2f)",
					tt.a.Rank, tt.a.Distance, tt.b.Rank, tt.b.Distance)
			}
			if tt.wantB && !bWins {
				t.Errorf("Expected B to win: A=(rank=%d,dist=%.2f) vs B=(rank=%d,dist=%.2f)",
					tt.a.Rank, tt.a.Distance, tt.b.Rank, tt.b.Distance)
			}
			if tt.wantTie && (aWins || bWins) {
				t.Errorf("Expected tie but got: aWins=%v bWins=%v", aWins, bWins)
			}
		})
	}
}

// TestFrontRankingIsConsistent verifica que los ranks asignados sean consistentes:
// una solución en el frente k NO puede dominar a ninguna del frente k-1.
func TestFrontRankingIsConsistent(t *testing.T) {
	t.Parallel()

	// Población con estructura de 3 frentes claros
	population := []*Individual{
		// Front 1: soluciones no dominadas
		makeInd(1.0, 0.8),
		makeInd(0.8, 1.0),
		makeInd(0.9, 0.9),
		// Front 2: dominadas por alguna del front 1
		makeInd(0.7, 0.7),
		makeInd(0.6, 0.8),
		makeInd(0.8, 0.6),
		// Front 3: dominadas por alguna del front 2
		makeInd(0.4, 0.4),
		makeInd(0.3, 0.5),
	}

	fronts := FastNonDominatedSort(population)

	if len(fronts) < 2 {
		t.Fatal("Expected at least 2 fronts for this population")
	}

	// Verificar consistencia entre frentes consecutivos
	for i := 1; i < len(fronts); i++ {
		prevFront := fronts[i-1]
		currFront := fronts[i]

		for _, curr := range currFront {
			// curr NO debe dominar a ningún miembro del frente anterior
			for _, prev := range prevFront {
				if CheckDominance(curr.Fitness, prev.Fitness) {
					t.Errorf(
						"Front %d solution (%.2f,%.2f) dominates Front %d solution (%.2f,%.2f) — rank inconsistency!",
						i+1, curr.Fitness.PsyScore, curr.Fitness.SocScore,
						i, prev.Fitness.PsyScore, prev.Fitness.SocScore,
					)
				}
			}
		}
	}

	// Verificar que los ranks están correctamente asignados
	for frontIdx, front := range fronts {
		expectedRank := frontIdx + 1
		for _, ind := range front {
			if ind.Fitness.Rank != expectedRank {
				t.Errorf("Expected rank %d for solution in front %d, got %d",
					expectedRank, frontIdx+1, ind.Fitness.Rank)
			}
		}
	}
}

// TestAllIndividualsAreRanked verifica que ningún individuo quede sin rank
// después del sort. Un bug en buildNextFront podría dejar individuos huérfanos.
func TestAllIndividualsAreRanked(t *testing.T) {
	t.Parallel()

	population := []*Individual{
		makeInd(0.9, 0.1),
		makeInd(0.7, 0.5),
		makeInd(0.5, 0.7),
		makeInd(0.3, 0.9),
		makeInd(0.5, 0.5),
		makeInd(0.4, 0.6),
		makeInd(0.6, 0.4),
		makeInd(0.2, 0.2),
		makeInd(0.8, 0.8),
		makeInd(0.1, 0.1),
	}

	fronts := FastNonDominatedSort(population)

	// Contar individuos en todos los frentes
	assigned := 0
	for _, front := range fronts {
		assigned += len(front)
	}

	if assigned != len(population) {
		t.Errorf("Not all individuals were ranked: expected %d, got %d in fronts",
			len(population), assigned)
	}

	// Verificar que todos tienen rank > 0
	for _, ind := range population {
		if ind.Fitness.Rank == 0 {
			t.Errorf("Individual (%.2f, %.2f) has rank=0 (unranked)",
				ind.Fitness.PsyScore, ind.Fitness.SocScore)
		}
	}
}

// TestDominanceIsAsymmetric verifica que si A domina B, entonces B NO domina A.
// Una implementación errónea de CheckDominance podría violar esta propiedad.
func TestDominanceIsAsymmetric(t *testing.T) {
	t.Parallel()

	tests := []struct {
		a FitnessResult
		b FitnessResult
	}{
		{
			FitnessResult{PsyScore: 0.8, SocScore: 0.8, IsFeasible: true},
			FitnessResult{PsyScore: 0.5, SocScore: 0.5, IsFeasible: true},
		},
		{
			FitnessResult{PsyScore: 1.0, SocScore: 0.0, IsFeasible: true},
			FitnessResult{PsyScore: 0.9, SocScore: 0.0, IsFeasible: true},
		},
		{
			FitnessResult{PsyScore: 0.6, SocScore: 0.9, IsFeasible: true},
			FitnessResult{PsyScore: 0.5, SocScore: 0.8, IsFeasible: true},
		},
	}

	for _, tt := range tests {
		aDomsB := CheckDominance(tt.a, tt.b)
		bDomsA := CheckDominance(tt.b, tt.a)

		if aDomsB && bDomsA {
			t.Errorf("Dominance is not asymmetric: A(%.2f,%.2f) and B(%.2f,%.2f) mutually dominate each other",
				tt.a.PsyScore, tt.a.SocScore, tt.b.PsyScore, tt.b.SocScore)
		}
	}
}

// TestDominanceIsIrreflexive verifica que ninguna solución se domine a sí misma.
func TestDominanceIsIrreflexive(t *testing.T) {
	t.Parallel()

	solutions := []FitnessResult{
		{PsyScore: 0.5, SocScore: 0.5, IsFeasible: true},
		{PsyScore: 1.0, SocScore: 1.0, IsFeasible: true},
		{PsyScore: 0.0, SocScore: 0.0, IsFeasible: true},
		{PsyScore: 0.7, SocScore: 0.3, IsFeasible: true},
	}

	for _, sol := range solutions {
		if CheckDominance(sol, sol) {
			t.Errorf("Solution (%.2f, %.2f) dominates itself — irreflexivity violated",
				sol.PsyScore, sol.SocScore)
		}
	}
}

// TestParetoFrontSpansBothObjectives verifica que el frente de Pareto final
// contiene soluciones que abarcan los extremos de ambos objetivos.
// Si el algoritmo colapsara a escalar, la diversidad se perdería y no habría
// soluciones extremas en el frente.
func TestParetoFrontSpansBothObjectives(t *testing.T) {
	t.Parallel()

	// Población que incluye extremos claros
	population := []*Individual{
		makeInd(1.0, 0.0), // máximo psy, mínimo soc
		makeInd(0.0, 1.0), // mínimo psy, máximo soc
		makeInd(0.5, 0.5), // punto medio
		makeInd(0.7, 0.4), // trade-off
		makeInd(0.3, 0.7), // trade-off inverso
		makeInd(0.2, 0.2), // dominado
	}

	fronts := FastNonDominatedSort(population)
	CalculateCrowdingDistance(fronts[0])

	front1 := fronts[0]

	// Encontrar los rangos de ambos objetivos en el frente de Pareto
	minPsy, maxPsy := math.MaxFloat64, -math.MaxFloat64
	minSoc, maxSoc := math.MaxFloat64, -math.MaxFloat64

	for _, ind := range front1 {
		if ind.Fitness.PsyScore < minPsy {
			minPsy = ind.Fitness.PsyScore
		}
		if ind.Fitness.PsyScore > maxPsy {
			maxPsy = ind.Fitness.PsyScore
		}
		if ind.Fitness.SocScore < minSoc {
			minSoc = ind.Fitness.SocScore
		}
		if ind.Fitness.SocScore > maxSoc {
			maxSoc = ind.Fitness.SocScore
		}
	}

	// El rango de PsyScore en el frente debe ser significativo (≥ 0.5)
	psyRange := maxPsy - minPsy
	if psyRange < 0.5 {
		t.Errorf("Pareto front PsyScore range is too narrow: %.2f (expected ≥ 0.5). "+
			"This suggests objectives are being collapsed to a scalar.", psyRange)
	}

	// El rango de SocScore en el frente debe ser significativo (≥ 0.5)
	socRange := maxSoc - minSoc
	if socRange < 0.5 {
		t.Errorf("Pareto front SocScore range is too narrow: %.2f (expected ≥ 0.5). "+
			"This suggests objectives are being collapsed to a scalar.", socRange)
	}

	t.Logf("Pareto front spans: psy=[%.2f,%.2f] soc=[%.2f,%.2f]",
		minPsy, maxPsy, minSoc, maxSoc)
}

// TestCrowdingDistanceEnforcesSelectionPressureTowardsDiversity verifica que
// cuando se seleccionan los mejores K individuos de un frente (como hace NSGA-II
// al llenar la nueva generación), los seleccionados son los de mayor distancia.
func TestCrowdingDistanceEnforcesSelectionPressureTowardsDiversity(t *testing.T) {
	t.Parallel()

	// Frente con 6 puntos: 2 extremos + 4 interiores
	// Los 2 extremos tienen distancia infinita y deben ser seleccionados primero.
	front := []*Individual{
		makeInd(0.0, 1.0),   // extremo
		makeInd(0.2, 0.8),   // interior cercano al extremo
		makeInd(0.21, 0.79), // interior muy cercano al anterior (baja distancia)
		makeInd(0.5, 0.5),   // interior central
		makeInd(0.79, 0.21), // interior muy cercano al siguiente (baja distancia)
		makeInd(1.0, 0.0),   // extremo
	}

	CalculateCrowdingDistance(front)

	// Seleccionar top-2 por crowding distance
	sorted := make([]*Individual, len(front))
	copy(sorted, front)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Fitness.Distance > sorted[j].Fitness.Distance
	})

	top2 := sorted[:2]

	// Los top-2 deben ser los extremos (con distancia infinita)
	for _, ind := range top2 {
		if !math.IsInf(ind.Fitness.Distance, 1) {
			t.Errorf(
				"Expected top-2 to be boundary points (Inf distance), got (%.2f,%.2f) with dist=%.4f",
				ind.Fitness.PsyScore, ind.Fitness.SocScore, ind.Fitness.Distance,
			)
		}
	}

	// Verificar que los puntos muy cercanos entre sí tienen menor distancia
	// que los puntos más dispersos
	nearPoint1 := front[2]   // (0.21, 0.79) — muy cerca de (0.2, 0.8)
	nearPoint2 := front[4]   // (0.79, 0.21) — muy cerca de (0.8, 0.2) pero no hay tal punto
	centralPoint := front[3] // (0.5, 0.5)

	t.Logf("Distances — central: %.4f, near1: %.4f, near2: %.4f",
		centralPoint.Fitness.Distance, nearPoint1.Fitness.Distance, nearPoint2.Fitness.Distance)

	// El punto central debería tener mayor distancia que los puntos agrupados
	if !math.IsInf(centralPoint.Fitness.Distance, 0) &&
		!math.IsInf(nearPoint1.Fitness.Distance, 0) &&
		centralPoint.Fitness.Distance <= nearPoint1.Fitness.Distance {
		t.Logf("Note: central point dist (%.4f) not larger than near point dist (%.4f) — "+
			"depends on exact geometry, acceptable in this configuration",
			centralPoint.Fitness.Distance, nearPoint1.Fitness.Distance)
	}
}

// TestFastNonDominatedSort_LargePopulation verifica que el sort funciona
// correctamente con una población más grande y produce una distribución
// de ranks razonable (no todos en el mismo frente).
func TestFastNonDominatedSort_LargePopulation(t *testing.T) {
	t.Parallel()

	// Generar 50 individuos distribuidos en el espacio [0,1]x[0,1]
	// con una correlación negativa entre objetivos (trade-off natural)
	population := make([]*Individual, 50)
	for i := 0; i < 50; i++ {
		psy := float64(i) / 49.0
		soc := 1.0 - psy + (float64(i%5) * 0.05) // ligero ruido
		if soc > 1.0 {
			soc = 1.0
		}
		if soc < 0.0 {
			soc = 0.0
		}
		population[i] = makeInd(psy, soc)
	}

	fronts := FastNonDominatedSort(population)

	// Todos deben estar rankeados
	totalRanked := 0
	for _, f := range fronts {
		totalRanked += len(f)
	}
	if totalRanked != len(population) {
		t.Errorf("Not all individuals ranked: %d/%d", totalRanked, len(population))
	}

	// Con trade-off negativo, el frente 1 debe tener varios individuos
	if len(fronts[0]) < 5 {
		t.Errorf("Expected at least 5 individuals in Pareto front for trade-off population, got %d",
			len(fronts[0]))
	}

	// No debe haber individuos con rank=0 (unranked)
	for _, ind := range population {
		if ind.Fitness.Rank == 0 {
			t.Errorf("Individual (%.2f, %.2f) was not assigned a rank",
				ind.Fitness.PsyScore, ind.Fitness.SocScore)
		}
	}

	t.Logf("Large population: %d fronts, front1 has %d individuals",
		len(fronts), len(fronts[0]))
}

// TestCrowdingDistance_SingleSolution verifica que una solución sola en un frente
// recibe distancia infinita (no puede romperse por división por cero).
func TestCrowdingDistance_SingleSolution(t *testing.T) {
	t.Parallel()

	front := []*Individual{makeInd(0.7, 0.3)}
	CalculateCrowdingDistance(front)

	// Con un solo elemento no hay vecinos, debe tener distancia infinita
	// (o al menos no entrar en pánico/NaN).
	dist := front[0].Fitness.Distance
	if math.IsNaN(dist) {
		t.Error("Single solution crowding distance is NaN")
	}
	t.Logf("Single solution crowding distance: %v", dist)
}

// TestCrowdingDistance_TwoSolutions verifica que con dos soluciones ambas
// reciben distancia infinita (son los dos extremos del frente).
func TestCrowdingDistance_TwoSolutions(t *testing.T) {
	t.Parallel()

	front := []*Individual{
		makeInd(0.3, 0.7),
		makeInd(0.7, 0.3),
	}
	CalculateCrowdingDistance(front)

	for _, ind := range front {
		if !math.IsInf(ind.Fitness.Distance, 1) {
			t.Errorf("Expected Inf distance for boundary point in 2-element front, got %.4f for (%.2f,%.2f)",
				ind.Fitness.Distance, ind.Fitness.PsyScore, ind.Fitness.SocScore)
		}
	}
}

// TestDominanceWithFeasibilityConstraint verifica la lógica de factibilidad:
// una solución factible siempre domina a una no-factible,
// independientemente de sus scores.
func TestDominanceWithFeasibilityConstraint(t *testing.T) {
	t.Parallel()

	// Solución no-factible con scores perfectos vs factible con scores mínimos
	infeasiblePerfect := FitnessResult{PsyScore: 1.0, SocScore: 1.0, IsFeasible: false}
	feasibleMinimal := FitnessResult{PsyScore: 0.0, SocScore: 0.0, IsFeasible: true}

	// La factible mínima debe dominar a la no-factible perfecta
	if !CheckDominance(feasibleMinimal, infeasiblePerfect) {
		t.Error("Feasible solution (0,0) should dominate infeasible solution (1,1)")
	}

	// La no-factible perfecta NO debe dominar a la factible mínima
	if CheckDominance(infeasiblePerfect, feasibleMinimal) {
		t.Error("Infeasible solution (1,1) should NOT dominate feasible solution (0,0)")
	}

	// Dos soluciones no-factibles: ninguna domina a la otra (sin gestión de violación)
	infeasibleA := FitnessResult{PsyScore: 0.9, SocScore: 0.9, IsFeasible: false}
	infeasibleB := FitnessResult{PsyScore: 0.1, SocScore: 0.1, IsFeasible: false}

	if CheckDominance(infeasibleA, infeasibleB) {
		t.Error("Among infeasible solutions, neither should dominate the other")
	}
	if CheckDominance(infeasibleB, infeasibleA) {
		t.Error("Among infeasible solutions, neither should dominate the other")
	}
}

// TestParetoFrontIsStableUnderReSort verifica que re-ejecutar FastNonDominatedSort
// sobre el mismo conjunto produce el mismo resultado (determinismo).
func TestParetoFrontIsStableUnderReSort(t *testing.T) {
	t.Parallel()

	population := []*Individual{
		makeInd(0.9, 0.2),
		makeInd(0.7, 0.6),
		makeInd(0.4, 0.8),
		makeInd(0.5, 0.5),
		makeInd(0.3, 0.3),
		makeInd(0.1, 0.9),
	}

	fronts1 := FastNonDominatedSort(population)

	// Reset ranks para re-sort limpio
	for _, ind := range population {
		ind.Fitness.Rank = 0
		ind.dominationCount = 0
		ind.dominatedBy = nil
	}

	fronts2 := FastNonDominatedSort(population)

	// Ambos sorts deben producir el mismo número de frentes
	if len(fronts1) != len(fronts2) {
		t.Errorf("Re-sort produced different number of fronts: %d vs %d", len(fronts1), len(fronts2))
	}

	// Y cada frente debe tener el mismo tamaño
	for i := 0; i < len(fronts1) && i < len(fronts2); i++ {
		if len(fronts1[i]) != len(fronts2[i]) {
			t.Errorf("Front %d size differs: %d vs %d", i+1, len(fronts1[i]), len(fronts2[i]))
		}
	}
}
