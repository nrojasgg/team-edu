package mapper

import (
	"fmt"
	"strings"

	"ms-core/internal/adapters/dto"
	"ms-core/internal/core/domain"
	"ms-core/internal/core/genetic"
)

// TeamFormationMapper maneja la conversión de DTOs a objetos del dominio
type TeamFormationMapper struct {
	// Catálogos de criterios conocidos por tipo
	numericCriteria     map[string]bool
	categoricalCriteria map[string]bool
	socialCriteria      map[string]bool
	resolvedCriteria    map[string]CriterionType
}

// NewTeamFormationMapper crea una nueva instancia del mapper con catálogos predefinidos
func NewTeamFormationMapper() *TeamFormationMapper {
	return &TeamFormationMapper{
		numericCriteria:     buildNumericCriteriaCatalog(),
		categoricalCriteria: buildCategoricalCriteriaCatalog(),
		socialCriteria:      buildSocialCriteriaCatalog(),
		resolvedCriteria:    make(map[string]CriterionType),
	}
}

// MapToDomain convierte el DTO de entrada en estructuras del dominio listas para el algoritmo genético
// Retorna: estudiantes, config, factory, context, error
func (m *TeamFormationMapper) MapToDomain(dto *dto.TeamFormationRequestDTO) (
	students []*domain.Student,
	config *domain.Config,
	factory *genetic.CriterionFactory,
	ctx domain.Context,
	err error,
) {
	// 1. Validar DTO de entrada
	if err := dto.Validate(); err != nil {
		return nil, nil, nil, domain.Context{}, fmt.Errorf("validation failed: %w", err)
	}

	// 2. Mapear estudiantes con sus atributos
	students, err = m.mapStudents(dto)
	if err != nil {
		return nil, nil, nil, domain.Context{}, fmt.Errorf("failed to map students: %w", err)
	}

	// 3. Construir configuración con pesos y targets normalizados
	config, err = m.buildConfig(dto.CriteriaParameters)
	if err != nil {
		return nil, nil, nil, domain.Context{}, fmt.Errorf("failed to build config: %w", err)
	}

	// 4. Construir factoría de criterios con detección automática de tipos
	factory, err = m.buildCriterionFactory(dto.CriteriaParameters)
	if err != nil {
		return nil, nil, nil, domain.Context{}, fmt.Errorf("failed to build criterion factory: %w", err)
	}

	// 5. Construir contexto de evaluación
	ctx = m.buildContext(dto, config)

	return students, config, factory, ctx, nil
}

// mapStudents convierte los datos demográficos en objetos Student del dominio
func (m *TeamFormationMapper) mapStudents(dto *dto.TeamFormationRequestDTO) ([]*domain.Student, error) {
	students := make([]*domain.Student, 0, len(dto.Candidates))

	for _, candidateID := range dto.Candidates {
		// Buscar datos demográficos para este candidato
		attrs, exists := dto.DemographicData[candidateID]
		if !exists {
			// Si no hay datos, crear estudiante con mapas vacíos (el algoritmo manejará datos faltantes)
			students = append(students, &domain.Student{
				ID:               candidateID,
				NumericAttrs:     make(map[string]float64),
				CategoricalAttrs: make(map[string]string),
			})
			continue
		}

		// Crear estudiante con atributos del DTO
		student := &domain.Student{
			ID:               candidateID,
			NumericAttrs:     make(map[string]float64),
			CategoricalAttrs: make(map[string]string),
		}

		// Copiar atributos numéricos
		if attrs.NumericAttrs != nil {
			for key, value := range attrs.NumericAttrs {
				student.NumericAttrs[key] = value
			}
		}

		// Copiar atributos categóricos
		if attrs.CategoricalAttrs != nil {
			for key, value := range attrs.CategoricalAttrs {
				student.CategoricalAttrs[key] = value
			}
		}

		students = append(students, student)
	}

	if len(students) == 0 {
		return nil, fmt.Errorf("no students mapped from candidates")
	}

	return students, nil
}

