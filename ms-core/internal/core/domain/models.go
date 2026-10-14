package domain

// Student represents a learner with dynamic attributes instead of fixed fields.
// All demographic, academic, and personality data is stored in flexible maps.
type Student struct {
	ID               string
	NumericAttrs     map[string]float64 // e.g., "age", "gpa", "openness", "conscientiousness"
	CategoricalAttrs map[string]string  // e.g., "sex", "major", "language"
}

// SocialGraph represents relationships between students as a weighted adjacency map.
// Structure: SocialGraph[studentID][peerID] = affinity_score
type SocialGraph map[string]map[string]float64

// Group represents a collection of student pointers forming a team.
type Group []*Student

// Chromosome represents a candidate solution in the genetic algorithm.
// It contains multiple groups, a fitness score, and optional stats for analysis.
type Chromosome struct {
	Groups  []Group                // The team formation solution
	Fitness float64                // Overall quality score
	Stats   map[string]interface{} // Additional metadata (e.g., variance, balance metrics)
}

// Config holds the dynamic configuration for team formation and fitness evaluation.
// Weights determine which criteria matter and how much.
// The keys in Weights dictate which criteria are evaluated (e.g., "psychometric", "sociometric", "demographic").
type Config struct {
	Weights             map[string]float64 // Criterion name -> weight (e.g., {"psychometric": 0.4, "balance": 0.3})
	Targets             map[string]float64 // Target values for specific metrics (e.g., {"variance": 0.06})
	UseQuadraticPenalty bool               // Whether to apply quadratic penalties for deviations

	// EquityStrategy switches the global fitness aggregation to leximin mode when set.
	// Supported values: "", "minimax", "leximin"
	// When set to "leximin", the fitness comparison uses LeximinCompare which sorts
	// team scores in descending order (worst team first) and compares lexicographically.
	// When set to "minimax", the fitness of a chromosome equals the score of its WORST team.
	// When empty/unsupported, the default is average across all teams.
	EquityStrategy string
}

// CriterionType identifies the type of criterion for factory instantiation.
type CriterionType string

const (
	CriterionPsychometric  CriterionType = "psychometric"
	CriterionSociometric   CriterionType = "sociometric"
	CriterionBalance       CriterionType = "balance"
	CriterionDemographic   CriterionType = "demographic"
	CriterionVariance      CriterionType = "variance"
	CriterionHomogeneity   CriterionType = "homogeneity"
	CriterionHeterogeneity CriterionType = "heterogeneity"
)

// Context provides runtime context for criterion calculations.
// It can hold social graphs, global statistics, constraints, or any evaluation metadata.
type Context struct {
	SocialGraph SocialGraph
	GlobalStats map[string]interface{} // Population-level statistics
	Targets     map[string]float64     // Reference targets from Config
	Constraints map[string]interface{} // Additional constraints or parameters

	// CategoryCounts holds the total number of distinct category values observed
	// across the ENTIRE population (section) for each categorical attribute key.
	// Example: {"career_code": 8, "sex": 2}
	//
	// This is used by CategoricalCriterion to normalise the Blau index with a
	// stable denominator (maxBlau = 1 - 1/k_global) instead of the k observed
	// in a single group, which inflates diversity scores for small groups that
	// happen to contain several rare categories.
	//
	// If a key is absent or zero the criterion falls back to k_observed (legacy).
	CategoryCounts map[string]int
}

// Criterion defines the interface for any evaluation metric.
// Each criterion calculates a score for a group given a context.
// Implementations should be stateless and thread-safe.
type Criterion interface {
	// Calculate computes a normalized score [0.0, 1.0] for the given group.
	// Higher scores indicate better alignment with the criterion.
	Calculate(g *Group, ctx Context) float64

	// Name returns the identifier for this criterion (matches CriterionType or custom name).
	Name() string
}
