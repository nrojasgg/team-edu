package genetic

import (
	"math"
	"testing"

	"ms-core/internal/core/domain"
)

// ─── evaluate-groups parity regression ───────────────────────────────────────
//
// Reproduces the exact payload from the user's report:
//   - form-flexible used goal="heterogeneity" for conscientiousness
//   - evaluate-groups was receiving goal="" which fell through to "target"
//     strategy, producing completely different scores.
//
// These tests assert that CalculateFitnessForGroup with the correct strategies
// reproduces the scores reported by form-flexible, and that using goal=""
// (which maps to "target") produces the wrong scores seen in evaluate-groups.

// studentData mirrors the demographic_data from the user's request.
var evaluateGroupsRegressionStudents = map[string]struct {
	conscientiousness float64
	careerCode        string
}{
	"38": {0.7, "Ingeniería en Computación"},
	"39": {0.25, "Ingeniería en Computación"},
	"40": {0.65, "Ingeniería en Computación"},
	"41": {0.4, "Ingeniería en Computación"},
	"42": {0.95, "Ingeniería en Computación"},
	"43": {0.5, "Ingeniería en Computación"},
	"44": {0.85, "Ingeniería en Computación"},
	"45": {0.3, "Ingeniería en Computación"},
	"46": {0.6, "Ingeniería en Electricidad"},
	"47": {0.7, "Ingeniería en Electricidad"},
	"48": {0.25, "Ingeniería en Electricidad"},
	"49": {0.65, "Ingeniería en Electricidad"},
	"50": {0.4, "Ingeniería en Electricidad"},
	"51": {0.95, "Ingeniería en Electricidad"},
	"52": {0.5, "Ingeniería Civil"},
	"53": {0.85, "Ingeniería Civil"},
	"54": {0.3, "Ingeniería Civil"},
	"55": {0.6, "Ingeniería Civil"},
	"56": {0.7, "Ingeniería Civil"},
	"57": {0.25, "Ingeniería Civil"},
	"58": {0.65, "Ingeniería Comercial"},
	"59": {0.4, "Ingeniería Comercial"},
	"60": {0.95, "Ingeniería Comercial"},
	"61": {0.5, "Ingeniería Comercial"},
	"62": {0.85, "Ingeniería Comercial"},
	"63": {0.3, "Ingeniería Comercial"},
}

func buildRegressionStudent(id string) *domain.Student {
	d := evaluateGroupsRegressionStudents[id]
	return &domain.Student{
		ID:               id,
		NumericAttrs:     map[string]float64{"conscientiousness": d.conscientiousness},
		CategoricalAttrs: map[string]string{"career_code": d.careerCode},
	}
}

func buildRegressionGroup(ids ...string) domain.Group {
	g := make(domain.Group, len(ids))
	for i, id := range ids {
		g[i] = buildRegressionStudent(id)
	}
	return g
}

// buildRegressionContext builds the evaluation context replicating the mapper:
// kGlobal = 4 (four distinct career codes in the full population of 26 students).
func buildRegressionContext(target float64) domain.Context {
	return domain.Context{
		Targets: map[string]float64{
			"career_code":       target,
			"conscientiousness": target,
		},
		CategoryCounts: map[string]int{
			"career_code": 4, // Computación, Electricidad, Civil, Comercial
		},
		SocialGraph: domain.SocialGraph{},
	}
}

// buildRegressionConfig builds Config with equal weights 0.5/0.5 (normalised).
func buildRegressionConfig() domain.Config {
	return domain.Config{
		Weights: map[string]float64{
			"career_code":       0.5,
			"conscientiousness": 0.5,
		},
		Targets: map[string]float64{
			"career_code":       0.5,
			"conscientiousness": 0.9,
		},
	}
}

// buildRegressionFactory builds the factory with the correct strategies:
//   - career_code   → categorical (Blau)
//   - conscientiousness → numeric, strategy from caller
func buildRegressionFactory(conscientiousnessStrategy string) *CriterionFactory {
	f := NewCriterionFactory()
	f.Register(NewCategoricalCriterion("career_code", "career_code"))
	f.Register(NewNumericCriterion("conscientiousness", "conscientiousness", conscientiousnessStrategy))
	return f
}

// TestEvaluateGroupsParity_Group1_HeterogeneityMatchesFormFlexible asserts that
// group 1 [61,45,40,42] with strategy="heterogeneity" reproduces the
// affinity_score ≈ 0.8422 reported by the thesis diversity formula (gamma=25.0).
func TestEvaluateGroupsParity_Group1_HeterogeneityMatchesFormFlexible(t *testing.T) {
	t.Parallel()

	group := buildRegressionGroup("61", "45", "40", "42")
	ctx := buildRegressionContext(0.5)
	ctx.Targets["conscientiousness"] = 0.9
	config := buildRegressionConfig()
	factory := buildRegressionFactory("heterogeneity")

	score := CalculateFitnessForGroup(&group, config, factory, ctx)

	want := 0.8422077922077922
	if math.Abs(score-want) > 1e-9 {
		t.Errorf("group1 heterogeneity score = %.10f, want ≈ %.10f", score, want)
	}
}

// TestEvaluateGroupsParity_Group1_TargetProducesWrongScore asserts that using
// strategy="target" (the bug: goal="" maps to "target") reproduces the wrong
// affinity_score=0.85 that evaluate-groups was returning before the fix.
func TestEvaluateGroupsParity_Group1_TargetProducesWrongScore(t *testing.T) {
	t.Parallel()

	group := buildRegressionGroup("61", "45", "40", "42")
	ctx := buildRegressionContext(0.5)
	ctx.Targets["conscientiousness"] = 0.9
	config := buildRegressionConfig()
	factory := buildRegressionFactory("target") // bug: goal="" → "target"

	score := CalculateFitnessForGroup(&group, config, factory, ctx)

	// evaluate-groups (buggy) reported 0.8500000000000001
	if math.Abs(score-0.85) > 1e-9 {
		t.Errorf("group1 target (buggy) score = %.10f, want ≈ 0.85", score)
	}
}

// TestEvaluateGroupsParity_AllGroups checks all 6 groups from the report.
func TestEvaluateGroupsParity_AllGroups(t *testing.T) {
	t.Parallel()

	type groupCase struct {
		name      string
		memberIDs []string
		wantScore float64 // thesis formula score (gamma=25.0)
	}

	cases := []groupCase{
		{"group1", []string{"61", "45", "40", "42"}, 0.8422077922077922},
		{"group2", []string{"56", "63", "59", "62"}, 0.8258318739054291},
		{"group3", []string{"53", "46", "47", "48"}, 0.8246478873239436},
		{"group4", []string{"58", "39", "44", "38"}, 0.8258318739054291},
		{"group5", []string{"43", "50", "41", "51", "49"}, 0.7378692493946730},
		{"group6", []string{"52", "54", "55", "57", "60"}, 0.8184015594541909},
	}

	ctx := buildRegressionContext(0.5)
	ctx.Targets["conscientiousness"] = 0.9
	config := buildRegressionConfig()
	factory := buildRegressionFactory("heterogeneity")

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			group := buildRegressionGroup(tc.memberIDs...)
			score := CalculateFitnessForGroup(&group, config, factory, ctx)
			if math.Abs(score-tc.wantScore) > 1e-9 {
				t.Errorf("%s: score = %.13f, want %.13f (diff=%.2e)",
					tc.name, score, tc.wantScore, math.Abs(score-tc.wantScore))
			}
		})
	}
}

// TestEvaluateGroupsParity_GlobalScore checks that the mean of the 6 group
// scores matches the global_score reported by the thesis formula.
func TestEvaluateGroupsParity_GlobalScore(t *testing.T) {
	t.Parallel()

	groups := [][]string{
		{"61", "45", "40", "42"},
		{"56", "63", "59", "62"},
		{"53", "46", "47", "48"},
		{"58", "39", "44", "38"},
		{"43", "50", "41", "51", "49"},
		{"52", "54", "55", "57", "60"},
	}

	ctx := buildRegressionContext(0.5)
	ctx.Targets["conscientiousness"] = 0.9
	config := buildRegressionConfig()
	factory := buildRegressionFactory("heterogeneity")

	total := 0.0
	for _, ids := range groups {
		g := buildRegressionGroup(ids...)
		total += CalculateFitnessForGroup(&g, config, factory, ctx)
	}
	globalScore := total / float64(len(groups))

	want := 0.8124650393652429
	if math.Abs(globalScore-want) > 1e-9 {
		t.Errorf("global score = %.13f, want %.13f", globalScore, want)
	}
}

// ─── minimize / maximize / default strategies ────────────────────────────────

