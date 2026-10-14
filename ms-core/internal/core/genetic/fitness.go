package genetic

import (
	"fmt"
	"math"

	"ms-core/internal/core/domain"
)

// EquityStats holds per-solution equity metrics computed after fitness evaluation.
// These values are attached to the metadata of the final response so callers can
// audit whether equity enforcement produced a more homogeneous distribution.
type EquityStats struct {
	LowestTeamScore  float64
	HighestTeamScore float64
	ScoreStdDev      float64
	ScoreVariance    float64
	EquityEnforced   bool
}

// ============================================================================
// CRITERION IMPLEMENTATIONS
// ============================================================================

// NumericCriterion evaluates diversity or balance for numeric attributes.
// It supports multiple strategies: variance, homogeneity, heterogeneity, target-based.
type NumericCriterion struct {
	name     string
	attrKey  string
	strategy string  // "variance", "homogeneity", "heterogeneity", "target", "minimize", "maximize"
	gamma    float64 // sensitivity factor for diversity index formula
}

// NewNumericCriterion creates a NumericCriterion with default gamma=25.0
func NewNumericCriterion(name, attrKey, strategy string) *NumericCriterion {
	return &NumericCriterion{
		name:     name,
		attrKey:  attrKey,
		strategy: strategy,
		gamma:    25.0,
	}
}

// NewNumericCriterionWithGamma creates a NumericCriterion with custom gamma
func NewNumericCriterionWithGamma(name, attrKey, strategy string, gamma float64) *NumericCriterion {
	return &NumericCriterion{
		name:     name,
		attrKey:  attrKey,
		strategy: strategy,
		gamma:    gamma,
	}
}

func (c *NumericCriterion) Name() string {
	return c.name
}

// CalculateDiversityIndex computes the normalized diversity index D_k according to thesis formula:
// D_k(G_j) = 1 - (1 / (1 + γ * σ²_k(G_j)))
// This returns a value in [0, 1] where higher values indicate more diversity.
func (c *NumericCriterion) CalculateDiversityIndex(values []float64) float64 {
	variance := calculateVariance(values)
	return 1.0 - (1.0 / (1.0 + c.gamma*variance))
}

// CalculateGap computes the gap φ_k(G_j) = |T_k - D_k(G_j)|
// T_k is the diversity target (0 for homogeneity, 1 for heterogeneity)
// Returns the absolute error between target and diversity index.
func (c *NumericCriterion) CalculateGap(values []float64, targetDiversity int) float64 {
	diversityIndex := c.CalculateDiversityIndex(values)
	target := float64(targetDiversity) // 0 or 1
	return math.Abs(target - diversityIndex)
}

func (c *NumericCriterion) Calculate(g *domain.Group, ctx domain.Context) float64 {
	if g == nil || len(*g) == 0 {
		return 0.0
	}

	values := make([]float64, 0, len(*g))
	for _, student := range *g {
		if val, exists := student.NumericAttrs[c.attrKey]; exists {
			values = append(values, val)
		}
	}

	if len(values) == 0 {
		return 0.5 // Neutral score if attribute missing
	}

	switch c.strategy {
	case "homogeneity":
		return c.homogeneityScore(values, ctx)
	case "heterogeneity":
		return c.heterogeneityScore(values, ctx)
	case "target":
		return c.targetScore(values, ctx)
	case "variance":
		return c.varianceScore(values)
	case "minimize":
		return c.minimizeScore(values)
	case "maximize":
		return c.maximizeScore(values)
	case "diversity":
		return c.diversityScore(values, ctx)
	default:
		return c.targetScore(values, ctx)
	}
}

// CalculateRealValue returns the raw underlying metric for the group, independent
// of the fitness transformation. Used by EvaluateGroups to populate RealValue in
// CriterionBreakdownEntry so the frontend can display the actual observed value.
//
//   - homogeneity / heterogeneity / variance → normalised variance (0–1)
//   - target / minimize / maximize / default → group mean (0–1)
func (c *NumericCriterion) CalculateRealValue(g *domain.Group, ctx domain.Context) float64 {
	if g == nil || len(*g) == 0 {
		return 0.0
	}
	values := make([]float64, 0, len(*g))
	for _, student := range *g {
		if val, exists := student.NumericAttrs[c.attrKey]; exists {
			values = append(values, val)
		}
	}
	if len(values) == 0 {
		return 0.5
	}
	switch c.strategy {
	case "homogeneity", "heterogeneity", "variance":
		return calculateVariance(values)
	default:
		return calculateMean(values)
	}
}

