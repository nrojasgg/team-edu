package adapters

import (
	"context"
	"fmt"

	"ms-core/internal/core"
)

// SociogramAffinityProvider adapta un core.Plugin de tipo sociograma a la interfaz
// AffinityProvider. Trabaja exclusivamente contra core.Plugin para no depender del
// paquete plugins; las respuestas del plugin se parsean como map[string]interface{}.
type SociogramAffinityProvider struct {
	plugin core.Plugin
	// cache de matrices de adyacencia por clave "sectionID:surveyID"
	cache map[string]*SociogramMatrix
}

// SociogramMatrix representa la matriz de adyacencia social entre estudiantes.
type SociogramMatrix struct {
	// Choices[a][b] == true → a eligió a b para trabajar
	Choices map[string]map[string]bool
	// Rejections[a][b] == true → a rechazó a b
	Rejections map[string]map[string]bool
	// PerceivedChoices[a][b] == true → a cree que b lo elegiría
	PerceivedChoices map[string]map[string]bool
	// PerceivedRejections[a][b] == true → a cree que b lo rechazaría
	PerceivedRejections map[string]map[string]bool
}

// NewSociogramAffinityProvider crea un nuevo adaptador para el plugin de sociograma.
func NewSociogramAffinityProvider(plugin core.Plugin) *SociogramAffinityProvider {
	return &SociogramAffinityProvider{
		plugin: plugin,
		cache:  make(map[string]*SociogramMatrix),
	}
}

// ProviderID implementa AffinityProvider.
func (p *SociogramAffinityProvider) ProviderID() string {
	return "sociogram"
}

// ProviderName implementa AffinityProvider.
func (p *SociogramAffinityProvider) ProviderName() string {
	return "Social Network Affinity (Sociogram)"
}

// CalculatePairwiseAffinity calcula la afinidad entre dos personas basándose en
// las elecciones y rechazos del cuestionario sociométrico.
//
// Algoritmo:
//   - Elección mutua:         +1.0
//   - Elección unilateral:    +0.5
//   - Rechazo mutuo:          -1.0
//   - Rechazo unilateral:     -0.5
//   - Conflicto (uno elige, otro rechaza): -0.3 adicional
//   - Percepción mutua de elección: +0.2
//   - Percepción mutua de rechazo:  -0.2
//
// El resultado final se normaliza al rango [-1.0, 1.0].
func (p *SociogramAffinityProvider) CalculatePairwiseAffinity(ctx context.Context, personA, personB string) (float64, error) {
	sectionID, ok := ctx.Value(core.ContextKeySectionID).(string)
	if !ok || sectionID == "" {
		return 0, fmt.Errorf("sociogram_adapter: section_id required in context")
	}

	surveyID, _ := ctx.Value(core.ContextKeySurveyID).(string)

	matrix, err := p.getOrBuildMatrix(ctx, sectionID, surveyID)
	if err != nil {
		return 0, err
	}
	if matrix == nil {
		return 0, nil
	}

	return p.calculateRelationshipScore(matrix, personA, personB), nil
}

// SupportsContext implementa AffinityProvider.
func (p *SociogramAffinityProvider) SupportsContext(ctx context.Context) bool {
	sectionID, ok := ctx.Value(core.ContextKeySectionID).(string)
	return ok && sectionID != ""
}

// GetMetadata implementa AffinityProvider.
func (p *SociogramAffinityProvider) GetMetadata() core.ProviderMetadata {
	return core.ProviderMetadata{
		ID:          "sociogram",
		Name:        "Social Network Affinity (Sociogram)",
		Description: "Calcula afinidad basándose en las elecciones y rechazos expresados por los estudiantes en el cuestionario sociométrico",
		Version:     "1.0.0",
		Category:    "social",
	}
}

// InvalidateCache invalida el cache para todas las entradas de una sección.
func (p *SociogramAffinityProvider) InvalidateCache(sectionID string) {
	for key := range p.cache {
		if len(key) >= len(sectionID) && key[:len(sectionID)] == sectionID {
			delete(p.cache, key)
		}
	}
}

// ---------------------------------------------------------------------------
// lógica de puntuación
// ---------------------------------------------------------------------------

// pairChoiceScore returns the score contribution and whether it applies,
// based on whether personA chose personB and vice versa.
func pairChoiceScore(aChoseB, bChoseA bool) (float64, bool) {
	switch {
	case aChoseB && bChoseA:
		return 1.0, true
	case aChoseB || bChoseA:
		return 0.5, true
	}
	return 0, false
}

