package dto

// TeamFormationRequestDTO representa el JSON de entrada para formación de equipos con criterios dinámicos
type TeamFormationRequestDTO struct {
	SectionID          string                           `json:"section_id"`
	Candidates         []string                         `json:"candidates"`
	TeamSize           int                              `json:"team_size"`
	MinTeamSize        int                              `json:"min_team_size,omitempty"`
	MaxTeamSize        int                              `json:"max_team_size,omitempty"`
	Algorithm          string                           `json:"algorithm,omitempty"` // "genetic", "greedy", etc.
	CriteriaParameters map[string]CriterionParameterDTO `json:"criteria_parameters"`
	DemographicData    map[string]StudentAttributesDTO  `json:"demographic_data"`
	// EquityStrategy defines the aggregation strategy for the genetic engine fitness.
	// Values: "minimax" (worst team drives fitness), "average" (default).
	EquityStrategy string `json:"equity_strategy,omitempty"`
}

// CriterionParameterDTO representa la configuración de un criterio individual
type CriterionParameterDTO struct {
	Target float64 `json:"target"` // Valor objetivo para el criterio (0.0 - 1.0)
	Weight float64 `json:"weight"` // Peso del criterio en la función de fitness (0.0 - 1.0)
	// Goal define la estrategia de evaluación para criterios numéricos.
	// Valores válidos: "minimize", "maximize", "target", "heterogeneity", "homogeneity".
	// Si está vacío, el mapper aplica el default "target".
	Goal  string `json:"goal,omitempty"`
	Type  string `json:"type,omitempty"`  // Tipo: "numeric", "categorical", "social", "relational", "statistical"
	Label string `json:"label,omitempty"` // Etiqueta descriptiva del criterio
}

// StudentAttributesDTO representa los atributos de un estudiante en el JSON de entrada
type StudentAttributesDTO struct {
	NumericAttrs      map[string]float64 `json:"numeric_attrs,omitempty"`
	CategoricalAttrs  map[string]string  `json:"categorical_attrs,omitempty"`
	SocialConnections map[string]float64 `json:"social_connections,omitempty"`
}

// Validate valida la estructura del DTO
func (dto *TeamFormationRequestDTO) Validate() error {
	if dto.SectionID == "" {
		return ErrMissingSectionID
	}

	if len(dto.Candidates) == 0 {
		return ErrNoCandidates
	}

	if dto.TeamSize < 2 {
		return ErrInvalidTeamSize
	}

	if len(dto.CriteriaParameters) == 0 {
		return ErrNoCriteria
	}

	// Validar que cada criterio tenga valores válidos
	for criterionName, params := range dto.CriteriaParameters {
		if params.Weight < 0 || params.Weight > 1 {
			return NewErrInvalidCriterionWeight(criterionName, params.Weight)
		}

		// Target ∈ [-1, 1]: para criterios sociales, S_p indica intencionalidad social
		// (-1 = friccion controlada, 0 = neutral, 1 = armonia)
		// Para otros criterios, target ∈ [0, 1]
		if params.Target < -1 || params.Target > 1 {
			return NewErrInvalidCriterionTarget(criterionName, params.Target)
		}
	}

	// Validar que haya datos demográficos para los candidatos
	if len(dto.DemographicData) == 0 {
		return ErrNoDemographicData
	}

	return nil
}

// GetAlgorithm retorna el algoritmo especificado o el valor por defecto
func (dto *TeamFormationRequestDTO) GetAlgorithm() string {
	if dto.Algorithm == "" {
		return "genetic"
	}
	return dto.Algorithm
}

// GetMinTeamSize retorna el tamaño mínimo de equipo o calcula uno por defecto
func (dto *TeamFormationRequestDTO) GetMinTeamSize() int {
	if dto.MinTeamSize > 0 {
		return dto.MinTeamSize
	}
	return dto.TeamSize - 1
}

// GetMaxTeamSize retorna el tamaño máximo de equipo o calcula uno por defecto
func (dto *TeamFormationRequestDTO) GetMaxTeamSize() int {
	if dto.MaxTeamSize > 0 {
		return dto.MaxTeamSize
	}
	return dto.TeamSize + 1
}