// buildConfig construye la configuración del algoritmo genético con pesos y targets normalizados
func (m *TeamFormationMapper) buildConfig(criteriaParams map[string]dto.CriterionParameterDTO) (*domain.Config, error) {
	config := &domain.Config{
		Weights:             make(map[string]float64),
		Targets:             make(map[string]float64),
		UseQuadraticPenalty: false, // Puede ser configurable en el futuro
	}

	// Calcular suma total de pesos para normalización
	totalWeight := 0.0
	for _, params := range criteriaParams {
		totalWeight += params.Weight
	}

	if totalWeight == 0 {
		return nil, fmt.Errorf("total weight cannot be zero")
	}

	// Normalizar pesos y asignar targets
	for criterionName, params := range criteriaParams {
		// Peso normalizado (suma = 1.0)
		config.Weights[criterionName] = params.Weight / totalWeight

		// Target sin modificar
		config.Targets[criterionName] = params.Target
	}

	return config, nil
}

// numericStrategy returns the fitness strategy string to use for a numeric criterion.
// According to thesis formula: φ = |Tₖ - Dₖ| where Tₖ ∈ [0, 1]
// The goal is used to determine the base strategy, but Tₖ (from target) overrides it.
// All strategies now use Tₖ directly via getTargetDiversity(ctx).
func numericStrategy(goal string) string {
	switch goal {
	case "minimize", "maximize", "target", "homogeneity", "heterogeneity", "balanced":
		return goal
	default:
		// For any unknown goal, use "homogeneity" as default
		// Tₖ (target) will determine the actual behavior
		return "homogeneity"
	}
}

// buildCriterionFactory construye la factoría de criterios con detección automática de tipos
func (m *TeamFormationMapper) buildCriterionFactory(criteriaParams map[string]dto.CriterionParameterDTO) (*genetic.CriterionFactory, error) {
	factory := genetic.NewCriterionFactory()

	for criterionName, params := range criteriaParams {
		// Detectar o resolver tipo de criterio (usa params.Type si viene definido, o detecta)
		criterionType, err := m.resolveCriterionType(criterionName, params)
		if err != nil {
			return nil, fmt.Errorf("failed to detect type for criterion '%s': %w", criterionName, err)
		}
		m.resolvedCriteria[criterionName] = criterionType
		if criterionType == CriterionTypeSocial {
			m.socialCriteria[strings.ToLower(strings.TrimSpace(criterionName))] = true
		}

		// Crear e instanciar el criterio según su tipo
		var criterion domain.Criterion

		switch criterionType {
		case CriterionTypeNumeric:
			// La estrategia se determina por el campo Goal enviado por el usuario.
			// Si no viene nada, se usa "target" como valor por defecto.
			strategy := numericStrategy(params.Goal)
			criterion = genetic.NewNumericCriterion(criterionName, criterionName, strategy)

		case CriterionTypeCategorical:
			// Para criterios categóricos, usar índice de Blau (diversidad)
			criterion = genetic.NewCategoricalCriterion(criterionName, criterionName)

		case CriterionTypeSocial:
			// Para criterios sociales, usar f2 con S_p del target (frontend ya convierte [0,1] → [-1,1])
			criterion = genetic.NewSocialCriterion(criterionName, params.Target)

		default:
			return nil, fmt.Errorf("unknown criterion type for '%s'", criterionName)
		}

		// Registrar el criterio en la factoría
		factory.Register(criterion)
	}

	return factory, nil
}

// buildContext construye el contexto de evaluación con grafo social y estadísticas globales.
// También calcula CategoryCounts: el número de valores distintos por atributo categórico
// observados en la población completa (todos los candidatos), de modo que CategoricalCriterion
// pueda normalizar el índice de Blau con un denominador estable (k_global) en lugar del
// k observado dentro de cada grupo individual, lo que inflaría los scores de grupos pequeños.
func (m *TeamFormationMapper) buildContext(dto *dto.TeamFormationRequestDTO, config *domain.Config) domain.Context {
	ctx := domain.Context{
		SocialGraph:    make(domain.SocialGraph),
		GlobalStats:    make(map[string]interface{}),
		Targets:        config.Targets,
		Constraints:    make(map[string]interface{}),
		CategoryCounts: make(map[string]int),
	}

	// Construir grafo social desde las conexiones en demographic_data.
	for studentID, attrs := range dto.DemographicData {
		if len(attrs.SocialConnections) > 0 {
			ctx.SocialGraph[studentID] = attrs.SocialConnections
		}
	}

	// Calcular CategoryCounts: recorrer todos los candidatos y acumular los valores
	// distintos por cada atributo categórico presente en los criterios configurados.
	//
	// Solo contamos los atributos que son criterios categóricos activos, para no
	// poblar el mapa con atributos irrelevantes.
	categoricalKeys := make(map[string]bool)
	for criterionName := range dto.CriteriaParameters {
		normalized := strings.ToLower(strings.TrimSpace(criterionName))
		if m.categoricalCriteria[normalized] {
			categoricalKeys[criterionName] = true
		}
	}

	if len(categoricalKeys) > 0 {
		// distinct[attrKey] = conjunto de valores únicos observados en la población
		distinct := make(map[string]map[string]struct{}, len(categoricalKeys))
		for key := range categoricalKeys {
			distinct[key] = make(map[string]struct{})
		}

		for _, attrs := range dto.DemographicData {
			if attrs.CategoricalAttrs == nil {
				continue
			}
			for key := range categoricalKeys {
				if val, ok := attrs.CategoricalAttrs[key]; ok && val != "" {
					distinct[key][val] = struct{}{}
				}
			}
		}

		for key, vals := range distinct {
			if len(vals) > 0 {
				ctx.CategoryCounts[key] = len(vals)
			}
		}
	}

	// Agregar metadatos al contexto.
	ctx.Constraints["section_id"] = dto.SectionID
	ctx.Constraints["team_size"] = dto.TeamSize
	ctx.Constraints["algorithm"] = dto.GetAlgorithm()

	return ctx
}