// pairRejectionScore returns the score contribution and whether it applies,
// based on whether personA rejected personB and vice versa.
func pairRejectionScore(aRejectedB, bRejectedA bool) (float64, bool) {
	switch {
	case aRejectedB && bRejectedA:
		return -1.0, true
	case aRejectedB || bRejectedA:
		return -0.5, true
	}
	return 0, false
}

// clampNorm clamps a normalized score to the range [-1.0, 1.0].
func clampNorm(v float64) float64 {
	if v > 1.0 {
		return 1.0
	}
	if v < -1.0 {
		return -1.0
	}
	return v
}

func (p *SociogramAffinityProvider) calculateRelationshipScore(matrix *SociogramMatrix, personA, personB string) float64 {
	score := 0.0
	factors := 0

	aChoseB := matrix.Choices[personA][personB]
	bChoseA := matrix.Choices[personB][personA]
	aRejectedB := matrix.Rejections[personA][personB]
	bRejectedA := matrix.Rejections[personB][personA]

	if delta, ok := pairChoiceScore(aChoseB, bChoseA); ok {
		score += delta
		factors++
	}
	if delta, ok := pairRejectionScore(aRejectedB, bRejectedA); ok {
		score += delta
		factors++
	}

	// Situación conflictiva: uno elige y el otro rechaza
	if (aChoseB && bRejectedA) || (bChoseA && aRejectedB) {
		score -= 0.3
		factors++
	}

	// Percepciones mutuas de elección
	if matrix.PerceivedChoices[personA][personB] && matrix.PerceivedChoices[personB][personA] {
		score += 0.2
		factors++
	}

	// Percepciones mutuas de rechazo
	if matrix.PerceivedRejections[personA][personB] && matrix.PerceivedRejections[personB][personA] {
		score -= 0.2
		factors++
	}

	if factors == 0 {
		return 0.0
	}
	return clampNorm(score / float64(factors))
}

// ---------------------------------------------------------------------------
// construcción de la matriz desde el plugin
// ---------------------------------------------------------------------------

func (p *SociogramAffinityProvider) getOrBuildMatrix(ctx context.Context, sectionID, surveyID string) (*SociogramMatrix, error) {
	cacheKey := fmt.Sprintf("%s:%s", sectionID, surveyID)
	if matrix, exists := p.cache[cacheKey]; exists {
		return matrix, nil
	}

	matrix, err := p.buildMatrix(ctx, sectionID, surveyID)
	if err != nil {
		return nil, err
	}
	if matrix != nil {
		p.cache[cacheKey] = matrix
	}
	return matrix, nil
}

func (p *SociogramAffinityProvider) buildMatrix(ctx context.Context, sectionID, surveyID string) (*SociogramMatrix, error) {
	// 1. Obtener encuestas de la sección
	surveysRaw, err := p.plugin.Execute(ctx, map[string]interface{}{
		"action":     "get_surveys",
		"section_id": sectionID,
	})
	if err != nil {
		return nil, fmt.Errorf("sociogram_adapter: failed to get surveys: %w", err)
	}

	surveys, ok := toSliceOfMaps(surveysRaw)
	if !ok || len(surveys) == 0 {
		return nil, nil
	}

	// 2. Seleccionar la encuesta objetivo
	targetSurveyID := p.selectSurveyID(surveys, surveyID)
	if targetSurveyID == "" {
		return nil, nil
	}

	// 3. Obtener respuestas de la encuesta seleccionada
	responsesRaw, err := p.plugin.Execute(ctx, map[string]interface{}{
		"action":     "get_responses",
		"survey_id":  targetSurveyID,
		"section_id": sectionID,
	})
	if err != nil {
		return nil, fmt.Errorf("sociogram_adapter: failed to get responses: %w", err)
	}

	responses, ok := toSliceOfMaps(responsesRaw)
	if !ok || len(responses) == 0 {
		return nil, nil
	}

	// 4. Construir la matriz
	matrix := &SociogramMatrix{
		Choices:             make(map[string]map[string]bool),
		Rejections:          make(map[string]map[string]bool),
		PerceivedChoices:    make(map[string]map[string]bool),
		PerceivedRejections: make(map[string]map[string]bool),
	}
	fillMatrixFromResponses(matrix, responses)
	return matrix, nil
}

