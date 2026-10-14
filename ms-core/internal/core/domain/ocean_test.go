package domain

import (
	"math"
	"testing"
)

// ---------------------------------------------------------------------------
// OCEANVector.Validate
// ---------------------------------------------------------------------------

func TestOCEANVector_Validate_Valid(t *testing.T) {
	v := &OCEANVector{
		Openness:          0.5,
		Conscientiousness: 0.7,
		Extraversion:      0.4,
		Agreeableness:     0.9,
		Neuroticism:       0.3,
	}
	if err := v.Validate(); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestOCEANVector_Validate_ExactBoundaries(t *testing.T) {
	v := &OCEANVector{
		Openness:          0.0,
		Conscientiousness: 1.0,
		Extraversion:      0.0,
		Agreeableness:     1.0,
		Neuroticism:       0.0,
	}
	if err := v.Validate(); err != nil {
		t.Errorf("expected no error at boundaries, got: %v", err)
	}
}

func TestOCEANVector_Validate_OpenessAboveOne(t *testing.T) {
	v := &OCEANVector{Openness: 1.1}
	if err := v.Validate(); err == nil {
		t.Error("expected error for openness > 1.0")
	}
}

func TestOCEANVector_Validate_NegativeValue(t *testing.T) {
	v := &OCEANVector{Conscientiousness: -0.1}
	if err := v.Validate(); err == nil {
		t.Error("expected error for negative value")
	}
}

func TestOCEANVector_Validate_NeuroticismAboveOne(t *testing.T) {
	v := &OCEANVector{Neuroticism: 1.5}
	if err := v.Validate(); err == nil {
		t.Error("expected error for neuroticism > 1.0")
	}
}

// ---------------------------------------------------------------------------
// OCEANVector.EuclideanDistance
// ---------------------------------------------------------------------------

func TestOCEANVector_EuclideanDistance_ZeroForIdentical(t *testing.T) {
	a := &OCEANVector{0.5, 0.5, 0.5, 0.5, 0.5}
	if d := a.EuclideanDistance(a); d != 0.0 {
		t.Errorf("expected 0.0 for identical vectors, got %f", d)
	}
}

func TestOCEANVector_EuclideanDistance_NilOther(t *testing.T) {
	a := &OCEANVector{0.5, 0.5, 0.5, 0.5, 0.5}
	if d := a.EuclideanDistance(nil); d != 0.0 {
		t.Errorf("expected 0 for nil other, got %f", d)
	}
}

func TestOCEANVector_EuclideanDistance_KnownValue(t *testing.T) {
	a := &OCEANVector{1.0, 0.0, 0.0, 0.0, 0.0}
	b := &OCEANVector{0.0, 0.0, 0.0, 0.0, 0.0}
	d := a.EuclideanDistance(b)
	if math.Abs(d-1.0) > 1e-9 {
		t.Errorf("expected 1.0, got %f", d)
	}
}

func TestOCEANVector_EuclideanDistance_MaxDistance(t *testing.T) {
	a := &OCEANVector{1.0, 1.0, 1.0, 1.0, 1.0}
	b := &OCEANVector{0.0, 0.0, 0.0, 0.0, 0.0}
	// sqrt(5) ≈ 2.236
	d := a.EuclideanDistance(b)
	if math.Abs(d-math.Sqrt(5)) > 1e-9 {
		t.Errorf("expected sqrt(5) ~2.236, got %f", d)
	}
}

// ---------------------------------------------------------------------------
// OCEANVector.CosineSimilarity
// ---------------------------------------------------------------------------

func TestOCEANVector_CosineSimilarity_IdenticalVectors(t *testing.T) {
	v := &OCEANVector{0.8, 0.6, 0.4, 0.7, 0.3}
	sim := v.CosineSimilarity(v)
	if math.Abs(sim-1.0) > 1e-9 {
		t.Errorf("expected 1.0 for identical vectors, got %f", sim)
	}
}

func TestOCEANVector_CosineSimilarity_NilOther(t *testing.T) {
	a := &OCEANVector{0.5, 0.5, 0.5, 0.5, 0.5}
	if sim := a.CosineSimilarity(nil); sim != 0.0 {
		t.Errorf("expected 0 for nil, got %f", sim)
	}
}

func TestOCEANVector_CosineSimilarity_ZeroVector(t *testing.T) {
	a := &OCEANVector{0.0, 0.0, 0.0, 0.0, 0.0}
	b := &OCEANVector{0.5, 0.5, 0.5, 0.5, 0.5}
	if sim := a.CosineSimilarity(b); sim != 0.0 {
		t.Errorf("expected 0 for zero vector, got %f", sim)
	}
}

// ---------------------------------------------------------------------------
// OCEANVector.Normalize
// ---------------------------------------------------------------------------

func TestOCEANVector_Normalize_ProducesUnitLength(t *testing.T) {
	v := &OCEANVector{0.8, 0.6, 0.4, 0.7, 0.3}
	v.Normalize()
	magnitude := math.Sqrt(
		v.Openness*v.Openness +
			v.Conscientiousness*v.Conscientiousness +
			v.Extraversion*v.Extraversion +
			v.Agreeableness*v.Agreeableness +
			v.Neuroticism*v.Neuroticism,
	)
	if math.Abs(magnitude-1.0) > 1e-9 {
		t.Errorf("expected unit length after Normalize, got magnitude %f", magnitude)
	}
}

func TestOCEANVector_Normalize_ZeroVector(t *testing.T) {
	v := &OCEANVector{0.0, 0.0, 0.0, 0.0, 0.0}
	// Should not panic
	v.Normalize()
}

// ---------------------------------------------------------------------------
// OCEANVector.Clone
// ---------------------------------------------------------------------------

func TestOCEANVector_Clone_IsDeepCopy(t *testing.T) {
	orig := &OCEANVector{0.1, 0.2, 0.3, 0.4, 0.5}
	clone := orig.Clone()
	if clone == orig {
		t.Error("expected different pointer")
	}
	if *clone != *orig {
		t.Errorf("expected clone to equal original: orig=%v, clone=%v", *orig, *clone)
	}
	// Mutating clone should not affect orig
	clone.Openness = 0.9
	if orig.Openness != 0.1 {
		t.Error("expected original to be unchanged after mutating clone")
	}
}

// ---------------------------------------------------------------------------
// CalculateTeamPsychometricFitness
// ---------------------------------------------------------------------------

func TestCalculateTeamPsychometricFitness_Empty(t *testing.T) {
	score := CalculateTeamPsychometricFitness(nil)
	if score != 0.0 {
		t.Errorf("expected 0.0 for nil, got %f", score)
	}
}

func TestCalculateTeamPsychometricFitness_EmptySlice(t *testing.T) {
	score := CalculateTeamPsychometricFitness([]TeamMember{})
	if score != 0.0 {
		t.Errorf("expected 0.0 for empty slice, got %f", score)
	}
}

func TestCalculateTeamPsychometricFitness_NoPsychometrics(t *testing.T) {
	members := []TeamMember{
		{ID: "A", Psychometrics: nil},
		{ID: "B", Psychometrics: nil},
	}
	score := CalculateTeamPsychometricFitness(members)
	// Neutral when no psychometric data
	if score != 0.5 {
		t.Errorf("expected 0.5 for no-ocean data, got %f", score)
	}
}

func TestCalculateTeamPsychometricFitness_WithOCEAN(t *testing.T) {
	members := []TeamMember{
		{
			ID: "A",
			Psychometrics: &OCEANVector{
				Openness: 0.7, Conscientiousness: 0.8, Extraversion: 0.6,
				Agreeableness: 0.9, Neuroticism: 0.2,
			},
		},
		{
			ID: "B",
			Psychometrics: &OCEANVector{
				Openness: 0.6, Conscientiousness: 0.75, Extraversion: 0.4,
				Agreeableness: 0.85, Neuroticism: 0.25,
			},
		},
	}
	score := CalculateTeamPsychometricFitness(members)
	if score < 0.0 || score > 1.0 {
		t.Errorf("expected score in [0,1], got %f", score)
	}
}

func TestCalculateTeamPsychometricFitness_MixedOCEAN(t *testing.T) {
	// Some members with and without OCEAN data
	members := []TeamMember{
		{
			ID: "A",
			Psychometrics: &OCEANVector{
				Openness: 0.7, Conscientiousness: 0.8, Extraversion: 0.6,
				Agreeableness: 0.9, Neuroticism: 0.2,
			},
		},
		{ID: "B", Psychometrics: nil},
	}
	score := CalculateTeamPsychometricFitness(members)
	// Only one OCEAN vector, still should be in [0,1]
	if score < 0.0 || score > 1.0 {
		t.Errorf("expected score in [0,1], got %f", score)
	}
}

// ---------------------------------------------------------------------------
// CalculateGroupFitnessWeighted – extra cases
// ---------------------------------------------------------------------------

func TestCalculateGroupFitnessWeighted_ZeroWeights(t *testing.T) {
	members := []OCEANVector{
		{0.5, 0.7, 0.6, 0.8, 0.3},
		{0.6, 0.65, 0.5, 0.75, 0.35},
	}
	weights := OCEANVector{0.0, 0.0, 0.0, 0.0, 0.0}
	if score := CalculateGroupFitnessWeighted(members, weights); score != 0.0 {
		t.Errorf("expected 0.0 for zero weights, got %f", score)
	}
}

func TestCalculateGroupFitnessWeighted_EmptyMembers(t *testing.T) {
	weights := OCEANVector{0.2, 0.2, 0.2, 0.2, 0.2}
	if score := CalculateGroupFitnessWeighted(nil, weights); score != 0.0 {
		t.Errorf("expected 0.0 for nil members, got %f", score)
	}
}

// ---------------------------------------------------------------------------
// CalculateOCEANStats – edge cases
// ---------------------------------------------------------------------------

func TestCalculateOCEANStats_SingleMember(t *testing.T) {
	members := []OCEANVector{{0.5, 0.7, 0.6, 0.8, 0.3}}
	stats := CalculateOCEANStats(members)
	if stats.Mean.Openness != 0.5 {
		t.Errorf("expected mean openness 0.5, got %f", stats.Mean.Openness)
	}
	// Variance of single element should be 0
	if stats.Variance.Openness != 0.0 {
		t.Errorf("expected variance 0 for single element, got %f", stats.Variance.Openness)
	}
}

func TestCalculateOCEANStats_TwoIdenticalMembers(t *testing.T) {
	v := OCEANVector{0.8, 0.7, 0.6, 0.5, 0.4}
	members := []OCEANVector{v, v}
	stats := CalculateOCEANStats(members)
	if math.Abs(stats.Mean.Openness-0.8) > 1e-9 {
		t.Errorf("expected mean 0.8, got %f", stats.Mean.Openness)
	}
	if stats.Variance.Conscientiousness != 0.0 {
		t.Errorf("expected 0 variance for identical members, got %f", stats.Variance.Conscientiousness)
	}
}