// CriterionType representa el tipo de un criterio
type CriterionType string

const (
	CriterionTypeNumeric     CriterionType = "numeric"
	CriterionTypeCategorical CriterionType = "categorical"
	CriterionTypeSocial      CriterionType = "social"
)

// parseCriterionType convierte strings de tipo explícitos del DTO/API a CriterionType
func parseCriterionType(rawType string) (CriterionType, bool) {
	switch strings.ToLower(strings.TrimSpace(rawType)) {
	case "numeric", "numerico", "numérico", "statistical", "estadistico", "estadístico":
		return CriterionTypeNumeric, true
	case "categorical", "categorico", "categórico":
		return CriterionTypeCategorical, true
	case "social", "relational", "relacional":
		return CriterionTypeSocial, true
	default:
		return "", false
	}
}

// resolveCriterionType resuelve el tipo de criterio priorizando el tipo explícito provisto en DTO
func (m *TeamFormationMapper) resolveCriterionType(criterionName string, params dto.CriterionParameterDTO) (CriterionType, error) {
	if params.Type != "" {
		if cType, ok := parseCriterionType(params.Type); ok {
			return cType, nil
		}
	}
	return m.detectCriterionType(criterionName)
}

// detectCriterionType detecta automáticamente el tipo de criterio basándose en su nombre
func (m *TeamFormationMapper) detectCriterionType(criterionName string) (CriterionType, error) {
	// Normalizar nombre (lowercase, trim)
	normalized := strings.ToLower(strings.TrimSpace(criterionName))

	// Verificar en catálogo de criterios sociales (más específico primero)
	if m.socialCriteria[normalized] {
		return CriterionTypeSocial, nil
	}

	// Verificar en catálogo de criterios numéricos
	if m.numericCriteria[normalized] {
		return CriterionTypeNumeric, nil
	}

	// Verificar en catálogo de criterios categóricos
	if m.categoricalCriteria[normalized] {
		return CriterionTypeCategorical, nil
	}

	// Si no se encuentra en ningún catálogo, retornar error
	return "", dto.NewErrUnknownCriterionType(criterionName)
}

// buildNumericCriteriaCatalog construye el catálogo de criterios numéricos conocidos
func buildNumericCriteriaCatalog() map[string]bool {
	return map[string]bool{
		// Atributos demográficos numéricos
		"age":         true,
		"edad":        true,
		"gpa":         true,
		"promedio":    true,
		"grade":       true,
		"nota":        true,
		"experience":  true,
		"experiencia": true,

		// Big Five (OCEAN)
		"bfi44":             true,
		"bfi_44":            true,
		"openness":          true,
		"apertura":          true,
		"conscientiousness": true,
		"responsabilidad":   true,
		"extraversion":      true,
		"extraversión":      true,
		"agreeableness":     true,
		"amabilidad":        true,
		"neuroticism":       true,
		"neuroticismo":      true,

		// Otros psicométricos numéricos
		"iq":          true,
		"eq":          true,
		"performance": true,
		"rendimiento": true,
		"attendance":  true,
		"asistencia":  true,

		// Métricas específicas de dominio
		"hours_available":   true,
		"horas_disponibles": true,
		"skill_level":       true,
		"nivel_habilidad":   true,
	}
}

