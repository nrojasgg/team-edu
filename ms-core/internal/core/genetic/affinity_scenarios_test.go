package genetic

import (
	"context"
	"math"
	"testing"

	"ms-core/internal/core/domain"
)

// ----------------------------------------------------------------------------
// Helper fixtures
// ----------------------------------------------------------------------------

func makeStudent(id string, numAttrs map[string]float64, catAttrs map[string]string) *domain.Student {
	s := &domain.Student{
		ID:               id,
		NumericAttrs:     make(map[string]float64),
		CategoricalAttrs: make(map[string]string),
	}
	for k, v := range numAttrs {
		s.NumericAttrs[k] = v
	}
	for k, v := range catAttrs {
		s.CategoricalAttrs[k] = v
	}
	return s
}

// ============================================================================
// ESCENARIO 1: 1 Característica Psicológica
// ============================================================================
// Verifica que con 1 solo rasgo (ej. openness), el cálculo de afinidad del grupo
// depende exclusivamente de ese rasgo según la estrategia configurada (target).
func TestAffinity_Scenario1_SinglePsychologicalCriterion(t *testing.T) {
	t.Parallel()

	factory := NewCriterionFactory()
	factory.Register(NewNumericCriterion("openness", "openness", "target"))

	config := domain.Config{
		Weights: map[string]float64{"openness": 1.0},
		Targets: map[string]float64{"openness": 0.80}, // Target deseado: media = 0.80
	}
	ctx := domain.Context{
		Targets: config.Targets,
	}

	// Grupo A: Estudiantes con alta apertura que promedian exactamente 0.80
	// Media: (0.85 + 0.75 + 0.80) / 3 = 0.80 -> Error = |0.80 - 0.80| = 0 -> Score = 1.00
	groupA := domain.Group{
		makeStudent("s1", map[string]float64{"openness": 0.85}, nil),
		makeStudent("s2", map[string]float64{"openness": 0.75}, nil),
		makeStudent("s3", map[string]float64{"openness": 0.80}, nil),
	}

	// Grupo B: Estudiantes con menor apertura que promedian 0.60
	// Media: (0.50 + 0.60 + 0.70) / 3 = 0.60 -> Error = |0.60 - 0.80| = 0.20 -> Score = 0.80
	groupB := domain.Group{
		makeStudent("s4", map[string]float64{"openness": 0.50}, nil),
		makeStudent("s5", map[string]float64{"openness": 0.60}, nil),
		makeStudent("s6", map[string]float64{"openness": 0.70}, nil),
	}

	scoreA := CalculateFitnessForGroup(&groupA, config, factory, ctx)
	scoreB := CalculateFitnessForGroup(&groupB, config, factory, ctx)

	// Verificación 1: Grupo A debe tener score perfecto 1.0
	if math.Abs(scoreA-1.00) > 1e-9 {
		t.Errorf("Scenario 1 - Group A: expected score=1.00, got %.6f", scoreA)
	}

	// Verificación 2: Grupo B debe tener score 0.80
	if math.Abs(scoreB-0.80) > 1e-9 {
		t.Errorf("Scenario 1 - Group B: expected score=0.80, got %.6f", scoreB)
	}

	// Verificación 3: Afinidad global como promedio de ambos grupos
	chromosome := &domain.Chromosome{Groups: []domain.Group{groupA, groupB}}
	globalScore := CalculateFitness(chromosome, config, factory, ctx)
	expectedGlobal := (1.00 + 0.80) / 2.0 // 0.90
	if math.Abs(globalScore-expectedGlobal) > 1e-9 {
		t.Errorf("Scenario 1 - GlobalScore: expected %.4f, got %.6f", expectedGlobal, globalScore)
	}
}

