package main

import (
	"fmt"
	"time"

	"ms-core/internal/core"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func parseOptionalDate(request map[string]interface{}, key string) *time.Time {
	str, ok := request[key].(string)
	if !ok || str == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, str)
	if err != nil {
		t, err = time.Parse("2006-01-02", str)
		if err != nil {
			return nil
		}
	}
	return &t
}

func parseObjectID(s string) (primitive.ObjectID, error) {
	id, err := primitive.ObjectIDFromHex(s)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("invalid ObjectID: %w", err)
	}
	return id, nil
}

func currentTime() time.Time {
	return time.Now().UTC()
}

func CalculateLikertDimension(questions map[string]MyQuestion, answers map[string]int, dimension string) float64 {
	var sum float64
	count := 0

	for id, q := range questions {
		if q.Dimension != dimension {
			continue
		}
		score, exists := answers[id]
		if !exists {
			continue
		}
		if q.IsReverse {
			score = 6 - score
		}
		sum += float64(score)
		count++
	}

	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

func NormalizeToOcean(rawScores map[string]float64) core.OCEANVector {
	normalize := func(score float64) float64 {
		return (score - 1) / 4
	}

	return core.OCEANVector{
		Extraversion:      normalize(rawScores["E"]),
		Agreeableness:     normalize(rawScores["A"]),
		Conscientiousness: normalize(rawScores["C"]),
		Neuroticism:       normalize(rawScores["N"]),
		Openness:          normalize(rawScores["O"]),
	}
}
