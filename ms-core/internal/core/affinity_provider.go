package core

import (
	"context"

	"ms-core/internal/core/domain"
)

// OCEANVector is an alias for domain.OCEANVector for convenience
type OCEANVector = domain.OCEANVector

// AffinityProvider define el contrato que cualquier fuente de afinidad debe implementar.
// El Core es agnóstico al origen de los datos (eneagrama, sociograma, etc.).
// Solo le interesa la pregunta: "¿Qué tan compatible es A con B?"
type AffinityProvider interface {
	// ProviderID retorna un identificador único del proveedor (ej. "enneagram", "sociogram")
	ProviderID() string

	// ProviderName retorna un nombre legible para el proveedor
	ProviderName() string

	// CalculatePairwiseAffinity calcula la afinidad entre dos personas.
	// Retorna un valor normalizado entre -1.0 y 1.0:
	//   - 1.0  = Máxima compatibilidad/afinidad
	//   - 0.0  = Neutral/Sin datos
	//   - -1.0 = Máximo rechazo/incompatibilidad
	//
	// El parámetro context puede contener información adicional como section_id, survey_id, etc.
	CalculatePairwiseAffinity(ctx context.Context, personA, personB string) (float64, error)

	// SupportsContext indica si el provider puede calcular afinidad para el contexto dado.
	// Esto permite al Core preguntar "¿Tienes datos para esta sección/grupo?"
	SupportsContext(ctx context.Context) bool

	// GetMetadata retorna información sobre el proveedor
	GetMetadata() ProviderMetadata
}

// ProviderMetadata contiene información descriptiva del proveedor
type ProviderMetadata struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	Description          string `json:"description"`
	Version              string `json:"version"`
	Category             string `json:"category"`              // "personality", "social", "performance", etc.
	SocialIntentionality int    `json:"social_intentionality"` // S_p: 1 for harmony, -1 for controlled friction
}

// ContextKey tipo para claves de contexto
type ContextKey string

const (
	// ContextKeySectionID clave para el ID de sección en el contexto
	ContextKeySectionID ContextKey = "section_id"
	// ContextKeySurveyID clave para el ID de encuesta en el contexto
	ContextKeySurveyID ContextKey = "survey_id"
	// ContextKeyGroupID clave para el ID de grupo en el contexto
	ContextKeyGroupID ContextKey = "group_id"
)

// PairwiseScore representa el score de afinidad entre dos personas
type PairwiseScore struct {
	PersonA    string  `json:"person_a"`
	PersonB    string  `json:"person_b"`
	Score      float64 `json:"score"` // -1.0 a 1.0
	ProviderID string  `json:"provider_id"`
	Confidence float64 `json:"confidence"` // 0.0 a 1.0 - qué tan seguro está el provider
}

// FitnessResult represents the multi-objective fitness evaluation of a solution
// Separates psychometric (OCEAN-based) and sociometric (Moreno-based) objectives
type FitnessResult struct {
	// PsychometricScore: OCEAN personality-based fitness [0.0, 1.0]
	// Measures team compatibility based on Big Five personality traits
	PsychometricScore float64 `json:"psychometric_score"`

	// SociometricScore: Moreno sociometry-based fitness [-∞, 1.0]
	// Measures team cohesion based on explicit student preferences/votes
	// Score < -500.0 indicates veto violation (infeasible solution)
	SociometricScore float64 `json:"sociometric_score"`

	// IsFeasible: False if solution violates hard constraints (vetoes)
	// Infeasible solutions are automatically rejected during selection
	IsFeasible bool `json:"is_feasible"`
}

// CombinedScore calculates the weighted sum of psychometric and sociometric scores
// Returns 0.0 if solution is infeasible
func (f FitnessResult) CombinedScore(psychometricWeight, sociometricWeight float64) float64 {
	if !f.IsFeasible {
		return -1000.0 // Infeasible solutions get worst possible score
	}

	// Normalize weights
	totalWeight := psychometricWeight + sociometricWeight
	if totalWeight == 0 {
		return 0.0
	}

	// Weighted sum
	return (f.PsychometricScore*psychometricWeight + f.SociometricScore*sociometricWeight) / totalWeight
}

// WeightedProvider envuelve un AffinityProvider con su peso configurado
type WeightedProvider struct {
	Provider AffinityProvider
	Weight   float64 // Peso relativo (ej. 0.7 para 70%)
}

