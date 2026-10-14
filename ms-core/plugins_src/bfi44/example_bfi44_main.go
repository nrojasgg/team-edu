//go:build ignore

package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("=== BFI-44 Plugin Example ===\n")

	plugin := NewBFI44Plugin()

	fmt.Printf("Plugin: %s v%s\n\n", plugin.Name(), plugin.Version())

	fmt.Println("--- Ejemplo 1: Respuestas Completas ---")
	completeInputs := generateCompleteResponses()

	results1, err := plugin.Calculate(completeInputs)
	if err != nil {
		log.Fatalf("Error calculating complete responses: %v", err)
	}

	printResults("Respuestas Completas", results1)

	fmt.Println("\n--- Ejemplo 2: Con Missing Data (Dentro del Límite) ---")
	partialInputs := generatePartialResponses()

	results2, err := plugin.Calculate(partialInputs)
	if err != nil {
		log.Fatalf("Error calculating partial responses: %v", err)
	}

	printResults("Con Missing Data", results2)

	fmt.Println("\n--- Ejemplo 3: Demostración de Reverse Scoring ---")
	demonstrateReverseScoring()

	fmt.Println("\n--- Ejemplo 4: Exceso de Missing Data (Debe Fallar) ---")
	excessiveMissingInputs := generateExcessiveMissingData()

	_, err = plugin.Calculate(excessiveMissingInputs)
	if err != nil {
		fmt.Printf("Error esperado: %v\n", err)
	} else {
		fmt.Println("ERROR: Debería haber fallado por exceso de missing data")
	}

	fmt.Println("\n--- Ejemplo 5: Uso a través de Execute ---")
	demonstrateExecute(plugin)
}

func generateCompleteResponses() map[int]int {
	inputs := make(map[int]int)

	inputs[1] = 4
	inputs[6] = 2
	inputs[11] = 5
	inputs[16] = 4
	inputs[21] = 2
	inputs[26] = 4
	inputs[31] = 2
	inputs[36] = 5

	inputs[2] = 2
	inputs[7] = 4
	inputs[12] = 2
	inputs[17] = 4
	inputs[22] = 4
	inputs[27] = 2
	inputs[32] = 5
	inputs[37] = 1
	inputs[42] = 4

	inputs[3] = 5
	inputs[8] = 2
	inputs[13] = 5
	inputs[18] = 2
	inputs[23] = 1
	inputs[28] = 5
	inputs[33] = 5
	inputs[38] = 4
	inputs[43] = 2

	inputs[4] = 2
	inputs[9] = 4
	inputs[14] = 2
	inputs[19] = 3
	inputs[24] = 4
	inputs[29] = 2
	inputs[34] = 4
	inputs[39] = 2

	inputs[5] = 5
	inputs[10] = 5
	inputs[15] = 4
	inputs[20] = 5
	inputs[25] = 4
	inputs[30] = 4
	inputs[35] = 2
	inputs[40] = 5
	inputs[41] = 2
	inputs[44] = 4

	return inputs
}

func generatePartialResponses() map[int]int {
	inputs := generateCompleteResponses()

	delete(inputs, 31)
	delete(inputs, 12)
	delete(inputs, 37)
	delete(inputs, 43)
	delete(inputs, 39)
	delete(inputs, 35)
	delete(inputs, 41)

	return inputs
}

func generateExcessiveMissingData() map[int]int {
	inputs := generateCompleteResponses()

	delete(inputs, 6)
	delete(inputs, 21)
	delete(inputs, 31)

	return inputs
}

func demonstrateReverseScoring() {
	plugin := NewBFI44Plugin()

	inputs := map[int]int{
		6:  5,
		1:  5,
		11: 4,
		16: 4,
		21: 2,
		26: 4,
		31: 2,
		36: 5,
	}

	for i := 2; i <= 44; i++ {
		if _, exists := inputs[i]; !exists {
			inputs[i] = 3
		}
	}

	results, _ := plugin.Calculate(inputs)

	fmt.Printf("Item 6 (inverso): Usuario respondio 5 -> Corregido a %d\n", results.CorrectedScores[6])
	fmt.Printf("Item 1 (directo): Usuario respondio 5 -> Corregido a %d\n", results.CorrectedScores[1])
	fmt.Printf("Extraversion final: %.2f\n", results.Extraversion)
}

func demonstrateExecute(plugin *BFI44Plugin) {
	answers := []map[string]interface{}{}
	inputs := generateCompleteResponses()

	for qid, score := range inputs {
		answers = append(answers, map[string]interface{}{
			"question_id": float64(qid),
			"score":       float64(score),
		})
	}

	request := map[string]interface{}{
		"action":     "submit_response",
		"test_id":    "507f1f77bcf86cd799439011",
		"student_id": "12345",
		"answers":    answers,
	}

	fmt.Println("Estructura de request para Execute:")
	fmt.Printf("  Action: %s\n", request["action"])
	fmt.Printf("  Test ID: %s\n", request["test_id"])
	fmt.Printf("  Student ID: %s\n", request["student_id"])
	fmt.Printf("  Numero de respuestas: %d\n", len(answers))
	fmt.Println("\n(Nota: Requiere MongoDB para ejecutar completamente)")
}

func printResults(title string, results BFI44Results) {
	fmt.Printf("\n%s:\n", title)
	fmt.Println("----------------------------------------")

	fmt.Println("Puntuaciones Raw (escala 1-5):")
	fmt.Printf("  Extraversion (E):      %.2f\n", results.Extraversion)
	fmt.Printf("  Amabilidad (A):        %.2f\n", results.Agreeableness)
	fmt.Printf("  Responsabilidad (C):   %.2f\n", results.Conscientiousness)
	fmt.Printf("  Neuroticismo (N):      %.2f\n", results.Neuroticism)
	fmt.Printf("  Apertura (O):          %.2f\n", results.Openness)

	fmt.Println("\nOCEAN Vector Normalizado (escala 0-1):")
	fmt.Printf("  Openness:              %.3f\n", results.OCEANVector.Openness)
	fmt.Printf("  Conscientiousness:     %.3f\n", results.OCEANVector.Conscientiousness)
	fmt.Printf("  Extraversion:          %.3f\n", results.OCEANVector.Extraversion)
	fmt.Printf("  Agreeableness:         %.3f\n", results.OCEANVector.Agreeableness)
	fmt.Printf("  Neuroticism:           %.3f\n", results.OCEANVector.Neuroticism)

	if len(results.MissingItems) > 0 {
		fmt.Println("\nItems Faltantes:")
		for dim, items := range results.MissingItems {
			fmt.Printf("  Dimension %s: %v\n", dim, items)
		}
	}

	if len(results.ImputedItems) > 0 {
		fmt.Println("\nItems Imputados (media intra-sujeto):")
		for dim, items := range results.ImputedItems {
			fmt.Printf("  Dimension %s: %v\n", dim, items)
		}
	}

	fmt.Println("\nEjemplos de Reverse Scoring:")
	reverseItems := []int{6, 21, 31, 2, 12, 27, 37, 8, 18, 23, 43, 9, 24, 34, 35, 41}
	count := 0
	for _, itemID := range reverseItems {
		if score, exists := results.CorrectedScores[itemID]; exists && count < 3 {
			fmt.Printf("  Item %d (inverso): Score corregido = %d\n", itemID, score)
			count++
		}
	}

	fmt.Println("----------------------------------------")
}
