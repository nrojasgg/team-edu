package domain

import (
	"math"
)

// TeamMember is forward-declared here to avoid circular imports
// The actual definition is in the parent core package
type TeamMember struct {
	ID            string
	Psychometrics *OCEANVector
	Sociometry    map[string]float64
}

// OCEANStats holds statistical measures (mean and variance) for each OCEAN trait
type OCEANStats struct {
	Mean     OCEANVector
	Variance OCEANVector
}

// CalculateOCEANStats computes the mean and variance for each OCEAN trait
// across all members in a group
func CalculateOCEANStats(members []OCEANVector) OCEANStats {
	if len(members) == 0 {
		return OCEANStats{}
	}

	n := float64(len(members))
	stats := OCEANStats{}

	// Calculate means
	for _, member := range members {
		stats.Mean.Openness += member.Openness
		stats.Mean.Conscientiousness += member.Conscientiousness
		stats.Mean.Extraversion += member.Extraversion
		stats.Mean.Agreeableness += member.Agreeableness
		stats.Mean.Neuroticism += member.Neuroticism
	}

	stats.Mean.Openness /= n
	stats.Mean.Conscientiousness /= n
	stats.Mean.Extraversion /= n
	stats.Mean.Agreeableness /= n
	stats.Mean.Neuroticism /= n

	// Calculate variances
	for _, member := range members {
		stats.Variance.Openness += math.Pow(member.Openness-stats.Mean.Openness, 2)
		stats.Variance.Conscientiousness += math.Pow(member.Conscientiousness-stats.Mean.Conscientiousness, 2)
		stats.Variance.Extraversion += math.Pow(member.Extraversion-stats.Mean.Extraversion, 2)
		stats.Variance.Agreeableness += math.Pow(member.Agreeableness-stats.Mean.Agreeableness, 2)
		stats.Variance.Neuroticism += math.Pow(member.Neuroticism-stats.Mean.Neuroticism, 2)
	}

	stats.Variance.Openness /= n
	stats.Variance.Conscientiousness /= n
	stats.Variance.Extraversion /= n
	stats.Variance.Agreeableness /= n
	stats.Variance.Neuroticism /= n

	return stats
}

// CalculateGroupFitness evaluates the quality of a team based on OCEAN personality traits
// using heuristic rules for each trait dimension
//
// Heuristic Rules:
// - Conscientiousness (C): Homogeneity is desired (low variance = high score)
// - Extraversion (E): Heterogeneity is desired (optimal variance ~0.06)
// - Agreeableness (A): High mean is desired (quadratic reward)
// - Neuroticism (N): Low mean is desired (linear penalty)
//
// Returns a fitness score between 0.0 and 1.0
func CalculateGroupFitness(members []OCEANVector) float64 {
	if len(members) == 0 {
		return 0.0
	}

	// 1. Calculate statistical vectors (mean and variance for O, C, E, A, N)
	stats := CalculateOCEANStats(members)

	// 2. Apply heuristic rules

	// Conscientiousness: Homogeneity (minimize variance)
	// Justification: Disparate work ethics create conflict
	// Formula: Score_C = 1 / (1 + σ²_C × 10)
	// Strongly penalizes high variance
	scoreC := 1.0 / (1.0 + (stats.Variance.Conscientiousness * 10.0))

	// Extraversion: Heterogeneity/Complementarity (optimal variance)
	// Justification: Teams need both leaders (E+) and listeners (E-)
	// All E+ compete; all E- don't initiate
	// Formula: Score_E = 1 - |σ²_E - 0.06|
	// Seeks ideal variance (~0.06, equivalent to normal distribution)
	// Penalizes both total homogeneity and extreme chaos
	targetVariance := 0.06
	scoreE := 1.0 - math.Abs(stats.Variance.Extraversion-targetVariance)
	if scoreE < 0 {
		scoreE = 0
	}

	// Agreeableness: Maximize mean
	// Justification: Average agreeableness predicts cohesion and psychological safety
	// Formula: Score_A = μ_A²
	// Quadratic to reward highly agreeable groups
	scoreA := stats.Mean.Agreeableness * stats.Mean.Agreeableness

	// Neuroticism: Minimize mean
	// Justification: High neuroticism "infects" the group with anxiety
	// Formula: Score_N = 1 - μ_N
	scoreN := 1.0 - stats.Mean.Neuroticism

	// 3. Weighted sum (weights adjustable by configuration)
	// Priority given to Conscientiousness (Performance) and Neuroticism (Stability)
	finalScore := (scoreC * 0.30) +
		(scoreE * 0.20) +
		(scoreA * 0.20) +
		(scoreN * 0.30)

	// Ensure score is within [0, 1]
	if finalScore < 0 {
		finalScore = 0
	}
	if finalScore > 1 {
		finalScore = 1
	}

	return finalScore
}

// CalculateGroupFitnessWeighted is a variant that allows custom weights for each trait
func CalculateGroupFitnessWeighted(members []OCEANVector, weights OCEANVector) float64 {
	if len(members) == 0 {
		return 0.0
	}

	stats := CalculateOCEANStats(members)

	// Apply the same heuristic rules
	scoreC := 1.0 / (1.0 + (stats.Variance.Conscientiousness * 10.0))

	targetVariance := 0.06
	scoreE := 1.0 - math.Abs(stats.Variance.Extraversion-targetVariance)
	if scoreE < 0 {
		scoreE = 0
	}

	scoreA := stats.Mean.Agreeableness * stats.Mean.Agreeableness
	scoreN := 1.0 - stats.Mean.Neuroticism

	// Openness: For weighted version, we can add a rule for Openness
	// Default: Moderate diversity is good (similar to Extraversion but less critical)
	targetVarianceO := 0.05
	scoreO := 1.0 - math.Abs(stats.Variance.Openness-targetVarianceO)
	if scoreO < 0 {
		scoreO = 0
	}

	// Weighted sum using custom weights
	totalWeight := weights.Openness + weights.Conscientiousness + weights.Extraversion +
		weights.Agreeableness + weights.Neuroticism

	if totalWeight == 0 {
		return 0.0
	}

	finalScore := (scoreO*weights.Openness +
		scoreC*weights.Conscientiousness +
		scoreE*weights.Extraversion +
		scoreA*weights.Agreeableness +
		scoreN*weights.Neuroticism) / totalWeight

	// Ensure score is within [0, 1]
	if finalScore < 0 {
		finalScore = 0
	}
	if finalScore > 1 {
		finalScore = 1
	}

	return finalScore
}

// CalculateTeamPsychometricFitness calculates psychometric fitness from TeamMembers
// Extracts OCEAN vectors and calls CalculateGroupFitness
func CalculateTeamPsychometricFitness(members []TeamMember) float64 {
	if len(members) == 0 {
		return 0.0
	}

	// Extract OCEAN vectors from team members
	oceanVectors := make([]OCEANVector, 0, len(members))
	for _, member := range members {
		if member.Psychometrics != nil {
			oceanVectors = append(oceanVectors, *member.Psychometrics)
		}
	}

	// If no psychometric data available, return neutral score
	if len(oceanVectors) == 0 {
		return 0.5
	}

	return CalculateGroupFitness(oceanVectors)
}