// fillMatrixFromResponses populates the social choice/rejection maps from raw plugin responses.
func fillMatrixFromResponses(matrix *SociogramMatrix, responses []map[string]interface{}) {
	for _, response := range responses {
		studentID, _ := response["student_id"].(string)
		if studentID == "" {
			continue
		}

		ensureInnerMap(matrix.Choices, studentID)
		ensureInnerMap(matrix.Rejections, studentID)
		ensureInnerMap(matrix.PerceivedChoices, studentID)
		ensureInnerMap(matrix.PerceivedRejections, studentID)

		// question1 → elecciones directas (response == 1)
		fillBoolMap(matrix.Choices[studentID], response["question1"])
		// question2 → percepciones de ser elegido
		fillBoolMap(matrix.PerceivedChoices[studentID], response["question2"])
		// question3 → rechazos directos
		fillBoolMap(matrix.Rejections[studentID], response["question3"])
		// question4 → percepciones de ser rechazado
		fillBoolMap(matrix.PerceivedRejections[studentID], response["question4"])
	}
}

// selectSurveyID devuelve el ID de la encuesta a usar: primero busca la que coincide
// con surveyID; si no hay coincidencia, usa la primera encuesta activa.
func (p *SociogramAffinityProvider) selectSurveyID(surveys []map[string]interface{}, preferredID string) string {
	for _, s := range surveys {
		id := mapStringVal(s, "id")
		if preferredID != "" && id == preferredID {
			return id
		}
	}
	// fallback: primera activa
	for _, s := range surveys {
		if isActive, _ := s["is_active"].(bool); isActive {
			if id := mapStringVal(s, "id"); id != "" {
				return id
			}
		}
	}
	return ""
}

// ---------------------------------------------------------------------------
// utilidades de conversión genérica
// ---------------------------------------------------------------------------

// toSliceOfMaps intenta convertir un valor genérico a []map[string]interface{}.
// Acepta tanto []map[string]interface{} directo como []interface{} cuyos elementos
// sean map[string]interface{}.
func toSliceOfMaps(raw interface{}) ([]map[string]interface{}, bool) {
	if raw == nil {
		return nil, false
	}

	// Caso 1: tipo nativo exacto
	if typed, ok := raw.([]map[string]interface{}); ok {
		return typed, true
	}

	// Caso 2: []interface{} con elementos map[string]interface{}
	if slice, ok := raw.([]interface{}); ok {
		result := make([]map[string]interface{}, 0, len(slice))
		for _, item := range slice {
			if m, ok := item.(map[string]interface{}); ok {
				result = append(result, m)
			}
		}
		return result, len(result) > 0
	}

	return nil, false
}

// isResponseOne reports whether v represents the integer value 1.
// It handles int, float64, and int64 as produced by JSON/BSON decoders.
func isResponseOne(v interface{}) bool {
	switch val := v.(type) {
	case int:
		return val == 1
	case float64:
		return val == 1
	case int64:
		return val == 1
	}
	return false
}

// toInterfaceSlice converts []map[string]interface{} to []interface{}.
func toInterfaceSlice(src []map[string]interface{}) []interface{} {
	out := make([]interface{}, len(src))
	for i, m := range src {
		out[i] = m
	}
	return out
}

// fillBoolMap rellena dst con las entradas de answers cuyo campo "response" sea 1.
// answers puede ser []interface{} o []map[string]interface{}.
func fillBoolMap(dst map[string]bool, answers interface{}) {
	if answers == nil {
		return
	}

	var items []interface{}
	switch v := answers.(type) {
	case []interface{}:
		items = v
	case []map[string]interface{}:
		items = toInterfaceSlice(v)
	default:
		return
	}

	for _, raw := range items {
		item, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		targetID, _ := item["student_id"].(string)
		if targetID == "" {
			continue
		}
		// response == 1 significa elección/rechazo positivo
		if isResponseOne(item["response"]) {
			dst[targetID] = true
		}
	}
}

// ensureInnerMap garantiza que la clave existe en el mapa externo.
func ensureInnerMap(outer map[string]map[string]bool, key string) {
	if outer[key] == nil {
		outer[key] = make(map[string]bool)
	}
}

// mapStringVal extrae un valor string de un mapa por clave.
func mapStringVal(m map[string]interface{}, key string) string {
	v, _ := m[key].(string)
	return v
}