// According to thesis formula:
// φₖ(Gⱼ) = |Tₖ - Dₖ(Gⱼ)|
// where Tₖ ∈ [0, 1] is the target diversity level
// and Dₖ is the diversity index (0-1)
// Lower φ is better (it's a cost), but the criterion interface returns a score (higher = better)
// So we return (1 - φ) which gives higher score for lower cost

// getTargetDiversity returns Tₖ from context, defaulting to 0 for homogeneity or 1 for heterogeneity
func (c *NumericCriterion) getTargetDiversity(ctx domain.Context) float64 {
	if target, exists := ctx.Targets[c.name]; exists {
		return target
	}
	if target, exists := ctx.Targets[c.attrKey]; exists {
		return target
	}
	if target, exists := ctx.Targets[c.attrKey+"_variance"]; exists {
		return target
	}
	if target, exists := ctx.Targets[c.attrKey+"_target"]; exists {
		return target
	}
	// Fallback based on strategy
	if c.strategy == "homogeneity" {
		return 0.0
	}
	return 1.0 // heterogeneity, maximize, etc.
}

func (c *NumericCriterion) homogeneityScore(values []float64, ctx domain.Context) float64 {
	// Tₖ = 0 for pure homogeneity, but we allow Tₖ from target for continuous control
	diversityIndex := c.CalculateDiversityIndex(values)
	Tk := c.getTargetDiversity(ctx)
	phi := math.Abs(Tk - diversityIndex) // φ = |Tₖ - Dₖ|
	return 1.0 - phi                     // Convert cost to score (higher = better)
}

func (c *NumericCriterion) heterogeneityScore(values []float64, ctx domain.Context) float64 {
	// Tₖ = 1 for pure heterogeneity, but we allow Tₖ from target for continuous control
	diversityIndex := c.CalculateDiversityIndex(values)
	Tk := c.getTargetDiversity(ctx)
	phi := math.Abs(Tk - diversityIndex) // φ = |Tₖ - Dₖ|
	return 1.0 - phi                     // Convert cost to score (higher = better)
}

// targetScore is for the "target" strategy where we want mean close to a target value
// This is NOT part of the thesis f1 formula - it's an additional strategy
func (c *NumericCriterion) targetScore(values []float64, ctx domain.Context) float64 {
	mean := calculateMean(values)
	targetVal := 0.5
	if target, exists := ctx.Targets[c.attrKey]; exists {
		targetVal = target
	}
	// Distance to target (normalized)
	distance := math.Abs(mean - targetVal)
	return math.Max(0.0, 1.0-distance)
}

func (c *NumericCriterion) varianceScore(values []float64) float64 {
	variance := calculateVariance(values)
	// Normalize variance to [0, 1] range (assuming variance typically < 1)
	return math.Max(0.0, 1.0-variance)
}

// minimizeScore rewards groups whose mean is as low as possible.
// Score = 1 - mean, so a mean of 0.0 → score 1.0 and mean of 1.0 → score 0.0.
// Useful for traits where lower is better (e.g. neuroticism for emotional stability).
func (c *NumericCriterion) minimizeScore(values []float64) float64 {
	mean := calculateMean(values)
	return math.Max(0.0, 1.0-mean)
}

// maximizeScore rewards groups whose mean is as high as possible.
// Score = mean, so a mean of 1.0 → score 1.0 and mean of 0.0 → score 0.0.
// Useful for traits where higher is better (e.g. agreeableness, conscientiousness).
func (c *NumericCriterion) maximizeScore(values []float64) float64 {
	return calculateMean(values)
}

// diversityScore implements the thesis formula for numeric attributes:
// D_k(G_j) = 1 - (1 / (1 + γ * σ²))
// φ_k(G_j) = |T_k - D_k|
// Where T_k is read from ctx.Targets[c.attrKey + "_target"] (0 for homogeneity, 1 for heterogeneity)
func (c *NumericCriterion) diversityScore(values []float64, ctx domain.Context) float64 {
	targetDiversity := 0 // default: homogeneity
	if t, exists := ctx.Targets[c.attrKey+"_target"]; exists {
		targetDiversity = int(t + 0.5) // convert 0/1 float to int
	}
	return c.CalculateGap(values, targetDiversity)
}

