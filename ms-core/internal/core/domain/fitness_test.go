package domain

import (
	"testing"
)

func TestCalculateOCEANStats(t *testing.T) {
	// Test with sample team members
	members := []OCEANVector{
		{Openness: 0.7, Conscientiousness: 0.8, Extraversion: 0.6, Agreeableness: 0.9, Neuroticism: 0.3},
		{Openness: 0.5, Conscientiousness: 0.7, Extraversion: 0.4, Agreeableness: 0.8, Neuroticism: 0.4},
		{Openness: 0.6, Conscientiousness: 0.75, Extraversion: 0.5, Agreeableness: 0.85, Neuroticism: 0.35},
	}

	stats := CalculateOCEANStats(members)

	// Check means (approximate)
	expectedMeanC := (0.8 + 0.7 + 0.75) / 3.0
	if stats.Mean.Conscientiousness < expectedMeanC-0.01 || stats.Mean.Conscientiousness > expectedMeanC+0.01 {
		t.Errorf("Expected Conscientiousness mean ~%.2f, got %.2f", expectedMeanC, stats.Mean.Conscientiousness)
	}

	// Variance should be positive
	if stats.Variance.Conscientiousness < 0 {
		t.Errorf("Variance should be non-negative, got %.4f", stats.Variance.Conscientiousness)
	}
}

func TestCalculateGroupFitness(t *testing.T) {
	tests := []struct {
		name    string
		members []OCEANVector
		wantMin float64
		wantMax float64
	}{
		{
			name: "Highly compatible team",
			members: []OCEANVector{
				{Openness: 0.7, Conscientiousness: 0.8, Extraversion: 0.6, Agreeableness: 0.9, Neuroticism: 0.2},
				{Openness: 0.6, Conscientiousness: 0.75, Extraversion: 0.4, Agreeableness: 0.85, Neuroticism: 0.25},
				{Openness: 0.65, Conscientiousness: 0.78, Extraversion: 0.5, Agreeableness: 0.88, Neuroticism: 0.22},
			},
			wantMin: 0.5,
			wantMax: 1.0,
		},
		{
			name: "High neuroticism team (should score lower)",
			members: []OCEANVector{
				{Openness: 0.7, Conscientiousness: 0.8, Extraversion: 0.6, Agreeableness: 0.9, Neuroticism: 0.8},
				{Openness: 0.6, Conscientiousness: 0.75, Extraversion: 0.4, Agreeableness: 0.85, Neuroticism: 0.85},
				{Openness: 0.65, Conscientiousness: 0.78, Extraversion: 0.5, Agreeableness: 0.88, Neuroticism: 0.82},
			},
			wantMin: 0.0,
			wantMax: 0.75,
		},
		{
			name: "Low agreeableness team (should score lower)",
			members: []OCEANVector{
				{Openness: 0.7, Conscientiousness: 0.8, Extraversion: 0.6, Agreeableness: 0.2, Neuroticism: 0.2},
				{Openness: 0.6, Conscientiousness: 0.75, Extraversion: 0.4, Agreeableness: 0.25, Neuroticism: 0.25},
				{Openness: 0.65, Conscientiousness: 0.78, Extraversion: 0.5, Agreeableness: 0.22, Neuroticism: 0.22},
			},
			wantMin: 0.0,
			wantMax: 0.75,
		},
		{
			name: "High variance in conscientiousness (should score lower)",
			members: []OCEANVector{
				{Openness: 0.7, Conscientiousness: 0.1, Extraversion: 0.6, Agreeableness: 0.9, Neuroticism: 0.2},
				{Openness: 0.6, Conscientiousness: 0.9, Extraversion: 0.4, Agreeableness: 0.85, Neuroticism: 0.25},
				{Openness: 0.65, Conscientiousness: 0.5, Extraversion: 0.5, Agreeableness: 0.88, Neuroticism: 0.22},
			},
			wantMin: 0.0,
			wantMax: 0.75,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := CalculateGroupFitness(tt.members)
			if score < tt.wantMin || score > tt.wantMax {
				t.Errorf("CalculateGroupFitness() = %.4f, want between %.2f and %.2f", score, tt.wantMin, tt.wantMax)
			}
			t.Logf("%s: fitness score = %.4f", tt.name, score)
		})
	}
}

func TestCalculateGroupFitnessEmpty(t *testing.T) {
	members := []OCEANVector{}
	score := CalculateGroupFitness(members)
	if score != 0.0 {
		t.Errorf("Expected 0.0 for empty group, got %.4f", score)
	}
}

func TestCalculateGroupFitnessWeighted(t *testing.T) {
	members := []OCEANVector{
		{Openness: 0.7, Conscientiousness: 0.8, Extraversion: 0.6, Agreeableness: 0.9, Neuroticism: 0.2},
		{Openness: 0.6, Conscientiousness: 0.75, Extraversion: 0.4, Agreeableness: 0.85, Neuroticism: 0.25},
		{Openness: 0.65, Conscientiousness: 0.78, Extraversion: 0.5, Agreeableness: 0.88, Neuroticism: 0.22},
	}

	// Custom weights emphasizing Conscientiousness and Neuroticism
	weights := OCEANVector{
		Openness:          0.1,
		Conscientiousness: 0.3,
		Extraversion:      0.2,
		Agreeableness:     0.1,
		Neuroticism:       0.3,
	}

	score := CalculateGroupFitnessWeighted(members, weights)
	if score < 0.0 || score > 1.0 {
		t.Errorf("Score should be between 0 and 1, got %.4f", score)
	}
	t.Logf("Weighted fitness score = %.4f", score)
}