// TeamMember representa a un miembro candidato para formar equipos
type TeamMember struct {
	ID string `json:"id"`

	// Psychometrics: The "DNA" of the student - OCEAN personality vector
	// This is the normalized multidimensional representation of personality
	Psychometrics *OCEANVector `json:"psychometrics,omitempty" bson:"psychometrics,omitempty"`

	// RawType: Original personality type metadata (e.g., {"enneagram": 5, "mbti": "INTJ"})
	// Preserves the discrete type information for backward compatibility and reference
	RawType map[string]interface{} `json:"raw_type,omitempty" bson:"raw_type,omitempty"`

	// Sociometry: Direct preferences from sociogram (Moreno method)
	// Key: student_id, Value: preference score (-1.0 to 1.0)
	// Positive values indicate preference, negative indicate rejection
	Sociometry map[string]float64 `json:"sociometry,omitempty" bson:"sociometry,omitempty"`

	// Attributes: Legacy field for backward compatibility
	// Contains additional metadata that doesn't fit in structured fields
	Attributes map[string]interface{} `json:"attributes,omitempty" bson:"attributes,omitempty"`
}

// Team representa un equipo formado
type Team struct {
	ID            string       `json:"id"`
	Members       []TeamMember `json:"members"`
	AffinityScore float64      `json:"affinity_score"` // Score combinado del equipo
	Metadata      TeamMetadata `json:"metadata"`
}

// CriterionBreakdownEntry holds the per-attribute detail for a single criterion
// evaluated on a specific group. Stored in MongoDB and returned to the frontend
// so the UI can render score / real-value / target / weight for each criterion.
type CriterionBreakdownEntry struct {
	// Score is the fitness value computed by the criterion's strategy.
	Score float64 `json:"score" bson:"score"`
	// RealValue is the actual normalised metric for the group (mean, variance, Blau index…).
	RealValue float64 `json:"real_value" bson:"real_value"`
	// Target is the desired value configured by the professor.
	Target float64 `json:"target" bson:"target"`
	// Weight is the normalised weight used in the fitness calculation.
	Weight float64 `json:"weight" bson:"weight"`
	// Goal is the evaluation strategy used for this criterion
	// ("minimize","maximize","target","homogeneity","heterogeneity").
	// Empty string means the default "target" strategy was applied.
	Goal string `json:"goal,omitempty" bson:"goal,omitempty"`
}

// TeamMetadata contiene información detallada sobre la formación del equipo
type TeamMetadata struct {
	ProviderScores    map[string]float64 `json:"provider_scores"`   // Score por cada provider
	TotalComparisons  int                `json:"total_comparisons"` // Número de pares evaluados
	AverageConfidence float64            `json:"average_confidence"`

	// Multi-objective fitness scores
	PsychometricScore float64 `json:"psychometric_score,omitempty"` // OCEAN-based fitness
	SociometricScore  float64 `json:"sociometric_score,omitempty"`  // Moreno-based fitness
	IsFeasible        bool    `json:"is_feasible"`                  // No veto violations

	// CriterionBreakdown stores the detailed per-criterion analysis used to compute
	// the group fitness. Each key matches a criterion name (e.g. "sex", "career_code").
	// Populated only in "criteria" mode (dynamic genetic algorithm).
	CriterionBreakdown map[string]CriterionBreakdownEntry `json:"criterion_breakdown,omitempty" bson:"criterion_breakdown,omitempty"`
}

// TeamFormationRequest representa una solicitud de formación de equipos
type TeamFormationRequest struct {
	SectionID   string       `json:"section_id"`
	Candidates  []TeamMember `json:"candidates"`
	TeamSize    int          `json:"team_size"`     // Tamaño deseado de cada equipo
	MinTeamSize int          `json:"min_team_size"` // Tamaño mínimo aceptable
	MaxTeamSize int          `json:"max_team_size"` // Tamaño máximo aceptable

	// Campos opcionales para configuración flexible por request
	Algorithm       string             `json:"algorithm,omitempty"`        // Algoritmo a usar (override)
	ProviderWeights map[string]float64 `json:"provider_weights,omitempty"` // Pesos personalizados por request

	// Multi-objective fitness weights (for genetic algorithm)
	// Default: PsychometricWeight=0.6, SociometricWeight=0.4
	PsychometricWeight float64 `json:"psychometric_weight,omitempty"` // Weight for OCEAN-based fitness
	SociometricWeight  float64 `json:"sociometric_weight,omitempty"`  // Weight for Moreno-based fitness
}