// CategoricalCriterion evaluates diversity using Blau's Index of heterogeneity.
// Blau Index = 1 - Σ(p_i²), where p_i is the proportion of category i.
//
// Scoring uses Absolute Error: Score = 1 - |Target - NormalizedBlau|
//
// The target parameter is a continuous value in [0.0, 1.0] that represents the
// desired diversity level. Both excess and deficit are penalised symmetrically:
//
//   - target = 0.0: Homogeneity desired. Any diversity > 0 reduces the score.
//     Special case: if Target=0 and Real=0.3, Score = 1 - |0-0.3| = 0.7.
//   - target = 0.5: Moderate diversity desired. Scores peak at exactly 50% diversity.
//   - target = 1.0: Maximum diversity desired. Scores peak at 100% diversity.
//
// Examples:
//   - Target=0.8, Real=0.8  → Score = 1 - |0.8-0.8| = 1.00 (perfect)
//   - Target=0.8, Real=1.0  → Score = 1 - |0.8-1.0| = 0.80 (excess penalised)
//   - Target=0.8, Real=0.5  → Score = 1 - |0.8-0.5| = 0.70 (deficit penalised)
//   - Target=0.0, Real=0.3  → Score = 1 - |0.0-0.3| = 0.70 (homogeneity violated)
type CategoricalCriterion struct {
	name    string
	attrKey string
}

// blauNormalized computes the normalised Blau diversity index for a frequency map.
//
// The Blau Index measures heterogeneity: B = 1 - Σ(pᵢ²).
// Its raw maximum depends on the number of equally-distributed categories:
//
//	maxBlau(k) = (k-1)/k
//
// We normalise by the GLOBAL k (number of distinct category values in the entire
// population/section) rather than the LOCAL k (categories observed in the group).
// Using the local k inflates diversity for small groups that happen to contain
// several rare categories:
//
//	Example — section with 8 distinct careers, group of 7 students (5+1+1):
//	  local  k=3  → maxBlau=0.667 → normalised = 0.449/0.667 = 67%  ← WRONG
//	  global k=8  → maxBlau=0.875 → normalised = 0.449/0.875 = 51%  ← CORRECT
//
// kGlobal must be ≥ kObserved; if it is 0 or < kObserved the function falls back
// to kObserved to remain safe (legacy behaviour).
func blauNormalized(counts map[string]int, kGlobal int) float64 {
	total := 0
	for _, c := range counts {
		total += c
	}
	if total == 0 {
		return 0.0
	}

	sumSquares := 0.0
	for _, c := range counts {
		p := float64(c) / float64(total)
		sumSquares += p * p
	}
	blauRaw := 1.0 - sumSquares

	// Choose denominator: prefer global k, fall back to observed k.
	k := kGlobal
	if k < len(counts) {
		k = len(counts)
	}
	maxBlau := 1.0 - (1.0 / float64(k))
	if maxBlau <= 0 {
		// k == 1: all members share one category → zero diversity
		return 0.0
	}

	return math.Max(0.0, math.Min(1.0, blauRaw/maxBlau))
}

// CalculateRealValue returns the raw normalised Blau diversity index for the group,
// independently of the target. This is used by callers (e.g. the plugin) that need
// to store the actual diversity value separately from the absolute-error score.
//
// It accepts an optional Context so it can use the global category count from
// ctx.CategoryCounts when available, producing an accurate denominator.
//
// Returns a value in [0.0, 1.0]:
//   - 0.0 → all members share the same category (perfect homogeneity)
//   - 1.0 → members are distributed as evenly as possible across ALL known categories
func (c *CategoricalCriterion) CalculateRealValue(g *domain.Group, ctx ...domain.Context) float64 {
	if g == nil || len(*g) == 0 {
		return 0.0
	}

	counts := make(map[string]int)
	for _, student := range *g {
		if val, exists := student.CategoricalAttrs[c.attrKey]; exists {
			counts[val]++
		}
	}

	if len(counts) == 0 {
		return 0.5 // attribute missing for all members
	}

	// Resolve global k from context if provided.
	kGlobal := 0
	if len(ctx) > 0 && ctx[0].CategoryCounts != nil {
		kGlobal = ctx[0].CategoryCounts[c.attrKey]
	}

	return blauNormalized(counts, kGlobal)
}