// ============================================================================
// ESCENARIO 2: 2 Características Psicológicas
// ============================================================================
// Verifica la ponderación y cálculo combinado de 2 rasgos (openness y conscientiousness).
func TestAffinity_Scenario2_TwoPsychologicalCriteria(t *testing.T) {
	t.Parallel()

	factory := NewCriterionFactory()
	factory.Register(NewNumericCriterion("openness", "openness", "target"))
	factory.Register(NewNumericCriterion("conscientiousness", "conscientiousness", "target"))

	// Pesos: openness = 0.6 (60%), conscientiousness = 0.4 (40%)
	config := domain.Config{
		Weights: map[string]float64{
			"openness":          0.6,
			"conscientiousness": 0.4,
		},
		Targets: map[string]float64{
			"openness":          0.80,
			"conscientiousness": 0.90,
		},
	}
	ctx := domain.Context{
		Targets: config.Targets,
	}

	// Grupo evaluado:
	// - openness: (0.85 + 0.75) / 2 = 0.80 -> Error = 0.0 -> Score1 = 1.00
	// - conscientiousness: (0.70 + 0.70) / 2 = 0.70 -> Error = |0.70 - 0.90| = 0.20 -> Score2 = 0.80
	// Score combinado esperado: 0.6 * 1.00 + 0.4 * 0.80 = 0.60 + 0.32 = 0.92
	group := domain.Group{
		makeStudent("s1", map[string]float64{"openness": 0.85, "conscientiousness": 0.70}, nil),
		makeStudent("s2", map[string]float64{"openness": 0.75, "conscientiousness": 0.70}, nil),
	}

	// Score individual de cada criterio
	critOpen := factory.Get("openness")
	critConsc := factory.Get("conscientiousness")
	sOpen := critOpen.Calculate(&group, ctx)
	sConsc := critConsc.Calculate(&group, ctx)

	if math.Abs(sOpen-1.00) > 1e-9 {
		t.Errorf("Scenario 2 - sOpen expected 1.00, got %.6f", sOpen)
	}
	if math.Abs(sConsc-0.80) > 1e-9 {
		t.Errorf("Scenario 2 - sConsc expected 0.80, got %.6f", sConsc)
	}

	// Score ponderado del grupo
	scoreGroup := CalculateFitnessForGroup(&group, config, factory, ctx)
	expectedScore := 0.6*sOpen + 0.4*sConsc // 0.92
	if math.Abs(scoreGroup-expectedScore) > 1e-9 {
		t.Errorf("Scenario 2 - Group AffinityScore: expected %.6f, got %.6f", expectedScore, scoreGroup)
	}
}

// ============================================================================
// ESCENARIO 3: 3 Características Psicológicas
// ============================================================================
// Verifica la normalización de pesos y combinación de 3 rasgos Big Five.
func TestAffinity_Scenario3_ThreePsychologicalCriteria(t *testing.T) {
	t.Parallel()

	factory := NewCriterionFactory()
	factory.Register(NewNumericCriterion("openness", "openness", "target"))
	factory.Register(NewNumericCriterion("extraversion", "extraversion", "target"))
	factory.Register(NewNumericCriterion("agreeableness", "agreeableness", "target"))

	// Pesos no normalizados (suman 5.0):
	// openness = 2.0 (40%), extraversion = 2.0 (40%), agreeableness = 1.0 (20%)
	config := domain.Config{
		Weights: map[string]float64{
			"openness":      2.0,
			"extraversion":  2.0,
			"agreeableness": 1.0,
		},
		Targets: map[string]float64{
			"openness":      0.75,
			"extraversion":  0.60,
			"agreeableness": 0.85,
		},
	}
	ctx := domain.Context{
		Targets: config.Targets,
	}

	group := domain.Group{
		makeStudent("s1", map[string]float64{"openness": 0.80, "extraversion": 0.50, "agreeableness": 0.85}, nil),
		makeStudent("s2", map[string]float64{"openness": 0.70, "extraversion": 0.70, "agreeableness": 0.85}, nil),
	}

	// Medias:
	// openness: (0.80 + 0.70)/2 = 0.75 -> Error = 0.0 -> Score = 1.00
	// extraversion: (0.50 + 0.70)/2 = 0.60 -> Error = 0.0 -> Score = 1.00
	// agreeableness: (0.85 + 0.85)/2 = 0.85 -> Error = 0.0 -> Score = 1.00
	// Con los 3 targets cumplidos al 100%, el score ponderado debe ser 1.00 independientemente de los pesos.
	scoreGroup := CalculateFitnessForGroup(&group, config, factory, ctx)
	if math.Abs(scoreGroup-1.00) > 1e-9 {
		t.Errorf("Scenario 3 - Perfect 3 traits: expected score=1.00, got %.6f", scoreGroup)
	}

	// Ahora probamos con desviación en uno de los criterios:
	groupPartial := domain.Group{
		makeStudent("s1", map[string]float64{"openness": 0.75, "extraversion": 0.60, "agreeableness": 0.65}, nil),
		makeStudent("s2", map[string]float64{"openness": 0.75, "extraversion": 0.60, "agreeableness": 0.65}, nil),
	}
	// agreeableness media = 0.65 -> Error = |0.65 - 0.85| = 0.20 -> Score_agree = 0.80
	// Pesos normalizados: w_open = 2/5 = 0.4, w_extra = 2/5 = 0.4, w_agree = 1/5 = 0.2
	// Expected score: 0.4 * 1.0 + 0.4 * 1.0 + 0.2 * 0.80 = 0.4 + 0.4 + 0.16 = 0.96
	scorePartial := CalculateFitnessForGroup(&groupPartial, config, factory, ctx)
	expectedPartial := 0.96
	if math.Abs(scorePartial-expectedPartial) > 1e-9 {
		t.Errorf("Scenario 3 - Partial deviation: expected score=%.4f, got %.6f", expectedPartial, scorePartial)
	}
}

