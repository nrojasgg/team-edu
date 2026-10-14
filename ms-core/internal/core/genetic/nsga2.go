package genetic

import (
	"math"
	"sort"
)

// FitnessResult holds the multi-objective scores and NSGA-II metrics
type FitnessResult struct {
	PsyScore float64 // Objective 1: Maximize (Structural Score - higher is better [0,1])
	SocScore float64 // Objective 2: Maximize (Social Score - higher is better [0,1])

	// Leximin scores: sorted team score vectors for leximin comparison
	// Scores[0] = f1 scores sorted descending (worst team first)
	// Scores[1] = f2 scores sorted descending (worst team first)
	Scores [][]float64

	// NSGA-II Attributes
	Rank       int     // Pareto Front (1, 2, 3...) - Lower is better
	Distance   float64 // Crowding Distance - Higher is better
	IsFeasible bool    // If the solution satisfies constraints
}

// CheckDominance returns true if solution A dominates solution B (for MAXIMIZATION)
// A dominates B if: (PsyA >= PsyB AND SocA >= SocB) AND (PsyA > PsyB OR SocA > SocB)
// Higher scores are better in maximization (fitness scores are in [0,1] where 1 is best)
func CheckDominance(a, b FitnessResult) bool {
	// If A violates constraints and B doesn't, B dominates automatically
	if !a.IsFeasible && b.IsFeasible {
		return false
	}
	if a.IsFeasible && !b.IsFeasible {
		return true
	}
	// If both infeasible, compare constraint violation (not implemented here, treating equal)
	if !a.IsFeasible && !b.IsFeasible {
		return false
	}

	// For MAXIMIZATION: A dominates B if A is better (higher) in at least one objective
	// and not worse (higher or equal) in any other
	betterInOne := (a.PsyScore > b.PsyScore) || (a.SocScore > b.SocScore)
	notWorseInAny := (a.PsyScore >= b.PsyScore) && (a.SocScore >= b.SocScore)

	return betterInOne && notWorseInAny
}

// CrowdedComparison is the selection operator for NSGA-II
// Returns true if 'a' is better than 'b'
func CrowdedComparison(a, b FitnessResult) bool {
	if a.Rank < b.Rank {
		return true // Prefer lower rank (better Pareto front)
	}
	if a.Rank == b.Rank {
		return a.Distance > b.Distance // If same rank, prefer larger crowding distance (diversity - maximization)
	}
	return false
}

// ConvergenceCallback is called each generation with the current Pareto front's HV
type ConvergenceCallback func(generation int, hv float64)

// Hypervolume2D calculates dominated hypervolume for a Pareto front (MAXIMIZATION: higher is better)
// refPoint should be (0.0, 0.0) as the nadir (worst possible point)
// For maximization: hv = area dominated by the front from (refX, refY) up to the front
func Hypervolume2D(front []*Individual, refX, refY float64) float64 {
	if len(front) == 0 {
		return 0.0
	}
	sortedFront := make([]*Individual, len(front))
	copy(sortedFront, front)
	// Sort ascending by PsyScore for correct dominated hypervolume calculation
	sort.Slice(sortedFront, func(i, j int) bool {
		return sortedFront[i].Fitness.PsyScore < sortedFront[j].Fitness.PsyScore
	})

	hv := 0.0
	prevX := refX // Start from the nadir reference point
	for _, ind := range sortedFront {
		f1, f2 := ind.Fitness.PsyScore, ind.Fitness.SocScore
		if f1 <= refX || f2 <= refY {
			continue
		}
		width := f1 - prevX
		hv += width * (f2 - refY)
		prevX = f1
	}
	return hv
}

// Individual represents a solution in the population
// We use an interface or a generic struct to hold the chromosome and its fitness
type Individual struct {
	Chromosome []int
	Fitness    FitnessResult
	// Internal use for sorting
	dominationCount int
	dominatedBy     []*Individual
	originalIndex   int
}

// computeDominationForOne populates p.dominatedBy and p.dominationCount
// relative to the full population.
func computeDominationForOne(p *Individual, population []*Individual) {
	p.dominationCount = 0
	p.dominatedBy = make([]*Individual, 0)
	for _, q := range population {
		if CheckDominance(p.Fitness, q.Fitness) {
			p.dominatedBy = append(p.dominatedBy, q)
		} else if CheckDominance(q.Fitness, p.Fitness) {
			p.dominationCount++
		}
	}
}

// buildNextFront advances the NSGA-II front iteration: decrements dominationCount
// for every individual dominated by members of the current front and collects
// those that reach zero (they belong to the next front).
func buildNextFront(front []*Individual, nextRank int) []*Individual {
	next := make([]*Individual, 0)
	for _, p := range front {
		for _, q := range p.dominatedBy {
			q.dominationCount--
			if q.dominationCount == 0 {
				q.Fitness.Rank = nextRank
				next = append(next, q)
			}
		}
	}
	return next
}

// FastNonDominatedSort sorts the population into Pareto fronts
func FastNonDominatedSort(population []*Individual) [][]*Individual {
	fronts := make([][]*Individual, 0)
	front1 := make([]*Individual, 0)

	for _, p := range population {
		computeDominationForOne(p, population)
		if p.dominationCount == 0 {
			p.Fitness.Rank = 1
			front1 = append(front1, p)
		}
	}

	fronts = append(fronts, front1)

	for i := 0; i < len(fronts); i++ {
		next := buildNextFront(fronts[i], i+2)
		if len(next) == 0 {
			break
		}
		fronts = append(fronts, next)
	}

	return fronts
}