func TestNumericCriterion_Minimize(t *testing.T) {
	criterion := NewNumericCriterion("neuroticism", "neuroticism", "minimize")
	ctx := domain.Context{}

	tests := []struct {
		name      string
		values    []float64
		wantAbove float64 // score must be >= this
		wantBelow float64 // score must be <= this
	}{
		{
			name: "low mean → high score (emotionally stable group)",
			// mean ≈ 0.1 → score ≈ 0.9
			values:    []float64{0.1, 0.1, 0.1, 0.1},
			wantAbove: 0.85,
			wantBelow: 1.0,
		},
		{
			name: "high mean → low score (neurotic group)",
			// mean ≈ 0.9 → score ≈ 0.1
			values:    []float64{0.9, 0.9, 0.9, 0.9},
			wantAbove: 0.0,
			wantBelow: 0.15,
		},
		{
			name:      "mixed group: lower mean beats higher mean",
			values:    []float64{0.2, 0.3, 0.4},
			wantAbove: 0.5,
			wantBelow: 1.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			group := make(domain.Group, len(tc.values))
			for i, v := range tc.values {
				group[i] = &domain.Student{
					ID:           string(rune('A' + i)),
					NumericAttrs: map[string]float64{"neuroticism": v},
				}
			}
			score := criterion.Calculate(&group, ctx)
			if score < tc.wantAbove || score > tc.wantBelow {
				t.Errorf("score %.4f not in [%.2f, %.2f]", score, tc.wantAbove, tc.wantBelow)
			}
		})
	}
}

func TestNumericCriterion_Minimize_ScoreMonotonicallyDecreasesWithMean(t *testing.T) {
	criterion := NewNumericCriterion("neuroticism", "neuroticism", "minimize")
	ctx := domain.Context{}

	means := []float64{0.0, 0.25, 0.5, 0.75, 1.0}
	prev := 2.0
	for _, m := range means {
		group := domain.Group{
			&domain.Student{ID: "1", NumericAttrs: map[string]float64{"neuroticism": m}},
		}
		score := criterion.Calculate(&group, ctx)
		if score > prev {
			t.Errorf("score %.4f at mean %.2f is higher than previous score %.4f (not monotonically decreasing)", score, m, prev)
		}
		prev = score
	}
}

func TestNumericCriterion_Maximize(t *testing.T) {
	criterion := NewNumericCriterion("agreeableness", "agreeableness", "maximize")
	ctx := domain.Context{}

	tests := []struct {
		name      string
		values    []float64
		wantAbove float64
		wantBelow float64
	}{
		{
			name: "high mean → high score (very agreeable group)",
			// mean ≈ 0.9 → score ≈ 0.9
			values:    []float64{0.9, 0.9, 0.9, 0.9},
			wantAbove: 0.85,
			wantBelow: 1.0,
		},
		{
			name: "low mean → low score (disagreeable group)",
			// mean ≈ 0.1 → score ≈ 0.1
			values:    []float64{0.1, 0.1, 0.1, 0.1},
			wantAbove: 0.0,
			wantBelow: 0.15,
		},
		{
			name:      "mixed group: score equals mean",
			values:    []float64{0.4, 0.6},
			wantAbove: 0.49,
			wantBelow: 0.51,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			group := make(domain.Group, len(tc.values))
			for i, v := range tc.values {
				group[i] = &domain.Student{
					ID:           string(rune('A' + i)),
					NumericAttrs: map[string]float64{"agreeableness": v},
				}
			}
			score := criterion.Calculate(&group, ctx)
			if score < tc.wantAbove || score > tc.wantBelow {
				t.Errorf("score %.4f not in [%.2f, %.2f]", score, tc.wantAbove, tc.wantBelow)
			}
		})
	}
}

func TestNumericCriterion_Maximize_ScoreMonotonicallyIncreasesWithMean(t *testing.T) {
	criterion := NewNumericCriterion("agreeableness", "agreeableness", "maximize")
	ctx := domain.Context{}

	means := []float64{0.0, 0.25, 0.5, 0.75, 1.0}
	prev := -1.0
	for _, m := range means {
		group := domain.Group{
			&domain.Student{ID: "1", NumericAttrs: map[string]float64{"agreeableness": m}},
		}
		score := criterion.Calculate(&group, ctx)
		if score < prev {
			t.Errorf("score %.4f at mean %.2f is lower than previous score %.4f (not monotonically increasing)", score, m, prev)
		}
		prev = score
	}
}

func TestNumericCriterion_MinimizeMaximize_AreOpposites(t *testing.T) {
	minimize := NewNumericCriterion("neuroticism", "neuroticism", "minimize")
	maximize := NewNumericCriterion("neuroticism", "neuroticism", "maximize")
	ctx := domain.Context{}

	values := []float64{0.3, 0.5, 0.7}
	group := make(domain.Group, len(values))
	for i, v := range values {
		group[i] = &domain.Student{
			ID:           string(rune('A' + i)),
			NumericAttrs: map[string]float64{"neuroticism": v},
		}
	}

	sMin := minimize.Calculate(&group, ctx)
	sMax := maximize.Calculate(&group, ctx)

	if math.Abs((sMin+sMax)-1.0) > 1e-9 {
		t.Errorf("minimize score %.6f + maximize score %.6f should sum to 1.0", sMin, sMax)
	}
}

func TestNumericCriterion_DefaultStrategy_FallsBackToTarget(t *testing.T) {
	// An unknown strategy string should fall back to "target" behaviour.
	unknown := NewNumericCriterion("openness", "openness", "unknown_strategy")
	target := NewNumericCriterion("openness", "openness", "target")
	ctx := domain.Context{
		Targets: map[string]float64{"openness": 0.6},
	}

	group := domain.Group{
		&domain.Student{ID: "1", NumericAttrs: map[string]float64{"openness": 0.5}},
		&domain.Student{ID: "2", NumericAttrs: map[string]float64{"openness": 0.7}},
	}

	sUnknown := unknown.Calculate(&group, ctx)
	sTarget := target.Calculate(&group, ctx)

	if math.Abs(sUnknown-sTarget) > 1e-9 {
		t.Errorf("unknown strategy score %.6f should equal target score %.6f", sUnknown, sTarget)
	}
}

func TestNumericCriterion_MissingAttribute_ReturnsNeutral(t *testing.T) {
	for _, strategy := range []string{"minimize", "maximize", "target", "homogeneity", "heterogeneity"} {
		t.Run(strategy, func(t *testing.T) {
			criterion := NewNumericCriterion("neuroticism", "neuroticism", strategy)
			ctx := domain.Context{Targets: map[string]float64{"neuroticism": 0.5}}

			// Group has no "neuroticism" key at all
			group := domain.Group{
				&domain.Student{ID: "1", NumericAttrs: map[string]float64{"other": 0.5}},
				&domain.Student{ID: "2", NumericAttrs: map[string]float64{"other": 0.8}},
			}
			score := criterion.Calculate(&group, ctx)
			if score != 0.5 {
				t.Errorf("[%s] expected neutral 0.5 for missing attr, got %.4f", strategy, score)
			}
		})
	}
}

func TestNumericCriterion_Homogeneity(t *testing.T) {
	criterion := NewNumericCriterion("age_balance", "age", "homogeneity")

	// Test case 1: Homogeneous group (low variance = high score)
	group := domain.Group{
		{ID: "1", NumericAttrs: map[string]float64{"age": 20}},
		{ID: "2", NumericAttrs: map[string]float64{"age": 21}},
		{ID: "3", NumericAttrs: map[string]float64{"age": 20}},
	}

	ctx := domain.Context{}
	score := criterion.Calculate(&group, ctx)

	if score < 0.10 {
		t.Errorf("Expected reasonable score for homogeneous group, got %f", score)
	}

	// Test case 2: Heterogeneous group (high variance = low score)
	group2 := domain.Group{
		{ID: "1", NumericAttrs: map[string]float64{"age": 18}},
		{ID: "2", NumericAttrs: map[string]float64{"age": 25}},
		{ID: "3", NumericAttrs: map[string]float64{"age": 35}},
	}

	score2 := criterion.Calculate(&group2, ctx)

	if score2 >= score {
		t.Errorf("Expected lower score for heterogeneous group, got %f vs %f", score2, score)
	}
}

func TestNumericCriterion_Heterogeneity(t *testing.T) {
	criterion := NewNumericCriterion("extraversion", "extraversion", "heterogeneity")

	// Test case: Moderate diversity (close to target variance)
	group := domain.Group{
		{ID: "1", NumericAttrs: map[string]float64{"extraversion": 0.3}},
		{ID: "2", NumericAttrs: map[string]float64{"extraversion": 0.5}},
		{ID: "3", NumericAttrs: map[string]float64{"extraversion": 0.7}},
	}

	ctx := domain.Context{
		Targets: map[string]float64{"extraversion_variance": 0.06},
	}

	score := criterion.Calculate(&group, ctx)

	if score < 0.35 || score > 1.0 {
		t.Errorf("Expected moderate score for heterogeneous criterion, got %f", score)
	}
}

// ============================================================================
// EnforceEquity — CalculateFitness tests (minimax semantics)
// ============================================================================

// perfectGroup returns a group of two students with identical GPA → score ≈ 1.0.
func perfectGroup(id1, id2 string) domain.Group {
	return domain.Group{
		{ID: id1, NumericAttrs: map[string]float64{"gpa": 0.8}},
		{ID: id2, NumericAttrs: map[string]float64{"gpa": 0.8}},
	}
}

// trashGroup returns a group with maximally spread GPA → score ≈ 0 (homogeneity criterion).
func trashGroup(id1, id2 string) domain.Group {
	return domain.Group{
		{ID: id1, NumericAttrs: map[string]float64{"gpa": 0.0}},
		{ID: id2, NumericAttrs: map[string]float64{"gpa": 1.0}},
	}
}

