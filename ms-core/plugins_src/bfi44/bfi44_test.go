package main

import (
	"testing"
)

func TestBFI44ReverseScoring(t *testing.T) {
	plugin := NewBFI44Plugin()

	inputs := map[int]int{
		1:  5,
		6:  5,
		11: 4,
		16: 4,
		21: 1,
		26: 4,
		31: 2,
		36: 5,

		2: 3, 7: 3, 12: 3, 17: 3, 22: 3, 27: 3, 32: 3, 37: 3, 42: 3,
		3: 3, 8: 3, 13: 3, 18: 3, 23: 3, 28: 3, 33: 3, 38: 3, 43: 3,
		4: 3, 9: 3, 14: 3, 19: 3, 24: 3, 29: 3, 34: 3, 39: 3,
		5: 3, 10: 3, 15: 3, 20: 3, 25: 3, 30: 3, 35: 3, 40: 3, 41: 3, 44: 3,
	}

	results, err := plugin.Calculate(inputs)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	if results.CorrectedScores[1] != 5 {
		t.Errorf("Item 1 (directo): expected 5, got %d", results.CorrectedScores[1])
	}
	if results.CorrectedScores[6] != 1 {
		t.Errorf("Item 6 (inverso): expected 1 (6-5), got %d", results.CorrectedScores[6])
	}
	if results.CorrectedScores[21] != 5 {
		t.Errorf("Item 21 (inverso): expected 5 (6-1), got %d", results.CorrectedScores[21])
	}
	if results.CorrectedScores[31] != 4 {
		t.Errorf("Item 31 (inverso): expected 4 (6-2), got %d", results.CorrectedScores[31])
	}
}

func TestBFI44MissingDataTolerance(t *testing.T) {
	plugin := NewBFI44Plugin()

	t.Run("WithinLimit_ShortDimension", func(t *testing.T) {
		inputs := generateCompleteInputs()
		delete(inputs, 31)

		_, err := plugin.Calculate(inputs)
		if err != nil {
			t.Errorf("Should not fail with 1 missing item in short dimension: %v", err)
		}
	})

	t.Run("WithinLimit_LongDimension", func(t *testing.T) {
		inputs := generateCompleteInputs()
		delete(inputs, 12)
		delete(inputs, 37)

		_, err := plugin.Calculate(inputs)
		if err != nil {
			t.Errorf("Should not fail with 2 missing items in long dimension: %v", err)
		}
	})

	t.Run("ExceedsLimit_ShortDimension", func(t *testing.T) {
		inputs := generateCompleteInputs()
		delete(inputs, 6)
		delete(inputs, 21)

		_, err := plugin.Calculate(inputs)
		if err == nil {
			t.Error("Should fail with 2 missing items in short dimension")
		}
	})

	t.Run("ExceedsLimit_LongDimension", func(t *testing.T) {
		inputs := generateCompleteInputs()
		delete(inputs, 2)
		delete(inputs, 12)
		delete(inputs, 27)

		_, err := plugin.Calculate(inputs)
		if err == nil {
			t.Error("Should fail with 3 missing items in long dimension")
		}
	})
}

func TestBFI44MeanImputation(t *testing.T) {
	plugin := NewBFI44Plugin()

	inputs := map[int]int{
		1:  4,
		6:  2,
		11: 4,
		16: 4,
		21: 2,
		26: 4,
		36: 4,

		2: 3, 7: 3, 12: 3, 17: 3, 22: 3, 27: 3, 32: 3, 37: 3, 42: 3,
		3: 3, 8: 3, 13: 3, 18: 3, 23: 3, 28: 3, 33: 3, 38: 3, 43: 3,
		4: 3, 9: 3, 14: 3, 19: 3, 24: 3, 29: 3, 34: 3, 39: 3,
		5: 3, 10: 3, 15: 3, 20: 3, 25: 3, 30: 3, 35: 3, 40: 3, 41: 3, 44: 3,
	}

	results, err := plugin.Calculate(inputs)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	if len(results.ImputedItems["E"]) != 1 || results.ImputedItems["E"][0] != 31 {
		t.Errorf("Item 31 should be in imputed items for dimension E")
	}

	if results.CorrectedScores[31] != 4 {
		t.Errorf("Item 31 should be imputed with value 4, got %d", results.CorrectedScores[31])
	}

	if results.Extraversion != 4.0 {
		t.Errorf("Extraversion should be 4.0, got %.2f", results.Extraversion)
	}
}

func TestBFI44DimensionalAverages(t *testing.T) {
	plugin := NewBFI44Plugin()

	inputs := map[int]int{
		1:  5,
		6:  1,
		11: 5,
		16: 5,
		21: 1,
		26: 5,
		31: 1,
		36: 5,

		2: 3, 7: 3, 12: 3, 17: 3, 22: 3, 27: 3, 32: 3, 37: 3, 42: 3,
		3: 3, 8: 3, 13: 3, 18: 3, 23: 3, 28: 3, 33: 3, 38: 3, 43: 3,
		4: 3, 9: 3, 14: 3, 19: 3, 24: 3, 29: 3, 34: 3, 39: 3,
		5: 3, 10: 3, 15: 3, 20: 3, 25: 3, 30: 3, 35: 3, 40: 3, 41: 3, 44: 3,
	}

	results, err := plugin.Calculate(inputs)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	if results.Extraversion != 5.0 {
		t.Errorf("Extraversion should be 5.0, got %.2f", results.Extraversion)
	}

	if results.Agreeableness != 3.0 {
		t.Errorf("Agreeableness should be 3.0, got %.2f", results.Agreeableness)
	}
	if results.Conscientiousness != 3.0 {
		t.Errorf("Conscientiousness should be 3.0, got %.2f", results.Conscientiousness)
	}
	if results.Neuroticism != 3.0 {
		t.Errorf("Neuroticism should be 3.0, got %.2f", results.Neuroticism)
	}
	if results.Openness != 3.0 {
		t.Errorf("Openness should be 3.0, got %.2f", results.Openness)
	}
}

