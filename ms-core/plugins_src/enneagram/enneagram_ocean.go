package main

type enneagramOCEAN struct {
	Openness          float64 `json:"openness"`
	Conscientiousness float64 `json:"conscientiousness"`
	Extraversion      float64 `json:"extraversion"`
	Agreeableness     float64 `json:"agreeableness"`
	Neuroticism       float64 `json:"neuroticism"`
}

var enneagramToOCEAN = map[int]enneagramOCEAN{
	1: {Openness: 0.30, Conscientiousness: 0.95, Extraversion: 0.40, Agreeableness: 0.50, Neuroticism: 0.60},
	2: {Openness: 0.50, Conscientiousness: 0.50, Extraversion: 0.75, Agreeableness: 0.90, Neuroticism: 0.50},
	3: {Openness: 0.50, Conscientiousness: 0.85, Extraversion: 0.80, Agreeableness: 0.40, Neuroticism: 0.30},
	4: {Openness: 0.90, Conscientiousness: 0.30, Extraversion: 0.30, Agreeableness: 0.50, Neuroticism: 0.85},
	5: {Openness: 0.95, Conscientiousness: 0.60, Extraversion: 0.15, Agreeableness: 0.40, Neuroticism: 0.50},
	6: {Openness: 0.40, Conscientiousness: 0.70, Extraversion: 0.45, Agreeableness: 0.55, Neuroticism: 0.80},
	7: {Openness: 0.85, Conscientiousness: 0.25, Extraversion: 0.90, Agreeableness: 0.60, Neuroticism: 0.30},
	8: {Openness: 0.55, Conscientiousness: 0.65, Extraversion: 0.85, Agreeableness: 0.20, Neuroticism: 0.35},
	9: {Openness: 0.50, Conscientiousness: 0.40, Extraversion: 0.25, Agreeableness: 0.85, Neuroticism: 0.20},
}

func getOCEANFromEnneagram(eneatype int) (enneagramOCEAN, bool) {
	v, ok := enneagramToOCEAN[eneatype]
	return v, ok
}

var enneagramCompatibilityMatrix = [9][9]int{
	{0, 1, 0, 1, 1, 1, 1, 0, 0},
	{1, 1, 1, 1, 1, 0, 1, 1, 0},
	{0, 1, 1, 0, 1, 0, 1, 0, 1},
	{1, 1, 0, 1, 0, 0, 0, 1, 0},
	{1, 1, 1, 0, 1, 1, 1, 0, 0},
	{1, 0, 0, 0, 1, 0, 1, 0, 1},
	{1, 1, 1, 0, 1, 1, 1, 1, 0},
	{0, 1, 0, 1, 0, 0, 1, 0, 1},
	{0, 0, 1, 0, 0, 1, 0, 1, 1},
}

func getCompatibility(e1, e2 int) int {
	if e1 < 1 || e1 > 9 || e2 < 1 || e2 > 9 {
		return 0
	}
	return enneagramCompatibilityMatrix[e1-1][e2-1]
}