// LeximinCompare compares two FitnessResult using the Leximin principle.
// Returns true if 'a' is better (leximin-dominant) than 'b'.
// Leximin first compares the worst team score (largest), then the second worst, etc.
// This is used when EnforceEquity is set to "leximin".
// For each objective, team scores are sorted in descending order and compared element-wise.
func LeximinCompare(a, b FitnessResult) bool {
	// Leximin compares each objective independently
	// For f1 (PsyScore): compare a.Scores[0] vs b.Scores[0]
	// For f2 (SocScore): compare a.Scores[1] vs b.Scores[1]
	// A solution dominates another if it's better in at least one and not worse in the other

	// Compare f1 vectors (sorted descending - worst team first)
	f1A := a.Scores[0]
	f1B := b.Scores[0]
	f2A := a.Scores[1]
	f2B := b.Scores[1]

	// Check lexicographic comparison for f1 and f2
	f1Better := false
	f1Worse := false
	f2Better := false
	f2Worse := false

	minLen := len(f1A)
	if len(f1B) < minLen {
		minLen = len(f1B)
	}

	for i := 0; i < minLen; i++ {
		if f1A[i] < f1B[i] {
			f1Better = true
		}
		if f1A[i] > f1B[i] {
			f1Worse = true
		}
		if f2A[i] < f2B[i] {
			f2Better = true
		}
		if f2A[i] > f2B[i] {
			f2Worse = true
		}
	}

	// a dominates b if better in at least one AND not worse in any
	// (same logic as Pareto dominance but applied to sorted vectors)
	notWorseInF1 := !f1Worse
	notWorseInF2 := !f2Worse
	atLeastOneBetter := f1Better || f2Better

	_ = notWorseInF1 && notWorseInF2 && atLeastOneBetter
	return (f1Better && !f1Worse && !f2Worse) || (f2Better && !f1Worse && !f2Worse) || (f1Better && f2Better)
}

// LeximinLess compares two sorted score vectors lexicographically.
// Returns true if a < b in lexicographic order (a is better).
// Assumes both vectors are sorted in descending order (largest first).
func LeximinLess(a, b []float64) bool {
	lenA := len(a)
	lenB := len(b)
	minLen := lenA
	if lenB < minLen {
		minLen = lenB
	}

	for i := 0; i < minLen; i++ {
		if a[i] < b[i] {
			return true // a is better (has smaller worst element)
		}
		if a[i] > b[i] {
			return false // a is worse
		}
		// a[i] == b[i], continue to next element
	}

	// All compared elements are equal
	// Longer vector with same prefix: the one with fewer elements is better
	// (less worst-off teams to worry about)
	return lenA < lenB
}

// SortScoresDescending sorts a slice of scores in descending order (largest first).
func SortScoresDescending(scores []float64) {
	sort.Slice(scores, func(i, j int) bool {
		return scores[i] > scores[j]
	})
}

// LeximinScores computes the leximin-sorted score vector for a chromosome.
// Returns team scores sorted in descending order (worst team first).
// This is used for both f1 and f2 objectives.
func LeximinScores(teamScores [][]float64) [][]float64 {
	// teamScores is [][2] where each row is [f1, f2] for a team
	// We need to return two sorted vectors: one for f1, one for f2

	result := make([][]float64, 2) // [f1Sorted, f2Sorted]

	f1Scores := make([]float64, len(teamScores))
	f2Scores := make([]float64, len(teamScores))
	for i, ts := range teamScores {
		f1Scores[i] = ts[0]
		f2Scores[i] = ts[1]
	}

	SortScoresDescending(f1Scores)
	SortScoresDescending(f2Scores)

	result[0] = f1Scores
	result[1] = f2Scores

	return result
}

// CalculateCrowdingDistance assigns crowding distance to individuals in a front
func CalculateCrowdingDistance(front []*Individual) {
	l := len(front)
	if l == 0 {
		return
	}

	// Initialize distance to 0
	for _, ind := range front {
		ind.Fitness.Distance = 0
	}

	// Sort by Objective 1 (PsyScore)
	sort.Slice(front, func(i, j int) bool {
		return front[i].Fitness.PsyScore < front[j].Fitness.PsyScore
	})

	// Boundary points get infinite distance
	front[0].Fitness.Distance = math.Inf(1)
	front[l-1].Fitness.Distance = math.Inf(1)

	// Range for Obj 1
	minPsy := front[0].Fitness.PsyScore
	maxPsy := front[l-1].Fitness.PsyScore
	rangePsy := maxPsy - minPsy

	if rangePsy > 0 {
		for i := 1; i < l-1; i++ {
			front[i].Fitness.Distance += (front[i+1].Fitness.PsyScore - front[i-1].Fitness.PsyScore) / rangePsy
		}
	}

	// Sort by Objective 2 (SocScore)
	sort.Slice(front, func(i, j int) bool {
		return front[i].Fitness.SocScore < front[j].Fitness.SocScore
	})

	// Boundary points (re-assert infinite, though usually covered)
	front[0].Fitness.Distance = math.Inf(1)
	front[l-1].Fitness.Distance = math.Inf(1)

	// Range for Obj 2
	minSoc := front[0].Fitness.SocScore
	maxSoc := front[l-1].Fitness.SocScore
	rangeSoc := maxSoc - minSoc

	if rangeSoc > 0 {
		for i := 1; i < l-1; i++ {
			front[i].Fitness.Distance += (front[i+1].Fitness.SocScore - front[i-1].Fitness.SocScore) / rangeSoc
		}
	}
}