func TestBFI44OCEANNormalization(t *testing.T) {
	plugin := NewBFI44Plugin()

	inputs := map[int]int{
		1: 5, 6: 1, 11: 5, 16: 5, 21: 1, 26: 5, 31: 1, 36: 5,

		2: 5, 7: 1, 12: 5, 17: 1, 22: 1, 27: 5, 32: 1, 37: 5, 42: 1,

		3: 3, 8: 3, 13: 3, 18: 3, 23: 3, 28: 3, 33: 3, 38: 3, 43: 3,

		4: 3, 9: 3, 14: 3, 19: 3, 24: 3, 29: 3, 34: 3, 39: 3,

		5: 3, 10: 3, 15: 3, 20: 3, 25: 3, 30: 3, 35: 3, 40: 3, 41: 3, 44: 3,
	}

	results, err := plugin.Calculate(inputs)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	tolerance := 0.01
	if abs(results.OCEANVector.Extraversion-1.0) > tolerance {
		t.Errorf("Extraversion should be normalized to 1.0, got %.3f", results.OCEANVector.Extraversion)
	}
	if abs(results.OCEANVector.Agreeableness-0.0) > tolerance {
		t.Errorf("Agreeableness should be normalized to 0.0, got %.3f", results.OCEANVector.Agreeableness)
	}
	if abs(results.OCEANVector.Conscientiousness-0.5) > tolerance {
		t.Errorf("Conscientiousness should be normalized to 0.5, got %.3f", results.OCEANVector.Conscientiousness)
	}
	if abs(results.OCEANVector.Neuroticism-0.5) > tolerance {
		t.Errorf("Neuroticism should be normalized to 0.5, got %.3f", results.OCEANVector.Neuroticism)
	}
	if abs(results.OCEANVector.Openness-0.5) > tolerance {
		t.Errorf("Openness should be normalized to 0.5, got %.3f", results.OCEANVector.Openness)
	}
}

func TestBFI44InvalidScoreRange(t *testing.T) {
	plugin := NewBFI44Plugin()

	t.Run("ScoreTooLow", func(t *testing.T) {
		inputs := generateCompleteInputs()
		inputs[1] = 0

		_, err := plugin.Calculate(inputs)
		if err == nil {
			t.Error("Should fail with score < 1")
		}
	})

	t.Run("ScoreTooHigh", func(t *testing.T) {
		inputs := generateCompleteInputs()
		inputs[1] = 6

		_, err := plugin.Calculate(inputs)
		if err == nil {
			t.Error("Should fail with score > 5")
		}
	})
}

func TestBFI44QuestionCount(t *testing.T) {
	questions := getBFI44Questions()

	if len(questions) != 44 {
		t.Errorf("Expected 44 questions, got %d", len(questions))
	}

	ids := make(map[int]bool)
	for _, q := range questions {
		if q.ID < 1 || q.ID > 44 {
			t.Errorf("Question ID %d is out of range [1, 44]", q.ID)
		}
		if ids[q.ID] {
			t.Errorf("Duplicate question ID: %d", q.ID)
		}
		ids[q.ID] = true
	}
}

func TestBFI44DimensionDistribution(t *testing.T) {
	questions := getBFI44Questions()

	distribution := map[string]int{
		"E": 0,
		"A": 0,
		"C": 0,
		"N": 0,
		"O": 0,
	}

	for _, q := range questions {
		distribution[q.Dimension]++
	}

	expected := map[string]int{
		"E": 8,
		"A": 9,
		"C": 9,
		"N": 8,
		"O": 10,
	}

	for dim, count := range expected {
		if distribution[dim] != count {
			t.Errorf("Dimension %s: expected %d items, got %d", dim, count, distribution[dim])
		}
	}
}

func TestBFI44ReverseItemsCount(t *testing.T) {
	questions := getBFI44Questions()

	reverseCount := map[string]int{
		"E": 0,
		"A": 0,
		"C": 0,
		"N": 0,
		"O": 0,
	}

	for _, q := range questions {
		if q.IsReverse {
			reverseCount[q.Dimension]++
		}
	}

	expected := map[string]int{
		"E": 3,
		"A": 4,
		"C": 4,
		"N": 3,
		"O": 2,
	}

	for dim, count := range expected {
		if reverseCount[dim] != count {
			t.Errorf("Dimension %s: expected %d reverse items, got %d", dim, count, reverseCount[dim])
		}
	}
}

func generateCompleteInputs() map[int]int {
	inputs := make(map[int]int)
	for i := 1; i <= 44; i++ {
		inputs[i] = 3
	}
	return inputs
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