func NewCategoricalCriterion(name, attrKey string) *CategoricalCriterion {
	return &CategoricalCriterion{
		name:    name,
		attrKey: attrKey,
	}
}

func (c *CategoricalCriterion) Name() string {
	return c.name
}

func (c *CategoricalCriterion) Calculate(g *domain.Group, ctx domain.Context) float64 {
	if g == nil || len(*g) == 0 {
		return 0.0
	}

	// Count frequency of each category value in the group.
	counts := make(map[string]int)
	for _, student := range *g {
		if val, exists := student.CategoricalAttrs[c.attrKey]; exists {
			counts[val]++
		}
	}

	if len(counts) == 0 {
		return 0.5 // attribute missing for all members
	}

	// Resolve global k from context so the denominator is stable across groups.
	// Falls back to k_observed when CategoryCounts is not populated.
	kGlobal := 0
	if ctx.CategoryCounts != nil {
		kGlobal = ctx.CategoryCounts[c.attrKey]
	}

	// normalizedBlau ∈ [0,1]: 0 = all same category, 1 = perfectly distributed
	// across ALL known categories in the population.
	normalizedBlau := blauNormalized(counts, kGlobal)

	// Get target from context (default to 0.5 if not specified)
	target := 0.5
	if t, exists := ctx.Targets[c.name]; exists {
		target = t
	} else if t, exists := ctx.Targets[c.attrKey]; exists {
		target = t
	}

	// Absolute Error scoring: Score = 1 - |Target - NormalizedBlau|
	// Both excess and deficit of diversity are penalised symmetrically.
	//
	// Examples (with correct global-k normalisation):
	//   Target=0.2, Real=0.20 (6/7 same career, k_global=8) → score ≈ 1.00
	//   Target=0.8, Real=0.80                                → score = 1.00
	//   Target=0.8, Real=1.00                                → score = 0.80
	//   Target=0.8, Real=0.50                                → score = 0.70
	//   Target=0.0, Real=0.28                                → score = 0.72
	score := 1.0 - math.Abs(target-normalizedBlau)

	return math.Max(0.0, math.Min(1.0, score))
}

// SocialCriterion evaluates group cohesion using the f2 relational cost formula:
// C'_ij = (1 - S_p * T_p(i,j)) / 2
// S_rel(G) = min(1, avg(C'_ij))
// Fitness = 1 - S_rel(G) (higher = better, lower conflict)
//
// Where:
//   - S_p ∈ [-1, 1] is the social intentionality (harmony=1, friction=-1)
//   - T_p(i,j) ∈ [-1, 1] is the raw compatibility from the sociogram
//     (1 = mutual choice, -1 = mutual rejection)
type SocialCriterion struct {
	name string
	sp   float64 // S_p: social intentionality from target [-1, 1]
}

func NewSocialCriterion(name string, sp float64) *SocialCriterion {
	return &SocialCriterion{name: name, sp: sp}
}

func (c *SocialCriterion) Name() string {
	return c.name
}

// collectEdges scans all directed edges among group members in both directions
// (A→B and B→A as independent observations). It returns:
//   - positiveSum:    sum of edge values > 0
//   - negativeCount: number of edges that equal -1 (rejections / vetoes)
//   - totalEdges:    total number of edges found (positive + negative)
func collectEdges(graph domain.SocialGraph, members []*domain.Student) (positiveSum float64, negativeCount int, totalEdges int) {
	for i := 0; i < len(members); i++ {
		for j := 0; j < len(members); j++ {
			if i == j {
				continue
			}
			peers, ok := graph[members[i].ID]
			if !ok {
				continue
			}
			val, found := peers[members[j].ID]
			if !found {
				continue
			}
			totalEdges++
			if val < 0 {
				// Rejection: -1 (or any negative value) is a veto
				negativeCount++
			} else {
				positiveSum += val
			}
		}
	}
	return
}