// ============================================================================
// ESCENARIO 4: Solo Características Sociales
// ============================================================================
// Verifica el cálculo de costo relacional C'_ij y cohesión S_rel(G) con grafo sociométrico.
func TestAffinity_Scenario4_SocialCriteriaOnly(t *testing.T) {
	t.Parallel()

	// Subcaso 4A: Armonía (S_p = 1.0)
	harmonyCriterion := NewSocialCriterion("social", 1.0)
	factoryHarmony := NewCriterionFactory()
	factoryHarmony.Register(harmonyCriterion)

	configHarmony := domain.Config{
		Weights: map[string]float64{"social": 1.0},
	}

	// Grafo de atracción mutua perfecta entre 3 miembros:
	// Todos se eligen mutuamente con +1.0
	graphFullAffinity := domain.SocialGraph{
		"A": {"B": 1.0, "C": 1.0},
		"B": {"A": 1.0, "C": 1.0},
		"C": {"A": 1.0, "B": 1.0},
	}
	group3 := domain.Group{
		makeStudent("A", nil, nil),
		makeStudent("B", nil, nil),
		makeStudent("C", nil, nil),
	}
	ctxHarmony := domain.Context{SocialGraph: graphFullAffinity}

	scoreHarmonyFull := CalculateFitnessForGroup(&group3, configHarmony, factoryHarmony, ctxHarmony)
	if math.Abs(scoreHarmonyFull-1.00) > 1e-9 {
		t.Errorf("Scenario 4 - Full mutual attraction: expected score=1.00, got %.6f", scoreHarmonyFull)
	}

	// Grafo con 1 rechazo mutuo entre B y C (-1.0), pero A se lleva bien con ambos (+1.0):
	// Pares:
	// - (A, B): Tp = (1.0 + 1.0)/2 = 1.0 -> C'_AB = (1 - 1*1)/2 = 0.0
	// - (A, C): Tp = (1.0 + 1.0)/2 = 1.0 -> C'_AC = (1 - 1*1)/2 = 0.0
	// - (B, C): Tp = (-1.0 + -1.0)/2 = -1.0 -> C'_BC = (1 - 1*(-1))/2 = 1.0
	// Costo total = 0.0 + 0.0 + 1.0 = 1.0
	// Costo promedio = 1.0 / 3 = 0.333333
	// S_rel = 0.333333 -> Score = 1 - S_rel = 0.666667
	graphWithRejection := domain.SocialGraph{
		"A": {"B": 1.0, "C": 1.0},
		"B": {"A": 1.0, "C": -1.0},
		"C": {"A": 1.0, "B": -1.0},
	}
	ctxRejection := domain.Context{SocialGraph: graphWithRejection}
	scoreRejection := CalculateFitnessForGroup(&group3, configHarmony, factoryHarmony, ctxRejection)
	expectedRejection := 1.0 - (1.0 / 3.0) // 2/3 ≈ 0.666667
	if math.Abs(scoreRejection-expectedRejection) > 1e-9 {
		t.Errorf("Scenario 4 - Group with 1 rejection: expected score=%.6f, got %.6f", expectedRejection, scoreRejection)
	}

	// Invarianza al orden: permutar estudiantes C, A, B debe dar EXACTAMENTE el mismo score
	groupPermuted := domain.Group{
		makeStudent("C", nil, nil),
		makeStudent("A", nil, nil),
		makeStudent("B", nil, nil),
	}
	scorePermuted := CalculateFitnessForGroup(&groupPermuted, configHarmony, factoryHarmony, ctxRejection)
	if math.Abs(scorePermuted-scoreRejection) > 1e-9 {
		t.Errorf("Scenario 4 - Symmetry violation! Order affected score: %.6f vs %.6f", scoreRejection, scorePermuted)
	}

	// Subcaso 4B: Fricción Controlada / Diversidad (S_p = -1.0)
	// Con S_p = -1, el algoritmo premia el desacuerdo/fricción y penaliza la atracción ciega:
	// Para el par (B, C) con rechazo (-1.0): C'_BC = (1 - (-1)*(-1))/2 = 0.0 (¡perfecto para fricción!)
	frictionCriterion := NewSocialCriterion("social", -1.0)
	factoryFriction := NewCriterionFactory()
	factoryFriction.Register(frictionCriterion)
	configFriction := domain.Config{Weights: map[string]float64{"social": 1.0}}

	scoreFrictionRejection := CalculateFitnessForGroup(&group3, configFriction, factoryFriction, ctxRejection)
	// Para (A,B) y (A,C) con Tp=+1.0: C' = (1 - (-1)*1)/2 = 1.0
	// Para (B,C) con Tp=-1.0: C' = (1 - (-1)*(-1))/2 = 0.0
	// Costo total = 1.0 + 1.0 + 0.0 = 2.0 -> Costo promedio = 2/3 -> Score = 1 - 2/3 = 1/3 ≈ 0.333333
	expectedFriction := 1.0 / 3.0
	if math.Abs(scoreFrictionRejection-expectedFriction) > 1e-9 {
		t.Errorf("Scenario 4 - Controlled friction score: expected %.6f, got %.6f", expectedFriction, scoreFrictionRejection)
	}
}