// buildCategoricalCriteriaCatalog construye el catálogo de criterios categóricos conocidos
func buildCategoricalCriteriaCatalog() map[string]bool {
	return map[string]bool{
		// Demográficos
		"sex":          true,
		"sexo":         true,
		"gender":       true,
		"genero":       true,
		"género":       true,
		"ethnicity":    true,
		"etnia":        true,
		"race":         true,
		"raza":         true,
		"nationality":  true,
		"nacionalidad": true,

		// Académicos
		"career":         true,
		"carrera":        true,
		"career_code":    true,
		"codigo_carrera": true,
		"major":          true,
		"especialidad":   true,
		"department":     true,
		"departamento":   true,
		"faculty":        true,
		"facultad":       true,
		"program":        true,
		"programa":       true,

		// Idiomas y ubicación
		"language":     true,
		"idioma":       true,
		"timezone":     true,
		"zona_horaria": true,
		"country":      true,
		"pais":         true,
		"país":         true,
		"city":         true,
		"ciudad":       true,

		// Eneagrama y MBTI
		"enneagram": true,
		"eneagrama": true,
		"eneatype":  true,
		"eneatipo":  true,
		"mbti":      true,
		"mbti_type": true,
		"tipo_mbti": true,

		// Roles y preferencias
		"role":        true,
		"rol":         true,
		"position":    true,
		"posicion":    true,
		"posición":    true,
		"preference":  true,
		"preferencia": true,
		"skill":       true,
		"habilidad":   true,
	}
}

// buildSocialCriteriaCatalog construye el catálogo de criterios sociales conocidos
func buildSocialCriteriaCatalog() map[string]bool {
	return map[string]bool{
		// Sociometría y afinidad social
		"social":          true,
		"sociometric":     true,
		"sociometrico":    true,
		"sociométrico":    true,
		"sociogram":       true,
		"sociograma":      true,
		"affinity":        true,
		"afinidad":        true,
		"social_affinity": true,
		"afinidad_social": true,
		"friendship":      true,
		"amistad":         true,
		"peer":            true,
		"compañero":       true,
		"companero":       true,
		"network":         true,
		"red":             true,
		"connection":      true,
		"conexion":        true,
		"conexión":        true,
	}
}

// RegisterCustomCriterion permite registrar un nuevo criterio personalizado en tiempo de ejecución
func (m *TeamFormationMapper) RegisterCustomCriterion(criterionName string, criterionType CriterionType) {
	normalized := strings.ToLower(strings.TrimSpace(criterionName))

	switch criterionType {
	case CriterionTypeNumeric:
		m.numericCriteria[normalized] = true
	case CriterionTypeCategorical:
		m.categoricalCriteria[normalized] = true
	case CriterionTypeSocial:
		m.socialCriteria[normalized] = true
	}
}

// GetCriterionType retorna el tipo de un criterio si está registrado
func (m *TeamFormationMapper) GetCriterionType(criterionName string) (CriterionType, error) {
	if cType, ok := m.resolvedCriteria[criterionName]; ok {
		return cType, nil
	}
	return m.detectCriterionType(criterionName)
}

// IsSocialCriterion returns true if the given criterion name is a social criterion.
func (m *TeamFormationMapper) IsSocialCriterion(criterionName string) bool {
	if cType, ok := m.resolvedCriteria[criterionName]; ok {
		return cType == CriterionTypeSocial
	}
	normalized := strings.ToLower(strings.TrimSpace(criterionName))
	return m.socialCriteria[normalized]
}

// GetSocialCriteriaNames returns the set of all registered social criterion names.
func (m *TeamFormationMapper) GetSocialCriteriaNames() map[string]bool {
	result := make(map[string]bool, len(m.socialCriteria)+len(m.resolvedCriteria))
	for k, v := range m.socialCriteria {
		result[k] = v
	}
	for k, cType := range m.resolvedCriteria {
		if cType == CriterionTypeSocial {
			result[strings.ToLower(strings.TrimSpace(k))] = true
		}
	}
	return result
}

// GetResolvedCriteriaTypes returns a copy of all dynamically resolved criteria types.
func (m *TeamFormationMapper) GetResolvedCriteriaTypes() map[string]CriterionType {
	res := make(map[string]CriterionType, len(m.resolvedCriteria))
	for k, v := range m.resolvedCriteria {
		res[k] = v
	}
	return res
}