func (c *SocialCriterion) Calculate(g *domain.Group, ctx domain.Context) float64 {
	if g == nil || len(*g) < 2 {
		return 0.5
	}

	if ctx.SocialGraph == nil {
		return 0.5
	}

	members := *g
	pairCount := 0
	var totalCost float64

	for i := 0; i < len(members); i++ {
		for j := i + 1; j < len(members); j++ {
			idI := members[i].ID
			idJ := members[j].ID

			// In directed sociometry, the relationship between i and j consists of both
			// directed edges: i -> j (how i rates j) and j -> i (how j rates i).
			// Averaging both directions ensures symmetry: T_p(i, j) == T_p(j, i),
			// making group score invariant to member ordering in the group slice.
			eIJ := 0.0
			if peers, ok := ctx.SocialGraph[idI]; ok {
				if val, found := peers[idJ]; found {
					eIJ = val
				}
			}
			eJI := 0.0
			if peers, ok := ctx.SocialGraph[idJ]; ok {
				if val, found := peers[idI]; found {
					eJI = val
				}
			}
			Tp := (eIJ + eJI) / 2.0

			Cij := (1.0 - c.sp*Tp) / 2.0
			totalCost += Cij
			pairCount++
		}
	}

	if pairCount == 0 {
		return 0.5
	}

	avgCost := totalCost / float64(pairCount)
	Srel := math.Min(1.0, avgCost)

	return 1.0 - Srel
}

// ============================================================================
// CRITERION FACTORY & REGISTRY
// ============================================================================

// CriterionFactory manages registration and instantiation of criteria.
type CriterionFactory struct {
	registry map[string]domain.Criterion
}

func NewCriterionFactory() *CriterionFactory {
	return &CriterionFactory{
		registry: make(map[string]domain.Criterion),
	}
}

// Register adds a criterion to the factory by name.
func (f *CriterionFactory) Register(criterion domain.Criterion) {
	f.registry[criterion.Name()] = criterion
}

// Get retrieves a criterion by name. Returns nil if not found.
func (f *CriterionFactory) Get(name string) domain.Criterion {
	return f.registry[name]
}

// RegisterDefaults registers commonly used criteria.
func (f *CriterionFactory) RegisterDefaults() {
	// Numeric criteria examples
	f.Register(NewNumericCriterion("age_balance", "age", "homogeneity"))
	f.Register(NewNumericCriterion("gpa_balance", "gpa", "homogeneity"))
	f.Register(NewNumericCriterion("openness", "openness", "heterogeneity"))
	f.Register(NewNumericCriterion("conscientiousness", "conscientiousness", "homogeneity"))
	f.Register(NewNumericCriterion("extraversion", "extraversion", "heterogeneity"))
	f.Register(NewNumericCriterion("agreeableness", "agreeableness", "target"))
	f.Register(NewNumericCriterion("neuroticism", "neuroticism", "target"))

	// Categorical criteria examples
	f.Register(NewCategoricalCriterion("sex", "sex"))
	f.Register(NewCategoricalCriterion("career_code", "career_code"))
	f.Register(NewCategoricalCriterion("major", "major"))
	f.Register(NewCategoricalCriterion("language", "language"))

	// Social criterion (S_p=1.0 for harmony as default)
	f.Register(NewSocialCriterion("social", 1.0))
}

// ============================================================================
// FITNESS CALCULATION - THESIS FORMULA IMPLEMENTATION
// ============================================================================
//
// The f1 structural cost is implemented according to the thesis formula:
//
//   f_1(x) = (1/m) * Σ_j ( Σ_k w_k * φ_k(G_j) )
//   φ_k(G_j) = |T_k - D_k(G_j)|
//
// For categorical attributes:
//   D_k(G_j) = (1 - Σ p_i²) / (1 - (1/q))
//
// For numeric attributes:
//   D_k(G_j) = 1 - (1 / (1 + γ * σ²_k(G_j)))
//
// ============================================================================

// calculateGroupScore returns the weighted fitness score for a single group.
// The score is computed as the weighted average of all criterion scores:
//
//	FitnessTotal = Σ(Score_i × Weight_i) / Σ(Weight_i)
//
// Only criteria with a registered factory entry contribute to the score.
// Weights are already normalised by the caller (CalculateFitness); this function
// performs a second pass exclusively over criteria that are actually present in the
// factory so missing criteria do not skew the denominator.
func calculateGroupScore(group domain.Group, weights map[string]float64, factory *CriterionFactory, ctx domain.Context, useQuadratic bool) float64 {
	weightedSum := 0.0
	activeWeight := 0.0
	for name, weight := range weights {
		criterion := factory.Get(name)
		if criterion == nil {
			continue
		}
		s := criterion.Calculate(&group, ctx)
		if useQuadratic {
			dev := 1.0 - s
			s = 1.0 - dev*dev
		}
		weightedSum += s * weight
		activeWeight += weight
	}
	if activeWeight == 0.0 {
		return 0.0
	}
	return weightedSum / activeWeight
}

