package domain

import (
	"fmt"
	"math"
)

// OCEANVector represents the Big Five personality traits normalized to 0.0-1.0
// O - Openness to experience
// C - Conscientiousness
// E - Extraversion
// A - Agreeableness
// N - Neuroticism
type OCEANVector struct {
	Openness          float64 `json:"openness" bson:"openness"`
	Conscientiousness float64 `json:"conscientiousness" bson:"conscientiousness"`
	Extraversion      float64 `json:"extraversion" bson:"extraversion"`
	Agreeableness     float64 `json:"agreeableness" bson:"agreeableness"`
	Neuroticism       float64 `json:"neuroticism" bson:"neuroticism"`
}

// Validate ensures all traits are within [0.0, 1.0] range
func (o *OCEANVector) Validate() error {
	traits := map[string]float64{
		"openness":          o.Openness,
		"conscientiousness": o.Conscientiousness,
		"extraversion":      o.Extraversion,
		"agreeableness":     o.Agreeableness,
		"neuroticism":       o.Neuroticism,
	}

	for name, value := range traits {
		if value < 0.0 || value > 1.0 {
			return fmt.Errorf("%s must be between 0.0 and 1.0, got: %.2f", name, value)
		}
	}
	return nil
}

// EuclideanDistance calculates the Euclidean distance between two OCEAN vectors
// Returns a value between 0.0 (identical) and ~2.24 (maximum distance in 5D unit hypercube)
func (o *OCEANVector) EuclideanDistance(other *OCEANVector) float64 {
	if other == nil {
		return 0
	}

	dO := o.Openness - other.Openness
	dC := o.Conscientiousness - other.Conscientiousness
	dE := o.Extraversion - other.Extraversion
	dA := o.Agreeableness - other.Agreeableness
	dN := o.Neuroticism - other.Neuroticism

	return math.Sqrt(dO*dO + dC*dC + dE*dE + dA*dA + dN*dN)
}

// CosineSimilarity calculates the cosine similarity between two OCEAN vectors
// Returns a value between -1.0 (opposite) and 1.0 (identical direction)
func (o *OCEANVector) CosineSimilarity(other *OCEANVector) float64 {
	if other == nil {
		return 0
	}

	dotProduct := o.Openness*other.Openness +
		o.Conscientiousness*other.Conscientiousness +
		o.Extraversion*other.Extraversion +
		o.Agreeableness*other.Agreeableness +
		o.Neuroticism*other.Neuroticism

	magnitudeA := math.Sqrt(o.Openness*o.Openness +
		o.Conscientiousness*o.Conscientiousness +
		o.Extraversion*o.Extraversion +
		o.Agreeableness*o.Agreeableness +
		o.Neuroticism*o.Neuroticism)

	magnitudeB := math.Sqrt(other.Openness*other.Openness +
		other.Conscientiousness*other.Conscientiousness +
		other.Extraversion*other.Extraversion +
		other.Agreeableness*other.Agreeableness +
		other.Neuroticism*other.Neuroticism)

	if magnitudeA == 0 || magnitudeB == 0 {
		return 0
	}

	return dotProduct / (magnitudeA * magnitudeB)
}

// Normalize ensures the vector has unit length
// This is useful for comparing personality profiles independent of intensity
func (o *OCEANVector) Normalize() {
	magnitude := math.Sqrt(o.Openness*o.Openness +
		o.Conscientiousness*o.Conscientiousness +
		o.Extraversion*o.Extraversion +
		o.Agreeableness*o.Agreeableness +
		o.Neuroticism*o.Neuroticism)

	if magnitude > 0 {
		o.Openness /= magnitude
		o.Conscientiousness /= magnitude
		o.Extraversion /= magnitude
		o.Agreeableness /= magnitude
		o.Neuroticism /= magnitude
	}
}

// Clone creates a deep copy of the OCEAN vector
func (o *OCEANVector) Clone() *OCEANVector {
	return &OCEANVector{
		Openness:          o.Openness,
		Conscientiousness: o.Conscientiousness,
		Extraversion:      o.Extraversion,
		Agreeableness:     o.Agreeableness,
		Neuroticism:       o.Neuroticism,
	}
}