// gpaFactory returns a CriterionFactory with the gpa_balance homogeneity criterion registered.
func gpaFactory() *CriterionFactory {
	f := NewCriterionFactory()
	f.Register(NewNumericCriterion("gpa_balance", "gpa", "homogeneity"))
	return f
}

// TestCalculateFitness_EnforceEquity_False verifies that without equity enforcement
// the fitness is the plain average of per-team scores (legacy behaviour).
func TestCalculateFitness_EnforceEquity_False(t *testing.T) {
	t.Parallel()

	chromosome := &domain.Chromosome{
		Groups: []domain.Group{perfectGroup("a", "b"), perfectGroup("c", "d")},
		Stats:  map[string]interface{}{},
	}
	config := domain.Config{
		Weights:        map[string]float64{"gpa_balance": 1.0},
		EquityStrategy: "",
	}
	ctx := domain.Context{}

	fitness := CalculateFitness(chromosome, config, gpaFactory(), ctx)

	// Both groups are perfectly homogeneous → average ≈ 1.0.
	if fitness < 0.9 {
		t.Errorf("expected fitness ≈ 1.0 for two uniform groups, got %.4f", fitness)
	}
}

// TestCalculateFitness_EnforceEquity_ReturnsWorstTeam verifies that with
// EnforceEquity=true the fitness equals the score of the worst team, not the
// average.  A solution with one perfect team and one trash team must score as
// low as the trash team — forcing the algorithm to lift the floor.
func TestCalculateFitness_EnforceEquity_ReturnsWorstTeam(t *testing.T) {
	t.Parallel()

	chromosome := &domain.Chromosome{
		Groups: []domain.Group{perfectGroup("a", "b"), trashGroup("c", "d")},
		Stats:  map[string]interface{}{},
	}
	factory := gpaFactory()
	ctx := domain.Context{}

	config := domain.Config{
		Weights:        map[string]float64{"gpa_balance": 1.0},
		EquityStrategy: "",
	}
	plainFitness := CalculateFitness(chromosome, config, factory, ctx)

	config.EquityStrategy = "minimax"
	equityFitness := CalculateFitness(chromosome, config, factory, ctx)

	// With equity the score must equal the worst team's score, which is well
	// below the plain average.
	if equityFitness >= plainFitness {
		t.Errorf(
			"minimax fitness (%.4f) should be lower than plain average (%.4f)",
			equityFitness, plainFitness,
		)
	}

	// Equity fitness must be ≤ the trash team score (which is ≈ 0).
	trashScore := CalculateFitnessForGroup(&chromosome.Groups[1], config, factory, ctx)
	if equityFitness > trashScore+1e-9 {
		t.Errorf(
			"minimax fitness (%.4f) must equal worst-team score (%.4f)",
			equityFitness, trashScore,
		)
	}
}

// TestCalculateFitness_EnforceEquity_UniformGroups verifies that when every team
// has the same score, minimax == average, so equity enforcement is a no-op.
func TestCalculateFitness_EnforceEquity_UniformGroups(t *testing.T) {
	t.Parallel()

	chromosome := &domain.Chromosome{
		Groups: []domain.Group{
			perfectGroup("a", "b"),
			perfectGroup("c", "d"),
			perfectGroup("e", "f"),
		},
		Stats: map[string]interface{}{},
	}
	factory := gpaFactory()
	ctx := domain.Context{}

	config := domain.Config{
		Weights:        map[string]float64{"gpa_balance": 1.0},
		EquityStrategy: "",
	}
	plain := CalculateFitness(chromosome, config, factory, ctx)

	config.EquityStrategy = "minimax"
	withEquity := CalculateFitness(chromosome, config, factory, ctx)

	// min == mean when all scores are identical.
	if math.Abs(plain-withEquity) > 0.05 {
		t.Errorf(
			"expected equal fitness for uniform groups (plain=%.4f, equity=%.4f)",
			plain, withEquity,
		)
	}
}

// TestCalculateFitness_EnforceEquity_NeverNegative ensures the minimax result
// is clamped to [0, 1] even when individual team scores are 0.
func TestCalculateFitness_EnforceEquity_NeverNegative(t *testing.T) {
	t.Parallel()

	chromosome := &domain.Chromosome{
		Groups: []domain.Group{perfectGroup("a", "b"), trashGroup("c", "d")},
		Stats:  map[string]interface{}{},
	}
	config := domain.Config{
		Weights:        map[string]float64{"gpa_balance": 1.0},
		EquityStrategy: "minimax",
	}
	ctx := domain.Context{}

	fitness := CalculateFitness(chromosome, config, gpaFactory(), ctx)
	if fitness < 0 {
		t.Errorf("fitness must never be negative, got %.4f", fitness)
	}
	if fitness > 1 {
		t.Errorf("fitness must never exceed 1.0, got %.4f", fitness)
	}
}

// TestCalculateFitness_EnforceEquity_MinimaxPicksLowest verifies that with three
// teams of different scores the minimax fitness equals the lowest one exactly.
func TestCalculateFitness_EnforceEquity_MinimaxPicksLowest(t *testing.T) {
	t.Parallel()

	// Three groups with predictable scores:
	//   high  → identical GPA 0.8      → score ≈ 1.0
	//   mid   → GPA spread 0.4–0.6     → intermediate score
	//   low   → GPA spread 0.0–1.0     → score ≈ 0
	highGroup := domain.Group{
		{ID: "a", NumericAttrs: map[string]float64{"gpa": 0.8}},
		{ID: "b", NumericAttrs: map[string]float64{"gpa": 0.8}},
	}
	midGroup := domain.Group{
		{ID: "c", NumericAttrs: map[string]float64{"gpa": 0.4}},
		{ID: "d", NumericAttrs: map[string]float64{"gpa": 0.6}},
	}
	lowGroup := domain.Group{
		{ID: "e", NumericAttrs: map[string]float64{"gpa": 0.0}},
		{ID: "f", NumericAttrs: map[string]float64{"gpa": 1.0}},
	}

	chromosome := &domain.Chromosome{
		Groups: []domain.Group{highGroup, midGroup, lowGroup},
		Stats:  map[string]interface{}{},
	}
	factory := gpaFactory()
	ctx := domain.Context{}
	config := domain.Config{
		Weights:        map[string]float64{"gpa_balance": 1.0},
		EquityStrategy: "minimax",
	}

	equityFitness := CalculateFitness(chromosome, config, factory, ctx)

	// Compute the actual score of the low group independently.
	lowGroupCopy := make(domain.Group, len(lowGroup))
	copy(lowGroupCopy, lowGroup)
	expectedWorst := CalculateFitnessForGroup(&lowGroupCopy, config, factory, ctx)

	if math.Abs(equityFitness-expectedWorst) > 1e-9 {
		t.Errorf(
			"minimax fitness (%.6f) should equal worst-team score (%.6f)",
			equityFitness, expectedWorst,
		)
	}
}

// ============================================================================
// ComputeEquityStats tests
// ============================================================================

// TestComputeEquityStats_Basic verifies that the equity stats correctly identify
// the lowest/highest team scores and a non-zero stddev for unbalanced groups.
func TestComputeEquityStats_Basic(t *testing.T) {
	t.Parallel()

	perfectGroup := domain.Group{
		{ID: "a", NumericAttrs: map[string]float64{"gpa": 0.8}},
		{ID: "b", NumericAttrs: map[string]float64{"gpa": 0.8}},
	}
	trashGroup := domain.Group{
		{ID: "c", NumericAttrs: map[string]float64{"gpa": 0.0}},
		{ID: "d", NumericAttrs: map[string]float64{"gpa": 1.0}},
	}
	chromosome := &domain.Chromosome{
		Groups: []domain.Group{perfectGroup, trashGroup},
		Stats:  map[string]interface{}{},
	}
	factory := NewCriterionFactory()
	factory.Register(NewNumericCriterion("gpa_balance", "gpa", "homogeneity"))

	config := domain.Config{
		Weights:        map[string]float64{"gpa_balance": 1.0},
		EquityStrategy: "minimax",
	}
	ctx := domain.Context{}

	stats := ComputeEquityStats(chromosome, config, factory, ctx)

	if !stats.EquityEnforced {
		t.Error("expected EquityEnforced=true")
	}
	if stats.LowestTeamScore >= stats.HighestTeamScore {
		t.Errorf(
			"expected lowest (%.4f) < highest (%.4f)",
			stats.LowestTeamScore, stats.HighestTeamScore,
		)
	}
	if stats.ScoreStdDev <= 0 {
		t.Errorf("expected positive stddev for unbalanced groups, got %.4f", stats.ScoreStdDev)
	}
	if stats.ScoreVariance <= 0 {
		t.Errorf("expected positive variance for unbalanced groups, got %.4f", stats.ScoreVariance)
	}
	// Variance should equal StdDev²
	expectedVariance := stats.ScoreStdDev * stats.ScoreStdDev
	if math.Abs(stats.ScoreVariance-expectedVariance) > 1e-9 {
		t.Errorf("variance (%.6f) != stddev² (%.6f)", stats.ScoreVariance, expectedVariance)
	}
}