// CalculateFitness computes the weighted fitness score for a chromosome.
// Implements the thesis formula: f_1(x) = (1/m) * Σ_j ( Σ_k w_k * φ_k(G_j) )
//
// Three equity strategies are supported:
//
//   - Default (EquityStrategy=""): fitness = mean(team_scores).
//     The algorithm maximises the average quality across all teams, which can
//     lead to "sacrificing" one trash team to let the others be perfect.
//
//   - Minimax (EquityStrategy="minimax"): fitness = min(team_scores).
//     The algorithm is forced to raise the floor: the only way to improve the
//     global fitness is to improve the worst team.
//
//   - Leximin (EquityStrategy="leximin"): fitness = max(team_scores) lexicographically.
//     Ensures fairness by comparing sorted team scores from worst to best.
func CalculateFitness(chromosome *domain.Chromosome, config domain.Config, factory *CriterionFactory, ctx domain.Context) float64 {
	if chromosome == nil || len(chromosome.Groups) == 0 {
		return 0.0
	}

	// Step 1: Normalize weights
	totalWeight := 0.0
	for _, weight := range config.Weights {
		if weight > 0 {
			totalWeight += weight
		}
	}

	if totalWeight == 0.0 {
		return 0.5 // Neutral score if no valid weights
	}

	normalizedWeights := make(map[string]float64)
	for name, weight := range config.Weights {
		if weight > 0 {
			normalizedWeights[name] = weight / totalWeight
		}
	}

	// Step 2: Calculate per-group scores
	teamScores := make([]float64, len(chromosome.Groups))
	for i, group := range chromosome.Groups {
		teamScores[i] = calculateGroupScore(group, normalizedWeights, factory, ctx, config.UseQuadraticPenalty)
	}

	// Step 3: Choose aggregation strategy
	switch config.EquityStrategy {
	case "leximin":
		// Leximin: sort team scores descending and compare lexicographically
		// Create a copy and sort descending (worst team first)
		sortedScores := make([]float64, len(teamScores))
		copy(sortedScores, teamScores)
		SortScoresDescending(sortedScores)
		return math.Max(0.0, math.Min(1.0, sortedScores[0])) // Return worst team score
	case "minimax":
		// Minimax: return the score of the worst team.
		worst := teamScores[0]
		for _, s := range teamScores[1:] {
			if s < worst {
				worst = s
			}
		}
		return math.Max(0.0, math.Min(1.0, worst))
	default:
		// Default: plain average clamped to [0, 1]
		return math.Max(0.0, math.Min(1.0, calculateMean(teamScores)))
	}
}

// ComputeEquityStats computes equity statistics for a chromosome after the
// genetic run has finished. These stats are purely informational and are
// attached to the response metadata so callers can audit the result.
//
// The chromosome must have already been evaluated (Fitness is populated).
// config and factory are needed to recompute per-team scores deterministically.
// Note: scores are always computed against the user-supplied targets — this
// function never modifies any target or weight in config.
func ComputeEquityStats(chromosome *domain.Chromosome, config domain.Config, factory *CriterionFactory, ctx domain.Context) EquityStats {
	if chromosome == nil || len(chromosome.Groups) == 0 {
		return EquityStats{EquityEnforced: config.EquityStrategy != ""}
	}

	// Normalise weights (same logic as CalculateFitness)
	totalWeight := 0.0
	for _, w := range config.Weights {
		if w > 0 {
			totalWeight += w
		}
	}
	normalizedWeights := make(map[string]float64)
	if totalWeight > 0 {
		for name, w := range config.Weights {
			if w > 0 {
				normalizedWeights[name] = w / totalWeight
			}
		}
	}

	teamScores := make([]float64, len(chromosome.Groups))
	for i, group := range chromosome.Groups {
		teamScores[i] = calculateGroupScore(group, normalizedWeights, factory, ctx, config.UseQuadraticPenalty)
	}

	lowest := teamScores[0]
	highest := teamScores[0]
	for _, s := range teamScores[1:] {
		if s < lowest {
			lowest = s
		}
		if s > highest {
			highest = s
		}
	}

	variance := calculateVariance(teamScores)
	stdDev := math.Sqrt(variance)

	return EquityStats{
		LowestTeamScore:  lowest,
		HighestTeamScore: highest,
		ScoreVariance:    variance,
		ScoreStdDev:      stdDev,
		EquityEnforced:   config.EquityStrategy != "",
	}
}

