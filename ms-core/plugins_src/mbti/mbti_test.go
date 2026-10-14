package main

import (
	"testing"
)

// ============================================================================
// TESTS DE CÁLCULO MBTI
// ============================================================================

// TestCalculateMBTI_TieBreaking_INFP verifica que un empate 5 vs 5 en todas las dimensiones
// produzca INFP (el polo B siempre gana en empate)
func TestCalculateMBTI_TieBreaking_INFP(t *testing.T) {
	// Generar respuestas con empate perfecto (5A, 5B) en cada dimensión
	answers := make([]MBTIAnswer, 0, 40)

	// Dimensión E-I (ítems 1-10): 5 responden A (E), 5 responden B (I)
	for i := 1; i <= 5; i++ {
		answers = append(answers, MBTIAnswer{QuestionID: i, Selection: "A"})
	}
	for i := 6; i <= 10; i++ {
		answers = append(answers, MBTIAnswer{QuestionID: i, Selection: "B"})
	}

	// Dimensión S-N (ítems 11-20): 5 responden A (S), 5 responden B (N)
	for i := 11; i <= 15; i++ {
		answers = append(answers, MBTIAnswer{QuestionID: i, Selection: "A"})
	}
	for i := 16; i <= 20; i++ {
		answers = append(answers, MBTIAnswer{QuestionID: i, Selection: "B"})
	}

	// Dimensión T-F (ítems 21-30): 5 responden A (T), 5 responden B (F)
	for i := 21; i <= 25; i++ {
		answers = append(answers, MBTIAnswer{QuestionID: i, Selection: "A"})
	}
	for i := 26; i <= 30; i++ {
		answers = append(answers, MBTIAnswer{QuestionID: i, Selection: "B"})
	}

	// Dimensión J-P (ítems 31-40): 5 responden A (J), 5 responden B (P)
	for i := 31; i <= 35; i++ {
		answers = append(answers, MBTIAnswer{QuestionID: i, Selection: "A"})
	}
	for i := 36; i <= 40; i++ {
		answers = append(answers, MBTIAnswer{QuestionID: i, Selection: "B"})
	}

	result := CalculateMBTI(answers)

	// Verificar tipo MBTI resultante
	expectedType := "INFP"
	if result.MBTIType != expectedType {
		t.Errorf("Expected MBTI type %s on tie-break, got %s", expectedType, result.MBTIType)
	}

	// Verificar que cada dimensión tiene el puntaje correcto
	for _, dim := range result.Dimensions {
		if dim.ScoreA != 5 || dim.ScoreB != 5 {
			t.Errorf("Dimension %s: expected scores 5/5, got %d/%d", dim.Name, dim.ScoreA, dim.ScoreB)
		}

		// PCI debe ser 0 en empate perfecto (|5-5|/(5+5)*100 = 0)
		if dim.PCI != 0 {
			t.Errorf("Dimension %s: expected PCI 0 on tie, got %.2f", dim.Name, dim.PCI)
		}
	}

	// Verificar preferencias individuales (todas deben ser polo B)
	expectedPreferences := map[MBTIDimension]string{
		DimensionEI: "I",
		DimensionSN: "N",
		DimensionTF: "F",
		DimensionJP: "P",
	}

	for _, dim := range result.Dimensions {
		expected := expectedPreferences[dim.Name]
		if dim.Preference != expected {
			t.Errorf("Dimension %s: expected preference %s on tie, got %s", dim.Name, expected, dim.Preference)
		}
	}
}

