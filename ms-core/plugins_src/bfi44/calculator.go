package main

import (
	"fmt"
)

func getBFI44Questions() []BFI44Question {
	return []BFI44Question{
		{ID: 1, Text: "Me veo como alguien que es hablador/a.", Dimension: "E", IsReverse: false, Required: true},
		{ID: 2, Text: "Me veo como alguien que tiende a ser callado/a.", Dimension: "E", IsReverse: true, Required: true},
		{ID: 3, Text: "Me veo como alguien que es lleno/a de energía.", Dimension: "E", IsReverse: false, Required: true},
		{ID: 4, Text: "Me veo como alguien que genera mucho entusiasmo.", Dimension: "E", IsReverse: false, Required: true},
		{ID: 5, Text: "Me veo como alguien que tiene una personalidad asertiva.", Dimension: "E", IsReverse: false, Required: true},
		{ID: 6, Text: "Me veo como alguien que es tímido/a e inhibido/a.", Dimension: "E", IsReverse: true, Required: true},
		{ID: 7, Text: "Me veo como alguien que es sociable y extrovertido/a.", Dimension: "E", IsReverse: false, Required: true},
		{ID: 8, Text: "Me veo como alguien reservado/a.", Dimension: "E", IsReverse: true, Required: true},

		{ID: 9, Text: "Me veo como alguien que tiene una naturaleza indulgente.", Dimension: "A", IsReverse: false, Required: true},
		{ID: 10, Text: "Me veo como alguien que es de ayuda y sin egoísmo con los demás.", Dimension: "A", IsReverse: false, Required: true},
		{ID: 11, Text: "Me veo como alguien que empieza peleas con otros.", Dimension: "A", IsReverse: true, Required: true},
		{ID: 12, Text: "Me veo como alguien que tiene un corazón blando.", Dimension: "A", IsReverse: false, Required: true},
		{ID: 13, Text: "Me veo como alguien que es confiable y de fiar.", Dimension: "A", IsReverse: false, Required: true},
		{ID: 14, Text: "Me veo como alguien que puede ser frío/a y distante.", Dimension: "A", IsReverse: true, Required: true},
		{ID: 15, Text: "Me veo como alguien que es considerado/a y amable con casi todo el mundo.", Dimension: "A", IsReverse: false, Required: true},
		{ID: 16, Text: "Me veo como alguien que a veces es rudo/a con los demás.", Dimension: "A", IsReverse: true, Required: true},
		{ID: 42, Text: "Me veo como alguien que le gusta cooperar con otros.", Dimension: "A", IsReverse: false, Required: true},
		{ID: 43, Text: "Me veo como alguien que es a veces descortés con los demás.", Dimension: "A", IsReverse: true, Required: true},

		{ID: 17, Text: "Me veo como alguien que hace un trabajo minucioso.", Dimension: "C", IsReverse: false, Required: true},
		{ID: 18, Text: "Me veo como alguien que tiende a ser desorganizado/a.", Dimension: "C", IsReverse: true, Required: true},
		{ID: 19, Text: "Me veo como alguien que puede ser algo descuidado/a.", Dimension: "C", IsReverse: true, Required: true},
		{ID: 20, Text: "Me veo como alguien que es un/a trabajador/a confiable.", Dimension: "C", IsReverse: false, Required: true},
		{ID: 21, Text: "Me veo como alguien que persevera hasta que la tarea esté terminada.", Dimension: "C", IsReverse: false, Required: true},
		{ID: 22, Text: "Me veo como alguien que tiende a ser perezoso/a.", Dimension: "C", IsReverse: true, Required: true},
		{ID: 23, Text: "Me veo como alguien que hace planes y los lleva a cabo.", Dimension: "C", IsReverse: false, Required: true},
		{ID: 24, Text: "Me veo como alguien que se distrae fácilmente.", Dimension: "C", IsReverse: true, Required: true},
		{ID: 25, Text: "Me veo como alguien que es eficiente en el trabajo.", Dimension: "C", IsReverse: false, Required: true},

		{ID: 26, Text: "Me veo como alguien que está deprimido/a, triste.", Dimension: "N", IsReverse: false, Required: true},
		{ID: 27, Text: "Me veo como alguien que es relajado/a, maneja el estrés bien.", Dimension: "N", IsReverse: true, Required: true},
		{ID: 28, Text: "Me veo como alguien que puede ser tenso/a.", Dimension: "N", IsReverse: false, Required: true},
		{ID: 29, Text: "Me veo como alguien que se preocupa mucho.", Dimension: "N", IsReverse: false, Required: true},
		{ID: 30, Text: "Me veo como alguien que es emocionalmente estable y no se molesta fácilmente.", Dimension: "N", IsReverse: true, Required: true},
		{ID: 31, Text: "Me veo como alguien que puede ser malhumorado/a.", Dimension: "N", IsReverse: false, Required: true},
		{ID: 32, Text: "Me veo como alguien que sigue siendo calmado/a en situaciones tensas.", Dimension: "N", IsReverse: true, Required: true},
		{ID: 33, Text: "Me veo como alguien que se pone nervioso/a con facilidad.", Dimension: "N", IsReverse: false, Required: true},

		{ID: 34, Text: "Me veo como alguien que es original, que tiene nuevas ideas.", Dimension: "O", IsReverse: false, Required: true},
		{ID: 35, Text: "Me veo como alguien que es curioso/a de muchas cosas diferentes.", Dimension: "O", IsReverse: false, Required: true},
		{ID: 36, Text: "Me veo como alguien que es ingenioso/a, un/a pensador/a profundo/a.", Dimension: "O", IsReverse: false, Required: true},
		{ID: 37, Text: "Me veo como alguien que tiene una imaginación activa.", Dimension: "O", IsReverse: false, Required: true},
		{ID: 38, Text: "Me veo como alguien que valora las experiencias artísticas y estéticas.", Dimension: "O", IsReverse: false, Required: true},
		{ID: 39, Text: "Me veo como alguien que prefiere el trabajo rutinario.", Dimension: "O", IsReverse: true, Required: true},
		{ID: 40, Text: "Me veo como alguien que le gusta reflexionar y jugar con las ideas.", Dimension: "O", IsReverse: false, Required: true},
		{ID: 41, Text: "Me veo como alguien que tiene poca interés en las artes.", Dimension: "O", IsReverse: true, Required: true},
		{ID: 44, Text: "Me veo como alguien sofisticado/a en arte, música o literatura.", Dimension: "O", IsReverse: false, Required: true},
	}
}