// TestComputeEquityStats_UniformGroups verifies zero stddev when all teams are equal.
func TestComputeEquityStats_UniformGroups(t *testing.T) {
	t.Parallel()

	uniformGroup := func(id1, id2 string) domain.Group {
		return domain.Group{
			{ID: id1, NumericAttrs: map[string]float64{"gpa": 0.75}},
			{ID: id2, NumericAttrs: map[string]float64{"gpa": 0.75}},
		}
	}
	chromosome := &domain.Chromosome{
		Groups: []domain.Group{
			uniformGroup("a", "b"),
			uniformGroup("c", "d"),
			uniformGroup("e", "f"),
		},
		Stats: map[string]interface{}{},
	}
	factory := NewCriterionFactory()
	factory.Register(NewNumericCriterion("gpa_balance", "gpa", "homogeneity"))

	config := domain.Config{
		Weights:        map[string]float64{"gpa_balance": 1.0},
		EquityStrategy: "",
	}
	ctx := domain.Context{}

	stats := ComputeEquityStats(chromosome, config, factory, ctx)

	if stats.EquityEnforced {
		t.Error("expected EquityEnforced=false when config has EnforceEquity=false")
	}
	if stats.ScoreStdDev > 1e-9 {
		t.Errorf("expected stddev≈0 for uniform groups, got %.6f", stats.ScoreStdDev)
	}
	if stats.LowestTeamScore != stats.HighestTeamScore {
		t.Errorf(
			"expected lowest==highest for uniform groups (%.4f vs %.4f)",
			stats.LowestTeamScore, stats.HighestTeamScore,
		)
	}
}

// TestComputeEquityStats_NilChromosome ensures no panic on nil input.
func TestComputeEquityStats_NilChromosome(t *testing.T) {
	t.Parallel()

	factory := NewCriterionFactory()
	config := domain.Config{EquityStrategy: "minimax"}
	ctx := domain.Context{}

	stats := ComputeEquityStats(nil, config, factory, ctx)

	if stats.LowestTeamScore != 0 || stats.HighestTeamScore != 0 || stats.ScoreStdDev != 0 {
		t.Errorf("expected zero stats for nil chromosome, got %+v", stats)
	}
	if !stats.EquityEnforced {
		t.Error("EquityEnforced should reflect the config value even for nil chromosome")
	}
}

// TestComputeEquityStats_SingleGroup verifies stats for a single-team chromosome.
func TestComputeEquityStats_SingleGroup(t *testing.T) {
	t.Parallel()

	chromosome := &domain.Chromosome{
		Groups: []domain.Group{
			{
				{ID: "a", NumericAttrs: map[string]float64{"gpa": 0.9}},
				{ID: "b", NumericAttrs: map[string]float64{"gpa": 0.9}},
			},
		},
		Stats: map[string]interface{}{},
	}
	factory := NewCriterionFactory()
	factory.Register(NewNumericCriterion("gpa_balance", "gpa", "homogeneity"))

	config := domain.Config{
		Weights:        map[string]float64{"gpa_balance": 1.0},
		EquityStrategy: "minimax",
	}
	ctx := domain.Context{}

	stats := ComputeEquityStats(chromosome, config, factory, ctx)

	if stats.LowestTeamScore != stats.HighestTeamScore {
		t.Errorf(
			"single group: expected lowest==highest (%.4f vs %.4f)",
			stats.LowestTeamScore, stats.HighestTeamScore,
		)
	}
	if stats.ScoreStdDev > 1e-9 {
		t.Errorf("single group: expected stddev≈0, got %.6f", stats.ScoreStdDev)
	}
}

func TestNumericCriterion_Target(t *testing.T) {
	criterion := NewNumericCriterion("agreeableness", "agreeableness", "target")

	// Test case: Group mean close to target
	group := domain.Group{
		{ID: "1", NumericAttrs: map[string]float64{"agreeableness": 0.7}},
		{ID: "2", NumericAttrs: map[string]float64{"agreeableness": 0.8}},
		{ID: "3", NumericAttrs: map[string]float64{"agreeableness": 0.75}},
	}

	ctx := domain.Context{
		Targets: map[string]float64{"agreeableness": 0.75},
	}

	score := criterion.Calculate(&group, ctx)

	if score < 0.9 {
		t.Errorf("Expected high score when mean matches target, got %f", score)
	}
}