// TestCalculateMBTI_ClearPreference verifica preferencias claras (sin empate)
func TestCalculateMBTI_ClearPreference(t *testing.T) {
	// Generar respuestas todas A para obtener ESTJ
	answers := make([]MBTIAnswer, 0, 40)
	for i := 1; i <= 40; i++ {
		answers = append(answers, MBTIAnswer{QuestionID: i, Selection: "A"})
	}

	result := CalculateMBTI(answers)

	expectedType := "ESTJ"
	if result.MBTIType != expectedType {
		t.Errorf("Expected MBTI type %s, got %s", expectedType, result.MBTIType)
	}

	// Verificar puntajes (10-0 en cada dimensión)
	for _, dim := range result.Dimensions {
		if dim.ScoreA != 10 || dim.ScoreB != 0 {
			t.Errorf("Dimension %s: expected scores 10/0, got %d/%d", dim.Name, dim.ScoreA, dim.ScoreB)
		}

		// PCI debe ser 100 (máxima claridad)
		if dim.PCI != 100 {
			t.Errorf("Dimension %s: expected PCI 100, got %.2f", dim.Name, dim.PCI)
		}
	}
}

// TestCalculateMBTI_AllB verifica respuestas todas B para obtener INFP
func TestCalculateMBTI_AllB(t *testing.T) {
	answers := make([]MBTIAnswer, 0, 40)
	for i := 1; i <= 40; i++ {
		answers = append(answers, MBTIAnswer{QuestionID: i, Selection: "B"})
	}

	result := CalculateMBTI(answers)

	expectedType := "INFP"
	if result.MBTIType != expectedType {
		t.Errorf("Expected MBTI type %s, got %s", expectedType, result.MBTIType)
	}

	// Verificar puntajes (0-10 en cada dimensión)
	for _, dim := range result.Dimensions {
		if dim.ScoreA != 0 || dim.ScoreB != 10 {
			t.Errorf("Dimension %s: expected scores 0/10, got %d/%d", dim.Name, dim.ScoreA, dim.ScoreB)
		}
	}
}

// TestCalculateMBTI_MixedType verifica un tipo mixto (INTP)
func TestCalculateMBTI_MixedType(t *testing.T) {
	answers := make([]MBTIAnswer, 0, 40)

	// E-I: mayoría B (I) - 3A, 7B
	for i := 1; i <= 3; i++ {
		answers = append(answers, MBTIAnswer{QuestionID: i, Selection: "A"})
	}
	for i := 4; i <= 10; i++ {
		answers = append(answers, MBTIAnswer{QuestionID: i, Selection: "B"})
	}

	// S-N: mayoría B (N) - 2A, 8B
	for i := 11; i <= 12; i++ {
		answers = append(answers, MBTIAnswer{QuestionID: i, Selection: "A"})
	}
	for i := 13; i <= 20; i++ {
		answers = append(answers, MBTIAnswer{QuestionID: i, Selection: "B"})
	}

	// T-F: mayoría A (T) - 8A, 2B
	for i := 21; i <= 28; i++ {
		answers = append(answers, MBTIAnswer{QuestionID: i, Selection: "A"})
	}
	for i := 29; i <= 30; i++ {
		answers = append(answers, MBTIAnswer{QuestionID: i, Selection: "B"})
	}

	// J-P: mayoría B (P) - 4A, 6B
	for i := 31; i <= 34; i++ {
		answers = append(answers, MBTIAnswer{QuestionID: i, Selection: "A"})
	}
	for i := 35; i <= 40; i++ {
		answers = append(answers, MBTIAnswer{QuestionID: i, Selection: "B"})
	}

	result := CalculateMBTI(answers)

	expectedType := "INTP"
	if result.MBTIType != expectedType {
		t.Errorf("Expected MBTI type %s, got %s", expectedType, result.MBTIType)
	}

	// Verificar dimensión E-I
	eiDim := result.Dimensions[0]
	if eiDim.ScoreA != 3 || eiDim.ScoreB != 7 {
		t.Errorf("E-I: expected 3/7, got %d/%d", eiDim.ScoreA, eiDim.ScoreB)
	}
	if eiDim.Preference != "I" {
		t.Errorf("E-I: expected preference I, got %s", eiDim.Preference)
	}

	// Verificar PCI de E-I: |3-7|/(3+7)*100 = 40
	if eiDim.PCI != 40 {
		t.Errorf("E-I PCI: expected 40, got %.2f", eiDim.PCI)
	}
}

