package core

import (
	"testing"
)

// TestCalculateSociometricFitness_PerfectCohesion tests a team where all members voted for each other
func TestCalculateSociometricFitness_PerfectCohesion(t *testing.T) {
	team := []TeamMember{
		{ID: "student1"},
		{ID: "student2"},
		{ID: "student3"},
	}

	// All members vote positively for each other (mutual preference)
	adjacencyMatrix := map[string]map[string]float64{
		"student1": {"student2": 3, "student3": 3},
		"student2": {"student1": 3, "student3": 3},
		"student3": {"student1": 3, "student2": 3},
	}

	score := CalculateSociometricFitness(team, adjacencyMatrix)

	// Expected: 3*6 = 18 (each of 6 directed edges has weight 3)
	expected := 18.0
	if score != expected {
		t.Errorf("Expected score %.2f, got %.2f", expected, score)
	}
}

// TestCalculateSociometricFitness_CompleteIsolation tests a team with no votes between members
func TestCalculateSociometricFitness_CompleteIsolation(t *testing.T) {
	team := []TeamMember{
		{ID: "student1"},
		{ID: "student2"},
		{ID: "student3"},
	}

	// No votes between members
	adjacencyMatrix := map[string]map[string]float64{}

	score := CalculateSociometricFitness(team, adjacencyMatrix)

	// Expected: 0 (cohesion) - 100*3 (all isolated) = -300
	expected := -300.0
	if score != expected {
		t.Errorf("Expected score %.2f, got %.2f", expected, score)
	}
}

// TestCalculateSociometricFitness_SingleIsolate tests a team with one isolated member
func TestCalculateSociometricFitness_SingleIsolate(t *testing.T) {
	team := []TeamMember{
		{ID: "student1"},
		{ID: "student2"},
		{ID: "student3"},
	}

	// student1 and student2 vote for each other, but nobody votes for student3
	adjacencyMatrix := map[string]map[string]float64{
		"student1": {"student2": 2},
		"student2": {"student1": 2},
		"student3": {},
	}

	score := CalculateSociometricFitness(team, adjacencyMatrix)

	// Expected: 4 (cohesion: 2+2) - 100*1 (one isolated) = -96
	expected := -96.0
	if score != expected {
		t.Errorf("Expected score %.2f, got %.2f", expected, score)
	}
}

// TestCalculateSociometricFitness_VetoViolation tests the hard constraint when a veto exists
func TestCalculateSociometricFitness_VetoViolation(t *testing.T) {
	team := []TeamMember{
		{ID: "student1"},
		{ID: "student2"},
		{ID: "student3"},
	}

	// student1 explicitly rejects student2 (negative vote)
	adjacencyMatrix := map[string]map[string]float64{
		"student1": {"student2": -1, "student3": 2},
		"student2": {"student1": 2, "student3": 2},
		"student3": {"student1": 2, "student2": 2},
	}

	score := CalculateSociometricFitness(team, adjacencyMatrix)

	// Expected: -1000 (veto penalty)
	expected := -1000.0
	if score != expected {
		t.Errorf("Expected veto penalty %.2f, got %.2f", expected, score)
	}
}

// TestCalculateSociometricFitness_MixedScenario tests a realistic mixed scenario
func TestCalculateSociometricFitness_MixedScenario(t *testing.T) {
	team := []TeamMember{
		{ID: "student1"},
		{ID: "student2"},
		{ID: "student3"},
		{ID: "student4"},
	}

	// Mixed preferences: some cohesion, some isolation
	adjacencyMatrix := map[string]map[string]float64{
		"student1": {"student2": 3, "student3": 1},
		"student2": {"student1": 3, "student4": 2},
		"student3": {"student1": 1},
		"student4": {"student2": 2},
	}

	score := CalculateSociometricFitness(team, adjacencyMatrix)

	// Cohesion: 3+3+1+1+2+2 = 12
	// Isolates: student3 receives votes (1 from student1), student4 receives votes (2 from student2)
	//           student1 receives votes (3 from student2, 1 from student3)
	//           student2 receives votes (3 from student1, 2 from student4)
	// No isolates in this case
	// Expected: 12 - 0 = 12
	expected := 12.0
	if score != expected {
		t.Errorf("Expected score %.2f, got %.2f", expected, score)
	}
}

// TestCalculateSociometricFitness_MutualVotes tests mutual preference bonus
func TestCalculateSociometricFitness_MutualVotes(t *testing.T) {
	team := []TeamMember{
		{ID: "student1"},
		{ID: "student2"},
	}

	// Mutual strong preference
	adjacencyMatrix := map[string]map[string]float64{
		"student1": {"student2": 3},
		"student2": {"student1": 3},
	}

	score := CalculateSociometricFitness(team, adjacencyMatrix)

	// Expected: 3 + 3 = 6 (both directions counted)
	expected := 6.0
	if score != expected {
		t.Errorf("Expected score %.2f, got %.2f", expected, score)
	}
}

// TestCalculateSociometricFitness_EmptyTeam tests edge case of empty team
func TestCalculateSociometricFitness_EmptyTeam(t *testing.T) {
	team := []TeamMember{}
	adjacencyMatrix := map[string]map[string]float64{}

	score := CalculateSociometricFitness(team, adjacencyMatrix)

	expected := 0.0
	if score != expected {
		t.Errorf("Expected score %.2f for empty team, got %.2f", expected, score)
	}
}

// TestCalculateSociometricFitness_SingleMember tests edge case of single member team
func TestCalculateSociometricFitness_SingleMember(t *testing.T) {
	team := []TeamMember{
		{ID: "student1"},
	}
	adjacencyMatrix := map[string]map[string]float64{}

	score := CalculateSociometricFitness(team, adjacencyMatrix)

	expected := 0.0
	if score != expected {
		t.Errorf("Expected score %.2f for single member team, got %.2f", expected, score)
	}
}

// TestCalculateSociometricFitnessNormalized_Range tests that normalized scores are in [0,1]
func TestCalculateSociometricFitnessNormalized_Range(t *testing.T) {
	team := []TeamMember{
		{ID: "student1"},
		{ID: "student2"},
		{ID: "student3"},
	}

	testCases := []struct {
		name            string
		adjacencyMatrix map[string]map[string]float64
		shouldBeInRange bool
	}{
		{
			name: "Perfect cohesion",
			adjacencyMatrix: map[string]map[string]float64{
				"student1": {"student2": 3, "student3": 3},
				"student2": {"student1": 3, "student3": 3},
				"student3": {"student1": 3, "student2": 3},
			},
			shouldBeInRange: true,
		},
		{
			name:            "Complete isolation",
			adjacencyMatrix: map[string]map[string]float64{},
			shouldBeInRange: true,
		},
		{
			name: "Veto violation",
			adjacencyMatrix: map[string]map[string]float64{
				"student1": {"student2": -1},
			},
			shouldBeInRange: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			score := CalculateSociometricFitnessNormalized(team, tc.adjacencyMatrix)
			if score < 0.0 || score > 1.0 {
				t.Errorf("Normalized score %.2f is outside [0,1] range for case: %s", score, tc.name)
			}
		})
	}
}