func TestCategoricalCriterion_BlauIndex(t *testing.T) {
	criterion := NewCategoricalCriterion("sex", "sex")

	// Balanced 2M/2F group → normalizedBlau = 1.0
	balanced := domain.Group{
		{ID: "1", CategoricalAttrs: map[string]string{"sex": "M"}},
		{ID: "2", CategoricalAttrs: map[string]string{"sex": "F"}},
		{ID: "3", CategoricalAttrs: map[string]string{"sex": "M"}},
		{ID: "4", CategoricalAttrs: map[string]string{"sex": "F"}},
	}

	// All-male group → normalizedBlau = 0.0
	homogeneous := domain.Group{
		{ID: "1", CategoricalAttrs: map[string]string{"sex": "M"}},
		{ID: "2", CategoricalAttrs: map[string]string{"sex": "M"}},
		{ID: "3", CategoricalAttrs: map[string]string{"sex": "M"}},
	}

	// 3:1 group → normalizedBlau = 0.75
	group3M1F := domain.Group{
		{ID: "1", CategoricalAttrs: map[string]string{"sex": "M"}},
		{ID: "2", CategoricalAttrs: map[string]string{"sex": "M"}},
		{ID: "3", CategoricalAttrs: map[string]string{"sex": "M"}},
		{ID: "4", CategoricalAttrs: map[string]string{"sex": "F"}},
	}

	tests := []struct {
		name      string
		group     domain.Group
		target    float64
		wantMin   float64
		wantMax   float64
		wantAbout float64 // optional exact expectation (0 means skip)
	}{
		{
			// Target=1.0, Real=1.0 → 1 - |1.0-1.0| = 1.00 (perfect)
			name:  "balanced group hits target=1.0",
			group: balanced, target: 1.0,
			wantMin: 0.99, wantMax: 1.01,
		},
		{
			// Target=0.0, Real=0.0 → 1 - |0.0-0.0| = 1.00 (homogeneity achieved)
			name:  "homogeneous group hits target=0.0",
			group: homogeneous, target: 0.0,
			wantMin: 0.99, wantMax: 1.01,
		},
		{
			// Target=0.0, Real=1.0 → 1 - |0.0-1.0| = 0.00 (homogeneity violated by full diversity)
			name:  "balanced group violates target=0.0 (homogeneity desired)",
			group: balanced, target: 0.0,
			wantMin: 0.0, wantMax: 0.01,
		},
		{
			// Target=1.0, Real=0.0 → 1 - |1.0-0.0| = 0.00 (diversity desired but group is homogeneous)
			name:  "homogeneous group violates target=1.0 (diversity desired)",
			group: homogeneous, target: 1.0,
			wantMin: 0.0, wantMax: 0.01,
		},
		{
			// Target=0.75, Real=0.75 → 1 - |0.75-0.75| = 1.00 (exact hit)
			name:  "3:1 group hits target=0.75 exactly",
			group: group3M1F, target: 0.75,
			wantMin: 0.99, wantMax: 1.01,
		},
		{
			// Target=0.9, Real=0.75 → 1 - |0.9-0.75| = 0.85 (deficit penalised)
			name:  "3:1 group below target=0.9 → deficit penalty",
			group: group3M1F, target: 0.9,
			wantMin: 0.84, wantMax: 0.86,
		},
		{
			// Target=0.5, Real=1.0 → 1 - |0.5-1.0| = 0.50 (excess penalised)
			name:  "balanced group above target=0.5 → excess penalty",
			group: balanced, target: 0.5,
			wantMin: 0.49, wantMax: 0.51,
		},
		{
			// Target=0.5, Real=0.75 → 1 - |0.5-0.75| = 0.75 (slight excess, small penalty)
			name:  "3:1 group above target=0.5 → small excess penalty",
			group: group3M1F, target: 0.5,
			wantMin: 0.74, wantMax: 0.76,
		},
		{
			// Target=0.5, Real=0.0 → 1 - |0.5-0.0| = 0.50 (deficit penalised symmetrically)
			name:  "homogeneous group below target=0.5 → deficit penalty",
			group: homogeneous, target: 0.5,
			wantMin: 0.49, wantMax: 0.51,
		},
		{
			// Target=0.0, Real=0.3 → 1 - |0.0-0.3| = 0.70 (critical case from spec)
			// Use a ~3:1 group as proxy for Real≈0.75; use exact 50/50 for Real=1.0
			// We test the formula with homogeneity target and intermediate diversity.
			name:  "spec critical case: target=0 with non-zero diversity → proportional reduction",
			group: group3M1F, target: 0.0,
			// Real=0.75 → 1 - |0.0-0.75| = 0.25
			wantMin: 0.24, wantMax: 0.26,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// sex has exactly 2 possible values → k_global=2 produces same result as k_observed=2
			ctx := domain.Context{
				Targets:        map[string]float64{"sex": tt.target},
				CategoryCounts: map[string]int{"sex": 2},
			}
			score := criterion.Calculate(&tt.group, ctx)
			if score < tt.wantMin || score > tt.wantMax {
				t.Errorf("target=%.2f → got score=%.4f, want [%.2f, %.2f]",
					tt.target, score, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// TestCategoricalCriterion_GlobalKNormalization verifies the bug fix:
// using k_global (population-level distinct categories) as the Blau denominator
// instead of k_observed (categories seen in the group alone).
//
// Bug scenario: a section has 8 distinct careers. A group of 7 students has
// 6 from the same career and 1 from a different one. The old code used
// k_observed=2 → maxBlau=0.5 → normalised ≈ 49%.
// With k_global=8  → maxBlau=0.875 → normalised ≈ 28%, which correctly
// signals low diversity against the full population vocabulary.
func TestCategoricalCriterion_GlobalKNormalization(t *testing.T) {
	criterion := NewCategoricalCriterion("career_code", "career_code")

	// Group: 6 students share career "ING", 1 student has career "MED"
	group6plus1 := domain.Group{
		{ID: "1", CategoricalAttrs: map[string]string{"career_code": "ING"}},
		{ID: "2", CategoricalAttrs: map[string]string{"career_code": "ING"}},
		{ID: "3", CategoricalAttrs: map[string]string{"career_code": "ING"}},
		{ID: "4", CategoricalAttrs: map[string]string{"career_code": "ING"}},
		{ID: "5", CategoricalAttrs: map[string]string{"career_code": "ING"}},
		{ID: "6", CategoricalAttrs: map[string]string{"career_code": "ING"}},
		{ID: "7", CategoricalAttrs: map[string]string{"career_code": "MED"}},
	}

	// Blau raw = 1 - ((6/7)² + (1/7)²) = 1 - (0.7347 + 0.0204) = 0.2449

	t.Run("without CategoryCounts (legacy k_observed=2)", func(t *testing.T) {
		// Falls back to k_observed=2 → maxBlau=0.5 → normalised=0.2449/0.5=0.490
		ctxNoGlobal := domain.Context{
			Targets: map[string]float64{"career_code": 0.2},
		}
		realVal := criterion.CalculateRealValue(&group6plus1, ctxNoGlobal)
		// Legacy result: ~49% diversity (inflated because denominator is k_observed=2)
		if realVal < 0.48 || realVal > 0.50 {
			t.Errorf("legacy k_observed=2: expected real_value≈0.49, got %.4f", realVal)
		}
	})

	t.Run("with CategoryCounts k_global=8 (correct)", func(t *testing.T) {
		// Section has 8 distinct careers → maxBlau=(8-1)/8=0.875
		// normalised = 0.2449 / 0.875 = 0.2799 ≈ 28%
		ctxGlobal8 := domain.Context{
			Targets:        map[string]float64{"career_code": 0.2},
			CategoryCounts: map[string]int{"career_code": 8},
		}
		realVal := criterion.CalculateRealValue(&group6plus1, ctxGlobal8)
		if realVal < 0.27 || realVal > 0.29 {
			t.Errorf("k_global=8: expected real_value≈0.28, got %.4f", realVal)
		}

		// Score with target=0.20 and real≈0.28: 1 - |0.20 - 0.28| = 1 - 0.08 = 0.92
		score := criterion.Calculate(&group6plus1, ctxGlobal8)
		if score < 0.91 || score > 0.93 {
			t.Errorf("k_global=8, target=0.20: expected score≈0.92 (high, group near target), got %.4f", score)
		}
	})

	t.Run("k_global=2 same as k_observed for binary attribute", func(t *testing.T) {
		// When the attribute only has 2 possible values globally, k_global=k_observed → same result
		ctxGlobal2 := domain.Context{
			Targets:        map[string]float64{"career_code": 0.2},
			CategoryCounts: map[string]int{"career_code": 2},
		}
		realVal := criterion.CalculateRealValue(&group6plus1, ctxGlobal2)
		// k_global=2 == k_observed=2 → same ~49%
		if realVal < 0.48 || realVal > 0.50 {
			t.Errorf("k_global=2 == k_observed=2: expected ~0.49, got %.4f", realVal)
		}
	})

	t.Run("all same career → real_value=0.0 regardless of k_global", func(t *testing.T) {
		allSame := domain.Group{
			{ID: "1", CategoricalAttrs: map[string]string{"career_code": "ING"}},
			{ID: "2", CategoricalAttrs: map[string]string{"career_code": "ING"}},
			{ID: "3", CategoricalAttrs: map[string]string{"career_code": "ING"}},
			{ID: "4", CategoricalAttrs: map[string]string{"career_code": "ING"}},
			{ID: "5", CategoricalAttrs: map[string]string{"career_code": "ING"}},
			{ID: "6", CategoricalAttrs: map[string]string{"career_code": "ING"}},
			{ID: "7", CategoricalAttrs: map[string]string{"career_code": "ING"}},
		}
		for _, kGlobal := range []int{0, 1, 2, 8, 20} {
			ctx := domain.Context{
				Targets:        map[string]float64{"career_code": 0.5},
				CategoryCounts: map[string]int{"career_code": kGlobal},
			}
			realVal := criterion.CalculateRealValue(&allSame, ctx)
			if realVal > 0.001 {
				t.Errorf("all-same group: expected real_value=0.0 for k_global=%d, got %.4f", kGlobal, realVal)
			}
		}
	})

	t.Run("maximum diversity group → real_value=1.0 when k_observed=k_global", func(t *testing.T) {
		// All 7 students have different careers AND k_global=7 → perfectly distributed → 100%
		allDiff := domain.Group{
			{ID: "1", CategoricalAttrs: map[string]string{"career_code": "A"}},
			{ID: "2", CategoricalAttrs: map[string]string{"career_code": "B"}},
			{ID: "3", CategoricalAttrs: map[string]string{"career_code": "C"}},
			{ID: "4", CategoricalAttrs: map[string]string{"career_code": "D"}},
			{ID: "5", CategoricalAttrs: map[string]string{"career_code": "E"}},
			{ID: "6", CategoricalAttrs: map[string]string{"career_code": "F"}},
			{ID: "7", CategoricalAttrs: map[string]string{"career_code": "G"}},
		}
		ctx := domain.Context{
			Targets:        map[string]float64{"career_code": 1.0},
			CategoryCounts: map[string]int{"career_code": 7},
		}
		realVal := criterion.CalculateRealValue(&allDiff, ctx)
		if realVal < 0.99 {
			t.Errorf("all-different group with k_global=k_observed: expected real_value≈1.0, got %.4f", realVal)
		}
	})

	t.Run("maximum diversity group is capped when k_global > k_observed", func(t *testing.T) {
		// All 7 students have different careers but k_global=20 → normalised < 1.0
		// blauRaw = (7-1)/7 = 0.857; maxBlau(k=20) = 19/20 = 0.95
		// normalised = 0.857 / 0.95 ≈ 0.902
		allDiff := domain.Group{
			{ID: "1", CategoricalAttrs: map[string]string{"career_code": "A"}},
			{ID: "2", CategoricalAttrs: map[string]string{"career_code": "B"}},
			{ID: "3", CategoricalAttrs: map[string]string{"career_code": "C"}},
			{ID: "4", CategoricalAttrs: map[string]string{"career_code": "D"}},
			{ID: "5", CategoricalAttrs: map[string]string{"career_code": "E"}},
			{ID: "6", CategoricalAttrs: map[string]string{"career_code": "F"}},
			{ID: "7", CategoricalAttrs: map[string]string{"career_code": "G"}},
		}
		ctx := domain.Context{
			Targets:        map[string]float64{"career_code": 1.0},
			CategoryCounts: map[string]int{"career_code": 20},
		}
		realVal := criterion.CalculateRealValue(&allDiff, ctx)
		// With k_global=20 the group only covers 7/20 categories → correctly < 1.0
		if realVal >= 1.0 {
			t.Errorf("7-of-20 careers: real_value should be < 1.0, got %.4f", realVal)
		}
		if realVal < 0.88 || realVal > 0.92 {
			t.Errorf("7-of-20 careers: expected real_value≈0.90, got %.4f", realVal)
		}
	})

	t.Run("k_global cannot be smaller than k_observed (safety clamp)", func(t *testing.T) {
		// If someone passes a k_global smaller than what's in the group,
		// the function must clamp to k_observed so it never produces nonsense > 1.
		ctx := domain.Context{
			Targets:        map[string]float64{"career_code": 0.5},
			CategoryCounts: map[string]int{"career_code": 1}, // absurd: only 1 global category
		}
		// group6plus1 has 2 observed categories → k is clamped to 2
		realVal := criterion.CalculateRealValue(&group6plus1, ctx)
		if realVal > 1.0 {
			t.Errorf("safety clamp: real_value must never exceed 1.0, got %.4f", realVal)
		}
		if realVal < 0.0 {
			t.Errorf("safety clamp: real_value must never be negative, got %.4f", realVal)
		}
	})
}

// TestBlauNormalized_CoreFormula tests the internal blauNormalized helper directly
// to verify the mathematical correctness of the denominator selection.
func TestBlauNormalized_CoreFormula(t *testing.T) {
	tests := []struct {
		name    string
		counts  map[string]int
		kGlobal int
		wantMin float64
		wantMax float64
	}{
		// ── Boundary: all same ──────────────────────────────────────────────
		{"all same, k=1", map[string]int{"A": 7}, 1, 0.0, 0.001},
		{"all same, k=8", map[string]int{"A": 7}, 8, 0.0, 0.001},

		// ── Bug case: 6+1 ───────────────────────────────────────────────────
		// Blau_raw = 0.2449; maxBlau(k=2)=0.5 → 0.490; maxBlau(k=8)=0.875 → 0.280
		{"6+1, k_observed=2 (no global)", map[string]int{"A": 6, "B": 1}, 0, 0.48, 0.50},
		{"6+1, k_global=8", map[string]int{"A": 6, "B": 1}, 8, 0.27, 0.29},
		{"6+1, k_global=2 (same as obs)", map[string]int{"A": 6, "B": 1}, 2, 0.48, 0.50},

		// ── Near-equal split ────────────────────────────────────────────────
		// 4+3, Blau_raw=0.490; maxBlau(k=2)=0.5 → 0.980; maxBlau(k=8)=0.875 → 0.560
		{"4+3 binary (k=2)", map[string]int{"A": 4, "B": 3}, 2, 0.97, 0.99},
		{"4+3, k_global=8", map[string]int{"A": 4, "B": 3}, 8, 0.55, 0.57},

		// ── Perfect spread within observed k ────────────────────────────────
		{"7 distinct, k_global=7", map[string]int{"A": 1, "B": 1, "C": 1, "D": 1, "E": 1, "F": 1, "G": 1}, 7, 0.99, 1.01},

		// ── Safety: k_global < k_observed is clamped ────────────────────────
		{"clamp: k_global=1 < k_obs=2", map[string]int{"A": 6, "B": 1}, 1, 0.48, 0.50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := blauNormalized(tt.counts, tt.kGlobal)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("blauNormalized(%v, k=%d) = %.4f, want [%.2f, %.2f]",
					tt.counts, tt.kGlobal, got, tt.wantMin, tt.wantMax)
			}
			if got < 0.0 || got > 1.0 {
				t.Errorf("blauNormalized result out of [0,1]: %.4f", got)
			}
		})
	}
}

func TestSocialCriterion_Calculate(t *testing.T) {
	criterion := NewSocialCriterion("social", 1.0)

	// All-positive graph: 3 members, all directed edges are +ve.
	// 6 directed edges (3*2), all positive.
	// positiveSum = 0.8+0.7+0.6+0.9+0.5+0.85 = 4.35, negatives=0, total=6
	// average Tp = 4.35/6 = 0.725
	// Relational cost C'_ij = (1 - Sp*Tp)/2 = (1 - 0.725)/2 = 0.1375
	// Score = 1 - 0.1375 = 0.8625
	socialGraph := domain.SocialGraph{
		"1": {"2": 0.8, "3": 0.6},
		"2": {"1": 0.7, "3": 0.9},
		"3": {"1": 0.5, "2": 0.85},
	}

	group := domain.Group{
		{ID: "1", NumericAttrs: map[string]float64{}},
		{ID: "2", NumericAttrs: map[string]float64{}},
		{ID: "3", NumericAttrs: map[string]float64{}},
	}

	ctx := domain.Context{
		SocialGraph: socialGraph,
	}

	score := criterion.Calculate(&group, ctx)

	// All positives: score should be high (> 0.8)
	if score < 0.8 || score > 1.0 {
		t.Errorf("Expected high social cohesion score for all-positive graph, got %f", score)
	}

	if math.Abs(score-0.8625) > 1e-9 {
		t.Errorf("all-positive graph: expected score=0.8625, got %.9f", score)
	}

	// No connections in graph → neutral (0.5)
	group2 := domain.Group{
		{ID: "10", NumericAttrs: map[string]float64{}},
		{ID: "20", NumericAttrs: map[string]float64{}},
	}

	score2 := criterion.Calculate(&group2, ctx)

	if score2 != 0.5 {
		t.Errorf("Expected neutral score 0.5 for no connections, got %f", score2)
	}
}

// TestSocialCriterion_VetoPenalty_SingleRejection verifies that a single
// rejection edge in one direction lowers the score compared to the all-positive
// baseline, even when the other direction is positive.
//
// Graph: A→B=+1, B→A=-1 (B vetoes A).
// Tp = (1.0 + (-1.0)) / 2 = 0.0 (conflict)
// C'_ij = (1 - 1*0) / 2 = 0.5 → score = 1.0 - 0.5 = 0.5
func TestSocialCriterion_VetoPenalty_SingleRejection(t *testing.T) {
	t.Parallel()
	criterion := NewSocialCriterion("social", 1.0)

	graph := domain.SocialGraph{
		"A": {"B": 1.0},
		"B": {"A": -1.0},
	}
	group := domain.Group{
		{ID: "A", NumericAttrs: map[string]float64{}},
		{ID: "B", NumericAttrs: map[string]float64{}},
	}
	ctx := domain.Context{SocialGraph: graph}

	score := criterion.Calculate(&group, ctx)

	if math.Abs(score-0.5) > 1e-9 {
		t.Errorf("A→B=+1, B→A=-1: expected score=0.5 (conflict), got %.9f", score)
	}
}

// TestSocialCriterion_VetoPenalty_Asymmetry validates the core sociometry
// requirement: "B rejects A" is toxic even if A chose B positively.
// A group where only ONE direction is negative must score lower than a
// group where both are positive.
func TestSocialCriterion_VetoPenalty_Asymmetry(t *testing.T) {
	t.Parallel()
	criterion := NewSocialCriterion("social", 1.0)

	// Baseline: both choose each other positively.
	graphPositive := domain.SocialGraph{
		"A": {"B": 1.0},
		"B": {"A": 1.0},
	}
	// Toxic: A chooses B, but B rejects A.
	graphVeto := domain.SocialGraph{
		"A": {"B": 1.0},
		"B": {"A": -1.0},
	}
	group := domain.Group{
		{ID: "A", NumericAttrs: map[string]float64{}},
		{ID: "B", NumericAttrs: map[string]float64{}},
	}

	scorePositive := criterion.Calculate(&group, domain.Context{SocialGraph: graphPositive})
	scoreVeto := criterion.Calculate(&group, domain.Context{SocialGraph: graphVeto})

	if scoreVeto >= scorePositive {
		t.Errorf(
			"toxic relationship must score lower than mutual attraction: veto=%.4f >= positive=%.4f",
			scoreVeto, scorePositive,
		)
	}
}

// TestSocialCriterion_VetoPenalty_MultipleRejections verifies that adding more
// rejection edges progressively lowers the score.
func TestSocialCriterion_VetoPenalty_MultipleRejections(t *testing.T) {
	t.Parallel()
	criterion := NewSocialCriterion("social", 1.0)

	// Group of 3; 3 pairs: (A,B), (A,C), (B,C).
	// All-positive baseline: all 6 edges = 1.0 → score=1.0
	graphAllPos := domain.SocialGraph{
		"A": {"B": 1.0, "C": 1.0},
		"B": {"A": 1.0, "C": 1.0},
		"C": {"A": 1.0, "B": 1.0},
	}
	// One rejection: C→A=-1, others +1 → (A,C) Tp = (1 - 1)/2 = 0.0, (A,B) Tp = 1.0, (B,C) Tp = 1.0
	// Cost = (0.5 + 0.0 + 0.0)/3 = 0.5/3 ≈ 0.1667 → score ≈ 0.8333
	graphOneVeto := domain.SocialGraph{
		"A": {"B": 1.0, "C": 1.0},
		"B": {"A": 1.0, "C": 1.0},
		"C": {"A": -1.0, "B": 1.0},
	}
	// Three rejections (A totally rejected): B→A=-1, C→A=-1, A→B=-1
	// More rejections → lower score
	graphThreeVeto := domain.SocialGraph{
		"A": {"B": -1.0, "C": 1.0},
		"B": {"A": -1.0, "C": 1.0},
		"C": {"A": -1.0, "B": 1.0},
	}

	group := domain.Group{
		{ID: "A", NumericAttrs: map[string]float64{}},
		{ID: "B", NumericAttrs: map[string]float64{}},
		{ID: "C", NumericAttrs: map[string]float64{}},
	}

	scoreAllPos := criterion.Calculate(&group, domain.Context{SocialGraph: graphAllPos})
	scoreOneVeto := criterion.Calculate(&group, domain.Context{SocialGraph: graphOneVeto})
	scoreThreeVeto := criterion.Calculate(&group, domain.Context{SocialGraph: graphThreeVeto})

	// Strict ordering: all-positive > one-veto > three-vetoes
	if scoreAllPos <= scoreOneVeto {
		t.Errorf("allPositive (%.4f) should be > oneVeto (%.4f)", scoreAllPos, scoreOneVeto)
	}
	if scoreOneVeto <= scoreThreeVeto {
		t.Errorf("oneVeto (%.4f) should be > threeVetoes (%.4f)", scoreOneVeto, scoreThreeVeto)
	}

	// Exact assertions
	if math.Abs(scoreAllPos-1.0) > 1e-9 {
		t.Errorf("all-positive: expected 1.0, got %.9f", scoreAllPos)
	}
	wantOneVeto := 1.0 - (0.5/3.0)
	if math.Abs(scoreOneVeto-wantOneVeto) > 1e-9 {
		t.Errorf("one-veto: expected %.4f, got %.9f", wantOneVeto, scoreOneVeto)
	}
}

// TestSocialCriterion_MaximizationBehavior verifies that a group of perfect
// mutual attractions receives the maximum possible score (1.0).
func TestSocialCriterion_MaximizationBehavior(t *testing.T) {
	t.Parallel()
	criterion := NewSocialCriterion("social", 1.0)

	graph := domain.SocialGraph{
		"A": {"B": 1.0, "C": 1.0, "D": 1.0},
		"B": {"A": 1.0, "C": 1.0, "D": 1.0},
		"C": {"A": 1.0, "B": 1.0, "D": 1.0},
		"D": {"A": 1.0, "B": 1.0, "C": 1.0},
	}
	group := domain.Group{
		{ID: "A", NumericAttrs: map[string]float64{}},
		{ID: "B", NumericAttrs: map[string]float64{}},
		{ID: "C", NumericAttrs: map[string]float64{}},
		{ID: "D", NumericAttrs: map[string]float64{}},
	}
	ctx := domain.Context{SocialGraph: graph}

	score := criterion.Calculate(&group, ctx)

	if math.Abs(score-1.0) > 1e-9 {
		t.Errorf("perfect mutual attraction group: expected score=1.0, got %.9f", score)
	}
}

func TestSocialCriterion_StrategyAndTargetSemantics(t *testing.T) {
	t.Parallel()

	// Mutual positive graph
	graphPos := domain.SocialGraph{
		"A": {"B": 1.0},
		"B": {"A": 1.0},
	}
	// Mutual negative graph
	graphNeg := domain.SocialGraph{
		"A": {"B": -1.0},
		"B": {"A": -1.0},
	}
	group := domain.Group{
		{ID: "A", NumericAttrs: map[string]float64{}},
		{ID: "B", NumericAttrs: map[string]float64{}},
	}

	harmony := NewSocialCriterion("social", 1.0)    // Sp = 1 (harmony)
	friction := NewSocialCriterion("social", -1.0)  // Sp = -1 (controlled friction)

	hPos := harmony.Calculate(&group, domain.Context{SocialGraph: graphPos})
	hNeg := harmony.Calculate(&group, domain.Context{SocialGraph: graphNeg})
	fPos := friction.Calculate(&group, domain.Context{SocialGraph: graphPos})
	fNeg := friction.Calculate(&group, domain.Context{SocialGraph: graphNeg})

	// Harmony: positive relationship scores 1.0, negative scores 0.0
	if math.Abs(hPos-1.0) > 1e-9 {
		t.Errorf("harmony with positive relationship: expected 1.0, got %.9f", hPos)
	}
	if math.Abs(hNeg-0.0) > 1e-9 {
		t.Errorf("harmony with negative relationship: expected 0.0, got %.9f", hNeg)
	}

	// Controlled friction: negative relationship scores 1.0, positive scores 0.0
	if math.Abs(fNeg-1.0) > 1e-9 {
		t.Errorf("friction with negative relationship: expected 1.0, got %.9f", fNeg)
	}
	if math.Abs(fPos-0.0) > 1e-9 {
		t.Errorf("friction with positive relationship: expected 0.0, got %.9f", fPos)
	}
}

func TestCriterionFactory_RegisterAndGet(t *testing.T) {
	factory := NewCriterionFactory()

	// Register custom criterion
	criterion := NewNumericCriterion("test_criterion", "test_attr", "homogeneity")
	factory.Register(criterion)

	// Retrieve criterion
	retrieved := factory.Get("test_criterion")
	if retrieved == nil {
		t.Fatal("Expected to retrieve registered criterion")
	}

	if retrieved.Name() != "test_criterion" {
		t.Errorf("Expected name 'test_criterion', got '%s'", retrieved.Name())
	}

	// Try to get non-existent criterion
	notFound := factory.Get("non_existent")
	if notFound != nil {
		t.Error("Expected nil for non-existent criterion")
	}
}

func TestCriterionFactory_RegisterDefaults(t *testing.T) {
	factory := NewCriterionFactory()
	factory.RegisterDefaults()

	// Check that default criteria are registered
	expectedCriteria := []string{
		"age_balance",
		"gpa_balance",
		"sex",
		"career_code",
		"social",
		"openness",
		"conscientiousness",
	}

	for _, name := range expectedCriteria {
		criterion := factory.Get(name)
		if criterion == nil {
			t.Errorf("Expected default criterion '%s' to be registered", name)
		}
	}
}

func TestCalculateFitness_BasicScenario(t *testing.T) {
	// Setup factory with criteria
	factory := NewCriterionFactory()
	factory.Register(NewCategoricalCriterion("sex", "sex"))
	factory.Register(NewNumericCriterion("age_balance", "age", "homogeneity"))

	// Setup config with weights
	config := domain.Config{
		Weights: map[string]float64{
			"sex":         0.4,
			"age_balance": 0.6,
		},
		UseQuadraticPenalty: false,
	}

	// Create chromosome with one group: balanced 2M/2F, similar ages
	// sex: target=1.0, real=1.0 → score_sex = 1 - |1.0-1.0| = 1.0
	// age: homogeneity, low variance → score_age ≈ high
	group := domain.Group{
		{ID: "1", NumericAttrs: map[string]float64{"age": 20}, CategoricalAttrs: map[string]string{"sex": "M"}},
		{ID: "2", NumericAttrs: map[string]float64{"age": 21}, CategoricalAttrs: map[string]string{"sex": "F"}},
		{ID: "3", NumericAttrs: map[string]float64{"age": 20}, CategoricalAttrs: map[string]string{"sex": "M"}},
		{ID: "4", NumericAttrs: map[string]float64{"age": 21}, CategoricalAttrs: map[string]string{"sex": "F"}},
	}

	chromosome := &domain.Chromosome{
		Groups: []domain.Group{group},
	}

	ctx := domain.Context{
		Targets: map[string]float64{"sex": 1.0}, // balanced 50/50 → hits target exactly
	}

	fitness := CalculateFitness(chromosome, config, factory, ctx)

	if fitness < 0.0 || fitness > 1.0 {
		t.Errorf("Fitness score should be in [0, 1], got %f", fitness)
	}

	// sex: target=1.0, real=1.0 → score = 1 - |1.0-1.0| = 1.0 (weight 0.4)
	// age_balance: ages 20,21,20,21 → variance=0.25 → score=1/(1+0.25*25)≈0.138 (weight 0.6)
	// calculateGroupScore normalises by activeWeight=1.0:
	//   (1.0*0.4 + 0.138*0.6) / 1.0 ≈ 0.483
	if fitness < 0.40 || fitness > 0.60 {
		t.Errorf("Expected fitness in [0.40, 0.60] for this group (sex hits target, age has moderate variance), got %f", fitness)
	}
}

func TestCalculateFitness_WeightNormalization(t *testing.T) {
	factory := NewCriterionFactory()
	factory.Register(NewCategoricalCriterion("sex", "sex"))
	factory.Register(NewNumericCriterion("age_balance", "age", "homogeneity"))

	// Weights that don't sum to 1.0
	config := domain.Config{
		Weights: map[string]float64{
			"sex":         2.0,
			"age_balance": 3.0,
		},
		UseQuadraticPenalty: false,
	}

	group := domain.Group{
		{ID: "1", NumericAttrs: map[string]float64{"age": 20}, CategoricalAttrs: map[string]string{"sex": "M"}},
		{ID: "2", NumericAttrs: map[string]float64{"age": 20}, CategoricalAttrs: map[string]string{"sex": "F"}},
	}

	chromosome := &domain.Chromosome{Groups: []domain.Group{group}}
	ctx := domain.Context{}

	fitness := CalculateFitness(chromosome, config, factory, ctx)

	// Should still produce valid fitness despite non-normalized weights
	if fitness < 0.0 || fitness > 1.0 {
		t.Errorf("Fitness score should be in [0, 1] even with non-normalized weights, got %f", fitness)
	}
}

func TestCalculateFitness_MissingCriterion(t *testing.T) {
	factory := NewCriterionFactory()
	// Only register one criterion
	factory.Register(NewCategoricalCriterion("sex", "sex"))

	// Config requests two criteria, but only one exists
	config := domain.Config{
		Weights: map[string]float64{
			"sex":               0.5,
			"non_existent_crit": 0.5,
		},
	}

	group := domain.Group{
		{ID: "1", CategoricalAttrs: map[string]string{"sex": "M"}},
		{ID: "2", CategoricalAttrs: map[string]string{"sex": "F"}},
	}

	chromosome := &domain.Chromosome{Groups: []domain.Group{group}}
	ctx := domain.Context{}

	fitness := CalculateFitness(chromosome, config, factory, ctx)

	// Should still calculate fitness using only the available criterion
	if fitness < 0.0 || fitness > 1.0 {
		t.Errorf("Fitness should handle missing criteria gracefully, got %f", fitness)
	}
}

func TestCalculateFitness_QuadraticPenalty(t *testing.T) {
	factory := NewCriterionFactory()
	factory.Register(NewCategoricalCriterion("sex", "sex"))

	// Create two configs: with and without quadratic penalty
	configLinear := domain.Config{
		Weights:             map[string]float64{"sex": 1.0},
		UseQuadraticPenalty: false,
	}

	configQuadratic := domain.Config{
		Weights:             map[string]float64{"sex": 1.0},
		UseQuadraticPenalty: true,
	}

	// Group with moderate diversity
	group := domain.Group{
		{ID: "1", CategoricalAttrs: map[string]string{"sex": "M"}},
		{ID: "2", CategoricalAttrs: map[string]string{"sex": "M"}},
		{ID: "3", CategoricalAttrs: map[string]string{"sex": "F"}},
	}

	chromosome := &domain.Chromosome{Groups: []domain.Group{group}}
	ctx := domain.Context{}

	fitnessLinear := CalculateFitness(chromosome, configLinear, factory, ctx)
	fitnessQuadratic := CalculateFitness(chromosome, configQuadratic, factory, ctx)

	// Quadratic penalty should produce different (typically higher) scores
	if fitnessLinear == fitnessQuadratic {
		t.Logf("Warning: Quadratic penalty didn't change score (linear=%f, quadratic=%f)", fitnessLinear, fitnessQuadratic)
	}
}

func TestCalculateFitness_MultipleGroups(t *testing.T) {
	factory := NewCriterionFactory()
	factory.Register(NewCategoricalCriterion("sex", "sex"))

	config := domain.Config{
		Weights: map[string]float64{"sex": 1.0},
	}

	// Both groups are perfectly balanced 1M/1F → normalizedBlau=1.0 each
	// With target=1.0: score = 1 - |1.0-1.0| = 1.0 per group → overall fitness = 1.0
	group1 := domain.Group{
		{ID: "1", CategoricalAttrs: map[string]string{"sex": "M"}},
		{ID: "2", CategoricalAttrs: map[string]string{"sex": "F"}},
	}

	group2 := domain.Group{
		{ID: "3", CategoricalAttrs: map[string]string{"sex": "F"}},
		{ID: "4", CategoricalAttrs: map[string]string{"sex": "M"}},
	}

	chromosome := &domain.Chromosome{Groups: []domain.Group{group1, group2}}
	ctx := domain.Context{
		Targets: map[string]float64{"sex": 1.0},
	}

	fitness := CalculateFitness(chromosome, config, factory, ctx)

	if fitness < 0.0 || fitness > 1.0 {
		t.Errorf("Fitness should be in [0, 1] for multiple groups, got %f", fitness)
	}

	// Both groups hit target=1.0 exactly → fitness should be 1.0
	if fitness < 0.99 {
		t.Errorf("Expected fitness ~1.0 for perfectly balanced groups hitting target, got %f", fitness)
	}
}

func TestCalculateFitnessForGroup(t *testing.T) {
	factory := NewCriterionFactory()
	factory.Register(NewCategoricalCriterion("sex", "sex"))

	config := domain.Config{
		Weights: map[string]float64{"sex": 1.0},
	}

	// 1M/1F with no explicit target → default target=0.5
	// normalizedBlau=1.0, target=0.5 → score = 1 - |0.5-1.0| = 0.5
	group := domain.Group{
		{ID: "1", CategoricalAttrs: map[string]string{"sex": "M"}},
		{ID: "2", CategoricalAttrs: map[string]string{"sex": "F"}},
	}

	ctx := domain.Context{}

	fitness := CalculateFitnessForGroup(&group, config, factory, ctx)

	if fitness < 0.0 || fitness > 1.0 {
		t.Errorf("Group fitness should be in [0, 1], got %f", fitness)
	}

	// With default target=0.5 and real=1.0: score = 1 - |0.5-1.0| = 0.5
	if fitness < 0.49 || fitness > 0.51 {
		t.Errorf("Expected fitness ~0.5 (default target=0.5 with balanced group), got %f", fitness)
	}
}

func TestCategoricalCriterion_AbsoluteError_SymmetricPenalty(t *testing.T) {
	criterion := NewCategoricalCriterion("sex", "sex")

	// Verify symmetry: excess and deficit at equal distance from target yield the same score.
	// Target=0.5:
	//   Real=0.75 (excess +0.25)  → 1 - |0.5-0.75| = 0.75
	//   Real=0.25 (deficit -0.25) → 1 - |0.5-0.25| = 0.75
	// We can't directly produce Real=0.25 with binary sex; use a 3-category attribute instead.

	criterion3 := NewCategoricalCriterion("lang", "lang")

	// 3 equal categories → normalizedBlau=1.0
	balanced3 := domain.Group{
		{ID: "1", CategoricalAttrs: map[string]string{"lang": "A"}},
		{ID: "2", CategoricalAttrs: map[string]string{"lang": "B"}},
		{ID: "3", CategoricalAttrs: map[string]string{"lang": "C"}},
	}
	// All same → normalizedBlau=0.0
	homogeneous3 := domain.Group{
		{ID: "1", CategoricalAttrs: map[string]string{"lang": "A"}},
		{ID: "2", CategoricalAttrs: map[string]string{"lang": "A"}},
		{ID: "3", CategoricalAttrs: map[string]string{"lang": "A"}},
	}

	ctx05 := domain.Context{Targets: map[string]float64{"lang": 0.5}}

	// Real=1.0, target=0.5 → excess of 0.5 → score = 1 - 0.5 = 0.5
	scoreExcess := criterion3.Calculate(&balanced3, ctx05)
	// Real=0.0, target=0.5 → deficit of 0.5 → score = 1 - 0.5 = 0.5
	scoreDeficit := criterion3.Calculate(&homogeneous3, ctx05)

	if scoreExcess < 0.49 || scoreExcess > 0.51 {
		t.Errorf("Expected score=0.5 for excess (real=1.0, target=0.5), got %f", scoreExcess)
	}
	if scoreDeficit < 0.49 || scoreDeficit > 0.51 {
		t.Errorf("Expected score=0.5 for deficit (real=0.0, target=0.5), got %f", scoreDeficit)
	}
	// Symmetry check
	if abs := scoreExcess - scoreDeficit; abs > 0.01 || abs < -0.01 {
		t.Errorf("Penalty should be symmetric: excess=%f deficit=%f", scoreExcess, scoreDeficit)
	}

	// Spec critical case: Target=0, Real=0.3 → score = 1 - |0-0.3| = 0.7
	// Approximate Real≈0.75 using a 2-sex 3:1 group, but for exact 0.3 we use
	// a 3-category group with 2 of one kind and 1 each of the other two:
	// counts: A=1, B=1, C=1 → Blau=1-3*(1/3)²=1-1/3=0.667; maxBlau=(3-1)/3=0.667 → normalizedBlau=1.0
	// To get a lower value we use 2 of A, 1 of B, 0 of C:
	groupLow := domain.Group{
		{ID: "1", CategoricalAttrs: map[string]string{"sex": "M"}},
		{ID: "2", CategoricalAttrs: map[string]string{"sex": "M"}},
		{ID: "3", CategoricalAttrs: map[string]string{"sex": "F"}},
	}
	// 2M/1F binary → Blau = 1-(4/9+1/9) = 1-5/9 = 4/9; maxBlau=0.5 → normalized=8/9≈0.889
	// target=0 → score = 1 - |0-0.889| = 0.111
	ctxZero := domain.Context{Targets: map[string]float64{"sex": 0.0}}
	scoreZero := criterion.Calculate(&groupLow, ctxZero)
	// For a binary group with non-zero diversity, score should be < 1.0 when target=0
	if scoreZero >= 1.0 {
		t.Errorf("Expected penalty when target=0 and group has diversity, got score=%f", scoreZero)
	}
	if scoreZero < 0.0 {
		t.Errorf("Score must be >= 0, got %f", scoreZero)
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test calculateMean
	values := []float64{10.0, 20.0, 30.0}
	mean := calculateMean(values)
	if mean != 20.0 {
		t.Errorf("Expected mean 20.0, got %f", mean)
	}

	// Test calculateVariance
	variance := calculateVariance(values)
	expectedVariance := 66.666666 // approximately
	if variance < expectedVariance-1 || variance > expectedVariance+1 {
		t.Errorf("Expected variance ~66.67, got %f", variance)
	}

	// Test with empty slice
	meanEmpty := calculateMean([]float64{})
	if meanEmpty != 0.0 {
		t.Errorf("Expected mean 0.0 for empty slice, got %f", meanEmpty)
	}

	varianceEmpty := calculateVariance([]float64{})
	if varianceEmpty != 0.0 {
		t.Errorf("Expected variance 0.0 for empty slice, got %f", varianceEmpty)
	}
}

func BenchmarkCalculateFitness(b *testing.B) {
	factory := NewCriterionFactory()
	factory.RegisterDefaults()

	config := domain.Config{
		Weights: map[string]float64{
			"sex":         0.3,
			"age_balance": 0.3,
			"social":      0.4,
		},
	}

	// Create realistic chromosome
	groups := make([]domain.Group, 10)
	for i := 0; i < 10; i++ {
		group := make(domain.Group, 5)
		for j := 0; j < 5; j++ {
			id := string(rune('A' + i*5 + j))
			group[j] = &domain.Student{
				ID:               id,
				NumericAttrs:     map[string]float64{"age": 20.0 + float64(j)},
				CategoricalAttrs: map[string]string{"sex": []string{"M", "F"}[j%2]},
			}
		}
		groups[i] = group
	}

	chromosome := &domain.Chromosome{Groups: groups}
	ctx := domain.Context{
		SocialGraph: make(domain.SocialGraph),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateFitness(chromosome, config, factory, ctx)
	}
}