// CalculateFitnessForGroup computes fitness for a single group (utility function).
func CalculateFitnessForGroup(group *domain.Group, config domain.Config, factory *CriterionFactory, ctx domain.Context) float64 {
	chromosome := &domain.Chromosome{
		Groups: []domain.Group{*group},
	}
	return CalculateFitness(chromosome, config, factory, ctx)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateVariance(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	mean := calculateMean(values)
	sumSquares := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	return sumSquares / float64(len(values))
}

// ============================================================================
// MULTI-OBJECTIVE FITNESS EVALUATION
// ============================================================================

// CriterionCategory identifies whether a criterion is psychometric or social.
type CriterionCategory string

const (
	CategoryPsychometric CriterionCategory = "psychometric"
	CategorySocial       CriterionCategory = "social"
)

// MultiObjectiveConfig holds the separated criterion weights for NSGA-II.
type MultiObjectiveConfig struct {
	PsyWeights map[string]float64
	SocWeights map[string]float64
	HasPsy     bool
	HasSoc     bool
}

// SplitCriteriaByCategory separates criterion weights into psychometric and social
// based on the provided social criterion names set.
func SplitCriteriaByCategory(weights map[string]float64, socialCriteria map[string]bool) MultiObjectiveConfig {
	psyWeights := make(map[string]float64)
	socWeights := make(map[string]float64)

	fmt.Printf("[SplitCriteriaByCategory] Input weights: %v\n", weights)
	fmt.Printf("[SplitCriteriaByCategory] Social criteria catalog: %v\n", socialCriteria)

	for name, weight := range weights {
		if weight <= 0 {
			continue
		}
		// Check if this criterion name is in the social criteria catalog
		isSocial := socialCriteria[name]
		fmt.Printf("[SplitCriteriaByCategory] Criterion '%s': weight=%.4f, isSocial=%v\n", name, weight, isSocial)
		if isSocial {
			socWeights[name] = weight
		} else {
			psyWeights[name] = weight
		}
	}

	result := MultiObjectiveConfig{
		PsyWeights: psyWeights,
		SocWeights: socWeights,
		HasPsy:     len(psyWeights) > 0,
		HasSoc:     len(socWeights) > 0,
	}
	fmt.Printf("[SplitCriteriaByCategory] Result: HasPsy=%v, HasSoc=%v, PsyWeights=%v, SocWeights=%v\n",
		result.HasPsy, result.HasSoc, result.PsyWeights, result.SocWeights)

	return result
}

// CalculateMultiObjectiveFitness evaluates a chromosome with 2 separate objectives.
// Returns a FitnessResult with PsyScore (f1) and SocScore (f2).
// Both scores are in [0, 1] where higher is better.
func CalculateMultiObjectiveFitness(chromosome *domain.Chromosome, moConfig MultiObjectiveConfig, factory *CriterionFactory, ctx domain.Context, useQuadratic bool) FitnessResult {
	if chromosome == nil || len(chromosome.Groups) == 0 {
		return FitnessResult{PsyScore: 0.0, SocScore: 0.0}
	}

	psyScores := make([]float64, len(chromosome.Groups))
	socScores := make([]float64, len(chromosome.Groups))

	for i, group := range chromosome.Groups {
		if moConfig.HasPsy {
			psyScores[i] = calculateGroupScore(group, moConfig.PsyWeights, factory, ctx, useQuadratic)
		}
		if moConfig.HasSoc {
			socScores[i] = calculateGroupScore(group, moConfig.SocWeights, factory, ctx, useQuadratic)
		}
	}

	psyAvg := calculateMean(psyScores)
	socAvg := calculateMean(socScores)

	return FitnessResult{
		PsyScore:   math.Max(0.0, math.Min(1.0, psyAvg)),
		SocScore:   math.Max(0.0, math.Min(1.0, socAvg)),
		IsFeasible: true,
	}
}