func (p *BFI44Plugin) Calculate(inputs map[int]int) (BFI44Results, error) {
	questions := getBFI44Questions()

	correctedScores, err := applyReverseScoring(questions, inputs)
	if err != nil {
		return BFI44Results{}, err
	}

	dimensionScores, missingItems, err := groupByDimension(questions, correctedScores)
	if err != nil {
		return BFI44Results{}, err
	}

	imputedItems, correctedScores := imputeMissingItems(dimensionScores, missingItems, correctedScores)

	results := BFI44Results{
		CorrectedScores: correctedScores,
		MissingItems:    missingItems,
		ImputedItems:    imputedItems,
	}

	results.Extraversion = dimensionAverage(dimensionScores["E"])
	results.Agreeableness = dimensionAverage(dimensionScores["A"])
	results.Conscientiousness = dimensionAverage(dimensionScores["C"])
	results.Neuroticism = dimensionAverage(dimensionScores["N"])
	results.Openness = dimensionAverage(dimensionScores["O"])

	results.OCEANVector = OCEANVectorInline{
		Openness:          normalizeBFI(results.Openness),
		Conscientiousness: normalizeBFI(results.Conscientiousness),
		Extraversion:      normalizeBFI(results.Extraversion),
		Agreeableness:     normalizeBFI(results.Agreeableness),
		Neuroticism:       normalizeBFI(results.Neuroticism),
	}

	if err := validateOCEAN(results.OCEANVector); err != nil {
		return results, fmt.Errorf("invalid OCEAN vector: %v", err)
	}
	return results, nil
}

func applyReverseScoring(questions []BFI44Question, inputs map[int]int) (map[int]int, error) {
	corrected := make(map[int]int)
	for _, q := range questions {
		score, exists := inputs[q.ID]
		if !exists {
			continue
		}
		if score < 1 || score > 5 {
			return nil, fmt.Errorf("score for question %d must be between 1 and 5, got %d", q.ID, score)
		}
		if q.IsReverse {
			corrected[q.ID] = 6 - score
		} else {
			corrected[q.ID] = score
		}
	}
	return corrected, nil
}

func groupByDimension(questions []BFI44Question, correctedScores map[int]int) (map[string][]int, map[string][]int, error) {
	dimensionScores := map[string][]int{"E": {}, "A": {}, "C": {}, "N": {}, "O": {}}
	missingItems := make(map[string][]int)

	for _, q := range questions {
		if score, exists := correctedScores[q.ID]; exists {
			dimensionScores[q.Dimension] = append(dimensionScores[q.Dimension], score)
		} else {
			missingItems[q.Dimension] = append(missingItems[q.Dimension], q.ID)
		}
	}

	toleranceLimits := map[string]int{"E": 1, "A": 2, "C": 2, "N": 1, "O": 2}
	for dim, missing := range missingItems {
		if len(missing) > toleranceLimits[dim] {
			return nil, nil, fmt.Errorf("too many missing items in dimension %s: %d missing (max allowed: %d)", dim, len(missing), toleranceLimits[dim])
		}
	}

	return dimensionScores, missingItems, nil
}

func imputeMissingItems(dimensionScores, missingItems map[string][]int, correctedScores map[int]int) (map[string][]int, map[int]int) {
	imputedItems := make(map[string][]int)
	dimensionMeans := make(map[string]float64)
	for dim, scores := range dimensionScores {
		if len(scores) > 0 {
			dimensionMeans[dim] = dimensionAverage(scores)
		}
	}
	for dim, missing := range missingItems {
		if len(missing) > 0 && dimensionMeans[dim] > 0 {
			imputed := int(dimensionMeans[dim] + 0.5)
			for _, itemID := range missing {
				correctedScores[itemID] = imputed
				dimensionScores[dim] = append(dimensionScores[dim], imputed)
				imputedItems[dim] = append(imputedItems[dim], itemID)
			}
		}
	}
	return imputedItems, correctedScores
}

func dimensionAverage(scores []int) float64 {
	if len(scores) == 0 {
		return 0
	}
	sum := 0
	for _, s := range scores {
		sum += s
	}
	return float64(sum) / float64(len(scores))
}

func normalizeBFI(val float64) float64 {
	return (val - 1.0) / 4.0
}

func validateOCEAN(o OCEANVectorInline) error {
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