// TestCalculateMBTI_PCICalculation verifica el cálculo correcto del PCI
func TestCalculateMBTI_PCICalculation(t *testing.T) {
	testCases := []struct {
		scoreA      int
		scoreB      int
		expectedPCI float64
	}{
		{10, 0, 100}, // Máxima claridad
		{0, 10, 100}, // Máxima claridad (polo opuesto)
		{5, 5, 0},    // Empate = sin claridad
		{7, 3, 40},   // |7-3|/10*100 = 40
		{8, 2, 60},   // |8-2|/10*100 = 60
		{9, 1, 80},   // |9-1|/10*100 = 80
		{6, 4, 20},   // |6-4|/10*100 = 20
	}

	for _, tc := range testCases {
		answers := make([]MBTIAnswer, 0, 10)
		// Solo usamos la primera dimensión (E-I, ítems 1-10)
		for i := 1; i <= tc.scoreA; i++ {
			answers = append(answers, MBTIAnswer{QuestionID: i, Selection: "A"})
		}
		for i := tc.scoreA + 1; i <= 10; i++ {
			answers = append(answers, MBTIAnswer{QuestionID: i, Selection: "B"})
		}

		result := CalculateMBTI(answers)

		eiDim := result.Dimensions[0]
		if eiDim.PCI != tc.expectedPCI {
			t.Errorf("PCI for %d/%d: expected %.2f, got %.2f", tc.scoreA, tc.scoreB, tc.expectedPCI, eiDim.PCI)
		}
	}
}

// TestCalculateMBTI_EmptyAnswers verifica el manejo de respuestas vacías
func TestCalculateMBTI_EmptyAnswers(t *testing.T) {
	answers := []MBTIAnswer{}
	result := CalculateMBTI(answers)

	// Con 0 respuestas, todos los polos B ganan (por tie-breaking)
	expectedType := "INFP"
	if result.MBTIType != expectedType {
		t.Errorf("Expected MBTI type %s on empty answers, got %s", expectedType, result.MBTIType)
	}

	// PCI debe ser 0 en todas las dimensiones (sin respuestas)
	for _, dim := range result.Dimensions {
		if dim.PCI != 0 {
			t.Errorf("Dimension %s: expected PCI 0 on empty, got %.2f", dim.Name, dim.PCI)
		}
	}
}

// TestCalculateMBTI_PartialAnswers verifica respuestas parciales
func TestCalculateMBTI_PartialAnswers(t *testing.T) {
	// Solo responder las primeras 10 preguntas (dimensión E-I)
	answers := make([]MBTIAnswer, 0, 10)
	for i := 1; i <= 10; i++ {
		answers = append(answers, MBTIAnswer{QuestionID: i, Selection: "A"})
	}

	result := CalculateMBTI(answers)

	// E-I debe ser E (10A, 0B)
	if result.Dimensions[0].Preference != "E" {
		t.Errorf("E-I: expected E, got %s", result.Dimensions[0].Preference)
	}

	// Las otras dimensiones deben ser polo B por tie-breaking (0 vs 0)
	if result.Dimensions[1].Preference != "N" {
		t.Errorf("S-N: expected N on tie, got %s", result.Dimensions[1].Preference)
	}
	if result.Dimensions[2].Preference != "F" {
		t.Errorf("T-F: expected F on tie, got %s", result.Dimensions[2].Preference)
	}
	if result.Dimensions[3].Preference != "P" {
		t.Errorf("J-P: expected P on tie, got %s", result.Dimensions[3].Preference)
	}

	expectedType := "ENFP"
	if result.MBTIType != expectedType {
		t.Errorf("Expected MBTI type %s, got %s", expectedType, result.MBTIType)
	}
}