// ============================================================================
// ESCENARIO 5: Característica Psicológica pero Estadística (Varianza / Dispersión)
// ============================================================================
// Evalúa un rasgo psicológico usando estrategias estadísticas: "variance",
// "homogeneity" (D_k = 1 - 1/(1+γσ²)), o "heterogeneity".
func TestAffinity_Scenario5_PsychologicalStatisticalCriterion(t *testing.T) {
	t.Parallel()

	// Subcaso 5A: Estrategia "variance" directa (Score = 1 - σ²)
	factoryVar := NewCriterionFactory()
	factoryVar.Register(NewNumericCriterion("neuroticism_var", "neuroticism", "variance"))

	configVar := domain.Config{Weights: map[string]float64{"neuroticism_var": 1.0}}
	ctx := domain.Context{}

	// Grupo A: Idéntico nivel de neuroticismo (σ² = 0.0) -> Score = 1.0
	groupHomogeneous := domain.Group{
		makeStudent("s1", map[string]float64{"neuroticism": 0.50}, nil),
		makeStudent("s2", map[string]float64{"neuroticism": 0.50}, nil),
		makeStudent("s3", map[string]float64{"neuroticism": 0.50}, nil),
	}
	scoreVarHom := CalculateFitnessForGroup(&groupHomogeneous, configVar, factoryVar, ctx)
	if math.Abs(scoreVarHom-1.00) > 1e-9 {
		t.Errorf("Scenario 5A - Homogeneous variance: expected 1.00, got %.6f", scoreVarHom)
	}

	// Grupo B: Niveles muy dispersos [0.1, 0.9] -> Media = 0.5, Varianza = ((0.4)² + (-0.4)²)/2 = 0.16
	// Score = 1 - 0.16 = 0.84
	groupDispersed := domain.Group{
		makeStudent("s4", map[string]float64{"neuroticism": 0.10}, nil),
		makeStudent("s5", map[string]float64{"neuroticism": 0.90}, nil),
	}
	scoreVarDisp := CalculateFitnessForGroup(&groupDispersed, configVar, factoryVar, ctx)
	if math.Abs(scoreVarDisp-0.84) > 1e-9 {
		t.Errorf("Scenario 5A - Dispersed variance: expected 0.84, got %.6f", scoreVarDisp)
	}

	// Subcaso 5B: Estrategia "homogeneity" con índice de diversidad de la tesis D_k(G)
	// D_k(G) = 1 - 1 / (1 + γ * σ²), donde γ = 25.0
	// φ_k(G) = |Tk - D_k(G)|, con Tk = 0.0 (homogeneidad)
	// Score = 1 - φ_k(G)
	factoryHomo := NewCriterionFactory()
	factoryHomo.Register(NewNumericCriterion("neuroticism_homo", "neuroticism", "homogeneity"))
	configHomo := domain.Config{
		Weights: map[string]float64{"neuroticism_homo": 1.0},
		Targets: map[string]float64{"neuroticism": 0.0},
	}
	ctxHomo := domain.Context{Targets: configHomo.Targets}

	// Grupo homogéneo: σ² = 0 -> D_k = 0 -> Score = 1.0
	scoreHomo := CalculateFitnessForGroup(&groupHomogeneous, configHomo, factoryHomo, ctxHomo)
	if math.Abs(scoreHomo-1.00) > 1e-9 {
		t.Errorf("Scenario 5B - Thesis homogeneity: expected 1.00, got %.6f", scoreHomo)
	}

	// Subcaso 5C: Estrategia "heterogeneity" (Tk = 1.0)
	// Score = D_k(G) = 1 - 1 / (1 + 25 * 0.16) = 1 - 1 / 5.0 = 0.80
	factoryHetero := NewCriterionFactory()
	factoryHetero.Register(NewNumericCriterion("neuroticism_hetero", "neuroticism", "heterogeneity"))
	configHetero := domain.Config{
		Weights: map[string]float64{"neuroticism_hetero": 1.0},
		Targets: map[string]float64{"neuroticism": 1.0},
	}
	ctxHetero := domain.Context{Targets: configHetero.Targets}
	scoreHetero := CalculateFitnessForGroup(&groupDispersed, configHetero, factoryHetero, ctxHetero)
	if math.Abs(scoreHetero-0.80) > 1e-9 {
		t.Errorf("Scenario 5C - Thesis heterogeneity: expected 0.80, got %.6f", scoreHetero)
	}
}

