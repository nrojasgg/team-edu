package core

import "ms-core/internal/core/domain"

// CalculateSociometricFitness evaluates the "social health" of a team based on
// explicit student preferences (votes) from sociometric analysis.
//
// This function implements Moreno's sociometric theory (1961), calculating:
// 1. Cohesion: Rewards positive votes between team members
// 2. Isolation Penalty: Penalizes members who receive no votes from teammates
// 3. Veto Constraint: Immediately rejects teams with explicit rejections
//
// Mathematical Formula:
//   F_soc = Cohesion(t) - (P × NumIsolates)
//   where Cohesion(t) = Σ(i∈t) Σ(j∈t, i≠j) A[i][j]
//
// Hard Constraint (Veto):
//   If any A[i][j] < 0 (explicit rejection), return -1000.0
//
// Parameters:
//   - team: Array of TeamMembers to evaluate
//   - adjacencyMatrix: Vote matrix where adjacencyMatrix[voterID][targetID] represents:
//     * Positive integer: Preference/affinity strength
//     * Negative integer: Rejection/veto
//     * Zero/missing: Neutral/no data
//
// Returns:
//   - float64: Fitness score (higher is better)
//     * Positive: Good cohesion
//     * Near zero: Neutral or balanced isolation
//     * Negative: Poor cohesion or isolation issues
//     * -1000.0: Veto violation (hard constraint failure)
// getVoteValue returns the vote value from voterID toward targetID in the adjacency matrix.
// Returns 0.0 if no vote exists.
func getVoteValue(adj map[string]map[string]float64, voterID, targetID string) float64 {
	if voterMap, ok := adj[voterID]; ok {
		if v, ok := voterMap[targetID]; ok {
			return v
		}
	}
	return 0.0
}

// processVotesForMember collects all other members' votes toward team[memberAIdx].
// Returns (scoreContribution, votesReceived, vetoDetected).
func processVotesForMember(team []TeamMember, adj map[string]map[string]float64, memberAIdx int) (float64, int, bool) {
	memberA := team[memberAIdx]
	score := 0.0
	received := 0
	for j, memberB := range team {
		if j == memberAIdx {
			continue
		}
		vote := getVoteValue(adj, memberB.ID, memberA.ID)
		if vote < 0 {
			return 0, 0, true // veto detected
		}
		if vote > 0 {
			score += vote
			received++
		}
	}
	return score, received, false
}

func CalculateSociometricFitness(team []TeamMember, adjacencyMatrix map[string]map[string]float64) float64 {
	// Penalty coefficients based on sociometric theory
	const (
		penaltyWeight = 100.0  // Severe penalty for each isolated member
		vetoPenalty   = 1000.0 // "Death" of solution - constraint violation [Moreno, 1961]
	)

	// Handle edge cases
	if len(team) == 0 {
		return 0.0
	}

	if len(team) == 1 {
		// Single member team has no social dynamics to evaluate
		return 0.0
	}

	score := 0.0
	isolates := 0

	for i := range team {
		contrib, received, veto := processVotesForMember(team, adjacencyMatrix, i)
		if veto {
			return -vetoPenalty
		}
		score += contrib
		if received == 0 {
			isolates++
		}
	}

	return score - float64(isolates)*penaltyWeight
}

// CalculateSociometricFitnessNormalized returns a normalized version of the
// sociometric fitness score in the range [0, 1] for easier comparison with
// other fitness metrics.
//
// This is useful when combining sociometric fitness with OCEAN-based fitness
// in a weighted multi-objective optimization.
func CalculateSociometricFitnessNormalized(team []TeamMember, adjacencyMatrix map[string]map[string]float64) float64 {
	rawScore := CalculateSociometricFitness(team, adjacencyMatrix)

	// Handle veto case
	if rawScore <= -500.0 {
		return 0.0 // Complete failure
	}

	// Estimate theoretical maximum cohesion for normalization
	// Assume max vote value is 3 (strong preference) and all pairs are mutual
	teamSize := len(team)
	if teamSize <= 1 {
		return 0.5 // Neutral for trivial cases
	}

	// Maximum possible cohesion: all pairs vote maximally for each other
	maxPossibleCohesion := float64(teamSize * (teamSize - 1) * 3)

	// Normalize to [0, 1] range
	// Add offset to handle negative scores from isolation penalties
	offset := 100.0 // Reasonable offset for typical team sizes
	normalized := (rawScore + offset) / (maxPossibleCohesion + offset)

	// Clamp to [0, 1]
	if normalized < 0.0 {
		normalized = 0.0
	}
	if normalized > 1.0 {
		normalized = 1.0
	}

	return normalized
}

// CalculateTeamSociometricFitness is a convenience wrapper that extracts
// the adjacency matrix from TeamMember.Sociometry fields
func CalculateTeamSociometricFitness(team []TeamMember) float64 {
	// Build adjacency matrix from individual sociometry data
	adjacencyMatrix := make(map[string]map[string]float64)

	for _, member := range team {
		if len(member.Sociometry) > 0 {
			adjacencyMatrix[member.ID] = member.Sociometry
		}
	}

	return CalculateSociometricFitness(team, adjacencyMatrix)
}

// EvaluateSolution calculates multi-objective fitness for a team
// Returns FitnessResult with both psychometric and sociometric scores
func (e *TeamFormationEngine) EvaluateSolution(team []TeamMember, adjacencyMatrix map[string]map[string]float64) FitnessResult {
	// 1. Calculate Psychometric Fitness (OCEAN)
	// Extract OCEAN vectors from team members
	oceanVectors := make([]domain.OCEANVector, 0, len(team))
	for _, member := range team {
		if member.Psychometrics != nil {
			oceanVectors = append(oceanVectors, *member.Psychometrics)
		}
	}

	// Calculate psychometric score
	psyScore := 0.5 // Default neutral score
	if len(oceanVectors) > 0 {
		psyScore = domain.CalculateGroupFitness(oceanVectors)
	}

	// 2. Calculate Sociometric Fitness (Moreno)
	socScore := CalculateSociometricFitness(team, adjacencyMatrix)

	// 3. Determine Feasibility (veto check)
	isFeasible := socScore > -500.0

	return FitnessResult{
		PsychometricScore: psyScore,
		SociometricScore:  socScore,
		IsFeasible:        isFeasible,
	}
}