// TestCalculateMBTI_InvalidQuestionIDs verifica que IDs inválidos se ignoren
func TestCalculateMBTI_InvalidQuestionIDs(t *testing.T) {
	answers := []MBTIAnswer{
		{QuestionID: 0, Selection: "A"},   // Inválido
		{QuestionID: -1, Selection: "A"},  // Inválido
		{QuestionID: 41, Selection: "A"},  // Inválido
		{QuestionID: 100, Selection: "A"}, // Inválido
		{QuestionID: 1, Selection: "A"},   // Válido
	}

	result := CalculateMBTI(answers)

	// Solo 1 respuesta válida en E-I
	if result.Dimensions[0].ScoreA != 1 {
		t.Errorf("Expected ScoreA=1, got %d", result.Dimensions[0].ScoreA)
	}
}

// ============================================================================
// TESTS DE PREGUNTAS SEED
// ============================================================================

// TestMBTIQuestionCount verifica que haya exactamente 40 preguntas
func TestMBTIQuestionCount(t *testing.T) {
	questions := GetMBTIQuestions()

	if len(questions) != 40 {
		t.Errorf("Expected 40 questions, got %d", len(questions))
	}
}

// TestMBTIQuestionIDs verifica IDs únicos en rango 1-40
func TestMBTIQuestionIDs(t *testing.T) {
	questions := GetMBTIQuestions()

	ids := make(map[int]bool)
	for _, q := range questions {
		if q.ID < 1 || q.ID > 40 {
			t.Errorf("Question ID %d is out of range [1, 40]", q.ID)
		}
		if ids[q.ID] {
			t.Errorf("Duplicate question ID: %d", q.ID)
		}
		ids[q.ID] = true
	}
}

// TestMBTIDimensionDistribution verifica la distribución correcta de preguntas por dimensión
func TestMBTIDimensionDistribution(t *testing.T) {
	questions := GetMBTIQuestions()

	distribution := map[MBTIDimension]int{
		DimensionEI: 0,
		DimensionSN: 0,
		DimensionTF: 0,
		DimensionJP: 0,
	}

	for _, q := range questions {
		distribution[q.Dimension]++
	}

	// Cada dimensión debe tener exactamente 10 preguntas
	for dim, count := range distribution {
		if count != 10 {
			t.Errorf("Dimension %s: expected 10 questions, got %d", dim, count)
		}
	}
}

// TestMBTIQuestionDimensionMapping verifica el mapeo correcto de IDs a dimensiones
func TestMBTIQuestionDimensionMapping(t *testing.T) {
	questions := GetMBTIQuestions()

	for _, q := range questions {
		var expectedDim MBTIDimension
		switch {
		case q.ID >= 1 && q.ID <= 10:
			expectedDim = DimensionEI
		case q.ID >= 11 && q.ID <= 20:
			expectedDim = DimensionSN
		case q.ID >= 21 && q.ID <= 30:
			expectedDim = DimensionTF
		case q.ID >= 31 && q.ID <= 40:
			expectedDim = DimensionJP
		}

		if q.Dimension != expectedDim {
			t.Errorf("Question %d: expected dimension %s, got %s", q.ID, expectedDim, q.Dimension)
		}
	}
}

// TestMBTIQuestionValues verifica los valores correctos de los polos
func TestMBTIQuestionValues(t *testing.T) {
	questions := GetMBTIQuestions()

	expectedValues := map[MBTIDimension]struct{ A, B string }{
		DimensionEI: {"E", "I"},
		DimensionSN: {"S", "N"},
		DimensionTF: {"T", "F"},
		DimensionJP: {"J", "P"},
	}

	for _, q := range questions {
		expected := expectedValues[q.Dimension]
		if q.ValueA != expected.A {
			t.Errorf("Question %d: expected ValueA=%s, got %s", q.ID, expected.A, q.ValueA)
		}
		if q.ValueB != expected.B {
			t.Errorf("Question %d: expected ValueB=%s, got %s", q.ID, expected.B, q.ValueB)
		}
	}
}