// ============================================================================
// ESCENARIO 6: Característica Categórica (Índice de Blau Normalizado)
// ============================================================================
// Evalúa categorías (ej. career_code o eneatipo) mediante el índice de Blau:
// B = 1 - Σ p_i², normalizado por k_global: maxBlau = 1 - 1/k_global
// Score = 1 - |Target - NormalizedBlau|
func TestAffinity_Scenario6_CategoricalCriterion(t *testing.T) {
	t.Parallel()

	factory := NewCriterionFactory()
	factory.Register(NewCategoricalCriterion("career_code", "career_code"))

	config := domain.Config{
		Weights: map[string]float64{"career_code": 1.0},
		Targets: map[string]float64{"career_code": 0.50}, // Buscamos diversidad moderada (target = 0.50)
	}
	// Población con 4 carreras (k_global = 4) -> maxBlau = 1 - 1/4 = 0.75
	ctx := domain.Context{
		Targets:        config.Targets,
		CategoryCounts: map[string]int{"career_code": 4},
	}

	// Grupo A: Todos de la misma carrera ("Computación")
	// Blau = 1 - 1² = 0.0 -> NormalizedBlau = 0.0 / 0.75 = 0.0
	// Target = 0.50 -> Score = 1 - |0.50 - 0.0| = 0.50
	groupAllSame := domain.Group{
		makeStudent("s1", nil, map[string]string{"career_code": "Computación"}),
		makeStudent("s2", nil, map[string]string{"career_code": "Computación"}),
		makeStudent("s3", nil, map[string]string{"career_code": "Computación"}),
		makeStudent("s4", nil, map[string]string{"career_code": "Computación"}),
	}
	scoreAllSame := CalculateFitnessForGroup(&groupAllSame, config, factory, ctx)
	if math.Abs(scoreAllSame-0.50) > 1e-9 {
		t.Errorf("Scenario 6 - All same category: expected 0.50, got %.6f", scoreAllSame)
	}

	// Grupo B: 2 de Computación, 2 de Civil (2 categorías equiprobables)
	// p_comp = 0.5, p_civil = 0.5 -> Blau = 1 - (0.25 + 0.25) = 0.50
	// NormalizedBlau = 0.50 / 0.75 = 2/3 ≈ 0.666667
	// Score = 1 - |0.50 - 0.666667| = 1 - 0.166667 = 0.833333
	groupBalanced2 := domain.Group{
		makeStudent("s1", nil, map[string]string{"career_code": "Computación"}),
		makeStudent("s2", nil, map[string]string{"career_code": "Computación"}),
		makeStudent("s3", nil, map[string]string{"career_code": "Civil"}),
		makeStudent("s4", nil, map[string]string{"career_code": "Civil"}),
	}
	scoreBalanced2 := CalculateFitnessForGroup(&groupBalanced2, config, factory, ctx)
	expectedBalanced2 := 1.0 - math.Abs(0.50-(0.50/0.75)) // ≈ 0.833333
	if math.Abs(scoreBalanced2-expectedBalanced2) > 1e-9 {
		t.Errorf("Scenario 6 - Balanced 2 categories: expected %.6f, got %.6f", expectedBalanced2, scoreBalanced2)
	}

	// Ahora si el target fuese exactamente 1.0 (máxima diversidad deseada):
	// Con un grupo de 4 estudiantes cada uno de una carrera distinta:
	// p = 0.25 cada uno -> Blau = 1 - 4*(0.0625) = 0.75
	// NormalizedBlau = 0.75 / 0.75 = 1.00
	// Score = 1 - |1.00 - 1.00| = 1.00 (perfecto)
	ctxMaxDiv := domain.Context{
		Targets:        map[string]float64{"career_code": 1.00},
		CategoryCounts: map[string]int{"career_code": 4},
	}
	configMaxDiv := domain.Config{
		Weights: map[string]float64{"career_code": 1.0},
		Targets: ctxMaxDiv.Targets,
	}
	groupAllDistinct := domain.Group{
		makeStudent("s1", nil, map[string]string{"career_code": "Computación"}),
		makeStudent("s2", nil, map[string]string{"career_code": "Civil"}),
		makeStudent("s3", nil, map[string]string{"career_code": "Electricidad"}),
		makeStudent("s4", nil, map[string]string{"career_code": "Comercial"}),
	}
	scoreAllDistinct := CalculateFitnessForGroup(&groupAllDistinct, configMaxDiv, factory, ctxMaxDiv)
	if math.Abs(scoreAllDistinct-1.00) > 1e-9 {
		t.Errorf("Scenario 6 - Maximum diversity group: expected 1.00, got %.6f", scoreAllDistinct)
	}
}