// CriterionConfig representa la configuración completa de un criterio de evaluación.
// Se usa tanto en la request de formación como en la respuesta, para que el frontend
// pueda reutilizar exactamente los mismos parámetros al llamar a evaluate-groups.
type CriterionConfig struct {
	Target float64 `json:"target"`           // Valor objetivo (0.0–1.0); siempre serializado para que el frontend pueda mostrar "Objetivo: X%"
	Weight float64 `json:"weight"`           // Peso relativo del criterio
	Goal   string  `json:"goal"`             // Estrategia: "minimize","maximize","target","homogeneity","heterogeneity"
	Type   string  `json:"type,omitempty"`  // Tipo: "numeric", "categorical", "social", "relational", "statistical"
	Label  string  `json:"label,omitempty"` // Etiqueta descriptiva del criterio
}

// TeamFormationResult representa el resultado de la formación de equipos
type TeamFormationResult struct {
	Teams              []Team                     `json:"teams"`
	UnassignedMembers  []TeamMember               `json:"unassigned_members,omitempty"`
	GlobalScore        float64                    `json:"global_score"`
	Algorithm          string                     `json:"algorithm"`
	ProvidersUsed      []string                   `json:"providers_used"`
	Weights            map[string]float64         `json:"weights"`
	Targets            map[string]float64         `json:"targets,omitempty"`   // Valores target de cada criterio
	Goals              map[string]string          `json:"goals,omitempty"`     // Estrategia de evaluación por criterio numérico
	CriteriaParameters map[string]CriterionConfig `json:"criteria_parameters"` // Configuración completa por criterio
	EquityStrategy     string                     `json:"equity_strategy"`     // Minimax aggregation strategy: "minimax" when active, "" when not
	ExecutionTimeMs    int64                      `json:"execution_time_ms"`
	Metadata           map[string]interface{}     `json:"metadata,omitempty"` // Información adicional
}

// ============================================================================
// THESIS FORMULA IMPLEMENTATION FOR f2 (RELATIONAL COST)
// ============================================================================
//
// The f2 relational cost is implemented according to the thesis formula:
//
//   C'_ij = Σ_p ω_p * (1 - S_p * T_p(i,j)) / 2
//   S_rel(G) = min(1, Σ_ij C'_ij / (|G|(|G|-1)/2))
//   f_2(x) = (1/m) * Σ_G S_rel(G)
//
// Where:
//   - T_p(i,j) ∈ [-1, 1] is the raw compatibility from provider p
//   - ω_p is the relative weight of provider p (Σ ω_p = 1)
//   - S_p ∈ {-1, 1} is the social intentionality operator
//     (1 = harmony, -1 = controlled friction)
//   - C'_ij is the interaction cost between pair (i,j)
//   - S_rel(G) is the social-relational cohesion of group G
//
// ============================================================================

// ProviderIntentionality holds the S_p and weight for a provider
type ProviderIntentionality struct {
	ProviderID           string
	Weight               float64
	SocialIntentionality int // 1 for harmony, -1 for controlled friction
}

// ComputeRelationalCost computes C'_ij for a pair using the thesis formula:
// C'_ij = Σ_p ω_p * (1 - S_p * T_p(i,j)) / 2
func ComputeRelationalCost(
	providerScores map[string]float64, // T_p(i,j) values per provider
	providers []ProviderIntentionality,
) float64 {
	var cost float64
	var totalWeight float64

	for _, p := range providers {
		Tp, ok := providerScores[p.ProviderID]
		if !ok {
			continue
		}
		// Clamp T_p to [-1, 1] as per thesis
		ifTp := Tp
		if ifTp < -1 {
			ifTp = -1
		}
		if ifTp > 1 {
			ifTp = 1
		}

		Sp := float64(p.SocialIntentionality) // S_p ∈ {-1, 1}
		Wp := p.Weight

		// C'_ij component = ω_p * (1 - S_p * T_p(i,j)) / 2
		contribution := Wp * (1.0 - Sp*ifTp) / 2.0
		cost += contribution
		totalWeight += Wp
	}

	if totalWeight == 0 {
		return 0.5 // Neutral when no providers
	}

	// Normalize by total weight
	return cost / totalWeight
}

// ComputeGroupRelationalCohesion computes S_rel(G) according to thesis:
// S_rel(G) = min(1, Σ_ij C'_ij / (|G|(|G|-1)/2))
func ComputeGroupRelationalCohesion(relationalCosts []float64, groupSize int) float64 {
	if groupSize <= 1 {
		return 0.0
	}

	// Sum all pairwise relational costs
	var sumCost float64
	for _, c := range relationalCosts {
		sumCost += c
	}

	// Max pairs = |G|(|G|-1)/2
	maxPairs := float64(groupSize * (groupSize - 1) / 2)
	if maxPairs <= 0 {
		return 0.0
	}

	// S_rel(G) = min(1, sumCost / maxPairs)
	cohesion := sumCost / maxPairs
	if cohesion > 1.0 {
		return 1.0
	}
	return cohesion
}