// TestMBTIQuestionsHaveContent verifica que todas las preguntas tengan contenido
func TestMBTIQuestionsHaveContent(t *testing.T) {
	questions := GetMBTIQuestions()

	for _, q := range questions {
		if q.Statement == "" {
			t.Errorf("Question %d: Statement is empty", q.ID)
		}
		if q.OptionA == "" {
			t.Errorf("Question %d: OptionA is empty", q.ID)
		}
		if q.OptionB == "" {
			t.Errorf("Question %d: OptionB is empty", q.ID)
		}
	}
}

// ============================================================================
// TESTS DE COMPATIBILIDAD MBTI
// ============================================================================

// TestMBTICompatibility_SameType verifica compatibilidad consigo mismo
func TestMBTICompatibility_SameType(t *testing.T) {
	plugin := NewMBTIPlugin()

	types := []string{"INTJ", "ENFP", "ISTP", "ESFJ"}
	for _, mbtiType := range types {
		score := plugin.GetMBTICompatibility(mbtiType, mbtiType)
		if score < 0.5 {
			t.Errorf("Same type %s compatibility should be >= 0.5, got %.2f", mbtiType, score)
		}
	}
}

// TestMBTICompatibility_InvalidType verifica manejo de tipos inválidos
func TestMBTICompatibility_InvalidType(t *testing.T) {
	plugin := NewMBTIPlugin()

	testCases := []struct {
		typeA string
		typeB string
	}{
		{"", "INTJ"},
		{"INTJ", ""},
		{"INT", "ENFP"},
		{"INTJX", "ENFP"},
	}

	for _, tc := range testCases {
		score := plugin.GetMBTICompatibility(tc.typeA, tc.typeB)
		if score != 0.5 {
			t.Errorf("Invalid types %s/%s should return 0.5, got %.2f", tc.typeA, tc.typeB, score)
		}
	}
}

// TestMBTICompatibility_ComplementaryTypes verifica compatibilidad de tipos complementarios
func TestMBTICompatibility_ComplementaryTypes(t *testing.T) {
	plugin := NewMBTIPlugin()

	pairs := []struct {
		typeA string
		typeB string
	}{
		{"INTJ", "ENTP"},
		{"INFJ", "ENFP"},
		{"INTP", "ENTJ"},
	}

	for _, pair := range pairs {
		score := plugin.GetMBTICompatibility(pair.typeA, pair.typeB)
		if score < 0.5 {
			t.Errorf("Complementary types %s/%s should have score >= 0.5, got %.2f",
				pair.typeA, pair.typeB, score)
		}
	}
}

// ============================================================================
// TESTS DE FUNCIONES HELPER
// ============================================================================

// TestCalculateMBTIFromMap verifica la función de cálculo desde mapa
func TestCalculateMBTIFromMap(t *testing.T) {
	answersMap := map[int]string{}

	// Todas A para las primeras 20, todas B para las últimas 20
	for i := 1; i <= 20; i++ {
		answersMap[i] = "A"
	}
	for i := 21; i <= 40; i++ {
		answersMap[i] = "B"
	}

	result := CalculateMBTIFromMap(answersMap)

	expectedType := "ESFP"
	if result.MBTIType != expectedType {
		t.Errorf("Expected %s, got %s", expectedType, result.MBTIType)
	}
}

// ============================================================================
// BENCHMARK TESTS
// ============================================================================

// BenchmarkCalculateMBTI mide el rendimiento del cálculo MBTI
func BenchmarkCalculateMBTI(b *testing.B) {
	answers := make([]MBTIAnswer, 40)
	for i := 0; i < 40; i++ {
		selection := "A"
		if i%2 == 0 {
			selection = "B"
		}
		answers[i] = MBTIAnswer{QuestionID: i + 1, Selection: selection}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateMBTI(answers)
	}
}