// ============================================================================
// ESCENARIO 7: Formación Completa de Equipos con Algoritmo Genético
// ============================================================================
// Verifica que el motor genético (Engine) forma grupos optimizando la afinidad
// combinada de múltiples características y que las métricas finales son exactas.
func TestAffinity_Scenario7_EndToEndGeneticEngine(t *testing.T) {
	t.Parallel()

	factory := NewCriterionFactory()
	factory.Register(NewNumericCriterion("openness", "openness", "target"))
	factory.Register(NewCategoricalCriterion("career_code", "career_code"))

	// Configuración mixta: 50% openness, 50% career_code
	config := domain.Config{
		Weights: map[string]float64{
			"openness":    0.5,
			"career_code": 0.5,
		},
		Targets: map[string]float64{
			"openness":    0.70, // target apertura 0.70
			"career_code": 0.80, // target diversidad alta
		},
	}
	ctx := domain.Context{
		Targets:        config.Targets,
		CategoryCounts: map[string]int{"career_code": 3},
	}

	// 12 estudiantes divididos en 3 carreras
	students := []*domain.Student{
		makeStudent("1", map[string]float64{"openness": 0.70}, map[string]string{"career_code": "A"}),
		makeStudent("2", map[string]float64{"openness": 0.70}, map[string]string{"career_code": "B"}),
		makeStudent("3", map[string]float64{"openness": 0.70}, map[string]string{"career_code": "C"}),
		makeStudent("4", map[string]float64{"openness": 0.70}, map[string]string{"career_code": "A"}),
		makeStudent("5", map[string]float64{"openness": 0.70}, map[string]string{"career_code": "B"}),
		makeStudent("6", map[string]float64{"openness": 0.70}, map[string]string{"career_code": "C"}),
		makeStudent("7", map[string]float64{"openness": 0.70}, map[string]string{"career_code": "A"}),
		makeStudent("8", map[string]float64{"openness": 0.70}, map[string]string{"career_code": "B"}),
		makeStudent("9", map[string]float64{"openness": 0.70}, map[string]string{"career_code": "C"}),
		makeStudent("10", map[string]float64{"openness": 0.70}, map[string]string{"career_code": "A"}),
		makeStudent("11", map[string]float64{"openness": 0.70}, map[string]string{"career_code": "B"}),
		makeStudent("12", map[string]float64{"openness": 0.70}, map[string]string{"career_code": "C"}),
	}

	engineConfig := EngineConfig{
		PopulationSize:  50,
		Generations:     30,
		ElitismRate:     0.15,
		MutationRate:    0.2,
		CrossoverRate:   0.8,
		TournamentSize:  3,
		TargetGroupSize: 4, // 3 grupos de 4
		MaxStagnation:   15,
		TargetFitness:   0.98,
		Seed:            42,
	}

	engine := NewEngine(engineConfig, factory, config, ctx)
	bestChromosome, stats, err := engine.Run(context.Background(), students)
	if err != nil {
		t.Fatalf("Engine run failed: %v", err)
	}
	if bestChromosome == nil {
		t.Fatal("Engine returned nil chromosome")
	}

	// 1. Verificar número de grupos formados (12 estudiantes / tamaño 4 = 3 grupos)
	if len(bestChromosome.Groups) != 3 {
		t.Errorf("Expected 3 groups, got %d", len(bestChromosome.Groups))
	}

	// 2. Verificar que todos los 12 estudiantes están asignados exactamente una vez
	assigned := make(map[string]int)
	for _, g := range bestChromosome.Groups {
		for _, s := range g {
			assigned[s.ID]++
		}
	}
	if len(assigned) != 12 {
		t.Errorf("Expected 12 unique students assigned, got %d", len(assigned))
	}
	for id, count := range assigned {
		if count != 1 {
			t.Errorf("Student %s assigned %d times (expected 1)", id, count)
		}
	}

	// 3. Verificar que cada grupo tiene tamaño 4
	for i, g := range bestChromosome.Groups {
		if len(g) != 4 {
			t.Errorf("Group %d has size %d (expected 4)", i, len(g))
		}
		// Verificar que el cálculo de afinidad del grupo es consistente
		groupScore := CalculateFitnessForGroup(&g, config, factory, ctx)
		if groupScore < 0.70 {
			t.Errorf("Group %d affinity score unexpectedly low: %.4f", i, groupScore)
		}
	}

	// 4. Verificar que el fitness global coincide con el promedio de scores de los grupos
	totalScores := 0.0
	for _, g := range bestChromosome.Groups {
		totalScores += CalculateFitnessForGroup(&g, config, factory, ctx)
	}
	expectedGlobal := totalScores / float64(len(bestChromosome.Groups))
	if math.Abs(bestChromosome.Fitness-expectedGlobal) > 1e-9 {
		t.Errorf("Global fitness mismatch: chromosome=%.6f vs mean(group_scores)=%.6f",
			bestChromosome.Fitness, expectedGlobal)
	}

	t.Logf("Scenario 7 Success - Formed 3 groups with global fitness: %.4f across %d generations",
		bestChromosome.Fitness, len(stats.Generations))
}
