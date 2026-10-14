package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"ms-core/internal/adapters/dto"
	"ms-core/internal/adapters/mapper"
	"ms-core/internal/core"
	"ms-core/internal/core/domain"
	"ms-core/internal/core/genetic"
	"ms-core/internal/database"
	"ms-core/internal/queue"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

// TeamMember is a local alias so async helpers in this package can reference
// the type without importing core directly.
type TeamMember = core.TeamMember

// oceanTraits is the set of Big Five trait names that NumericCriterion can evaluate.
// Used to detect when a criteria_parameters request needs OCEAN enrichment.
var oceanTraits = map[string]bool{
	"openness":          true,
	"conscientiousness": true,
	"extraversion":      true,
	"agreeableness":     true,
	"neuroticism":       true,
}

// TeamFormationHandler maneja las solicitudes de formación de equipos
type TeamFormationHandler struct {
	engine   *core.TeamFormationEngine
	registry *core.ProviderRegistry
	// tfPlugin es el plugin de formación de equipos (subprocess via hashicorp/go-plugin).
	// Si es nil, se usa el engine directamente (modo legado / sin plugin cargado).
	tfPlugin core.Plugin
	// pm es el gestor de plugins. Se usa para obtener los plugins de personalidad
	// (enneagram, mbti) y enriquecer los numeric_attrs con los vectores OCEAN
	// antes de ejecutar el algoritmo genético en modo criteria.
	// Puede ser nil; en ese caso el enriquecimiento OCEAN se omite silenciosamente.
	pm *core.PluginManager
	// db es la referencia a MongoDB para guardar el historial de formaciones.
	// Puede ser nil; en ese caso el guardado se omite silenciosamente.
	db *mongo.Database
	// broker is the async task queue broker. When non-nil the *Async endpoints
	// enqueue jobs instead of executing them inline.
	broker queue.Broker
	// resultStore persists job results so the polling handler can return them.
	resultStore queue.ResultStore
}

// NewTeamFormationHandler crea una nueva instancia del handler.
// tfPlugin puede ser nil; en ese caso se usa el engine directamente (fallback).
func NewTeamFormationHandler(engine *core.TeamFormationEngine, registry *core.ProviderRegistry, tfPlugin core.Plugin) *TeamFormationHandler {
	return &TeamFormationHandler{
		engine:   engine,
		registry: registry,
		tfPlugin: tfPlugin,
	}
}

// WithPluginManager inyecta el PluginManager para que el handler pueda consultar
// los plugins de personalidad (enneagram/mbti) al enriquecer los vectores OCEAN.
func (h *TeamFormationHandler) WithPluginManager(pm *core.PluginManager) *TeamFormationHandler {
	h.pm = pm
	return h
}

// WithDB establece la referencia a MongoDB para persistir el historial de formaciones.
func (h *TeamFormationHandler) WithDB(db *mongo.Database) *TeamFormationHandler {
	h.db = db
	return h
}

// groupHistoryTeam es el subdocumento de un equipo dentro del historial.
type groupHistoryTeam struct {
	TeamID             string                                  `bson:"team_id" json:"team_id"`
	MemberIDs          []string                                `bson:"member_ids" json:"member_ids"`
	AffinityScore      float64                                 `bson:"affinity_score" json:"affinity_score"`
	ProviderScores     map[string]float64                      `bson:"provider_scores,omitempty" json:"provider_scores,omitempty"`
	IsFeasible         bool                                    `bson:"is_feasible" json:"is_feasible"`
	CriterionBreakdown map[string]core.CriterionBreakdownEntry `bson:"criterion_breakdown,omitempty" json:"criterion_breakdown,omitempty"`
}

// GroupHistoryRecord es el documento completo que se almacena en groups_historydata.
type GroupHistoryRecord struct {
	SectionID          string                          `bson:"section_id" json:"section_id"`
	FormedAt           time.Time                       `bson:"formed_at" json:"formed_at"`
	Algorithm          string                          `bson:"algorithm" json:"algorithm"`
	ProvidersUsed      []string                        `bson:"providers_used" json:"providers_used"`
	Weights            map[string]float64              `bson:"weights,omitempty" json:"weights,omitempty"`
	Targets            map[string]float64              `bson:"targets,omitempty" json:"targets,omitempty"`
	Goals              map[string]string               `bson:"goals,omitempty" json:"goals,omitempty"`
	CriteriaParameters map[string]core.CriterionConfig `bson:"criteria_parameters,omitempty" json:"criteria_parameters,omitempty"`
	EquityStrategy     string                          `bson:"equity_strategy" json:"equity_strategy"`
	GlobalScore        float64                         `bson:"global_score" json:"global_score"`
	ExecutionMs        int64                           `bson:"execution_time_ms" json:"execution_time_ms"`
	Teams              []groupHistoryTeam              `bson:"teams" json:"teams"`
	Metadata           map[string]interface{}          `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

// saveGroupHistory persists a formation result to groups_historydata (best-effort, non-blocking).
func (h *TeamFormationHandler) saveGroupHistory(sectionID string, result *core.TeamFormationResult) {
	if h.db == nil || result == nil {
		return
	}
	teams := make([]groupHistoryTeam, len(result.Teams))
	for i, t := range result.Teams {
		memberIDs := make([]string, len(t.Members))
		for j, m := range t.Members {
			memberIDs[j] = m.ID
		}
		teams[i] = groupHistoryTeam{
			TeamID:             t.ID,
			MemberIDs:          memberIDs,
			AffinityScore:      t.AffinityScore,
			ProviderScores:     t.Metadata.ProviderScores,
			IsFeasible:         t.Metadata.IsFeasible,
			CriterionBreakdown: t.Metadata.CriterionBreakdown,
		}
	}
	record := GroupHistoryRecord{
		SectionID:          sectionID,
		FormedAt:           time.Now().UTC(),
		Algorithm:          result.Algorithm,
		ProvidersUsed:      result.ProvidersUsed,
		Weights:            result.Weights,
		Targets:            result.Targets,
		Goals:              result.Goals,
		CriteriaParameters: result.CriteriaParameters,
		EquityStrategy:     result.EquityStrategy,
		GlobalScore:        result.GlobalScore,
		ExecutionMs:        result.ExecutionTimeMs,
		Teams:              teams,
		Metadata:           result.Metadata,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := h.db.Collection(database.GroupsHistoryCollection).InsertOne(ctx, record); err != nil {
		log.Printf("[warn] saveGroupHistory: %v", err)
	}
}

// teamFormationPluginReq es la estructura que se envía al plugin team_formation.
type teamFormationPluginReq struct {
	Mode            string                        `json:"mode"`
	SectionID       string                        `json:"section_id,omitempty"`
	Candidates      []string                      `json:"candidates,omitempty"`
	TeamSize        int                           `json:"team_size"`
	MinTeamSize     int                           `json:"min_team_size"`
	MaxTeamSize     int                           `json:"max_team_size"`
	Algorithm       string                        `json:"algorithm,omitempty"`
	AffinityMatrix  map[string]map[string]float64 `json:"affinity_matrix,omitempty"`
	ProviderIDs     []string                      `json:"provider_ids,omitempty"`
	ProviderWeights map[string]float64            `json:"provider_weights,omitempty"`
	GeneticPopSize  int                           `json:"genetic_pop_size,omitempty"`
	GeneticGens     int                           `json:"genetic_gens,omitempty"`
	GeneticMutation float64                       `json:"genetic_mutation,omitempty"`
	CriteriaRequest *dto.TeamFormationRequestDTO  `json:"criteria_request,omitempty"`
	EquityStrategy  string                        `json:"equity_strategy,omitempty"`
}

// marshalToTeamFormationResult convierte la respuesta genérica del plugin (interface{})
// al tipo concreto mediante doble marshal JSON.
func marshalToTeamFormationResult(raw interface{}) (*core.TeamFormationResult, error) {
	if raw == nil {
		return nil, fmt.Errorf("team_formation plugin returned nil result")
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal plugin result: %w", err)
	}
	var result core.TeamFormationResult
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plugin result: %w", err)
	}
	return &result, nil
}

// FormTeamsRequest representa la solicitud HTTP para formar equipos
type FormTeamsRequest struct {
	SectionID   string   `json:"section_id"`
	Candidates  []string `json:"candidates"`
	TeamSize    int      `json:"team_size"`
	MinTeamSize int      `json:"min_team_size"`
	MaxTeamSize int      `json:"max_team_size"`
	Algorithm   string   `json:"algorithm"` // "greedy" o "genetic" (opcional, usa config por defecto)
}

// FlexibleFormTeamsRequest representa la solicitud flexible para formar equipos
// Permite al usuario elegir qué providers usar y sus pesos
type FlexibleFormTeamsRequest struct {
	SectionID   string   `json:"section_id"`
	Candidates  []string `json:"candidates"`
	TeamSize    int      `json:"team_size"`
	MinTeamSize int      `json:"min_team_size"`
	MaxTeamSize int      `json:"max_team_size"`
	Algorithm   string   `json:"algorithm"` // "greedy" o "genetic" (opcional, default: "genetic")

	// Configuración flexible de providers
	// El usuario puede elegir qué fuentes de afinidad usar
	ProviderConfig *ProviderConfigRequest `json:"provider_config,omitempty"`

	// Configuración dinámica de criterios (nuevo formato)
	// Permite especificar target y weight para cada criterio
	CriteriaParameters map[string]CriterionConfig `json:"criteria_parameters,omitempty"`

	// Datos demográficos y atributos de estudiantes
	DemographicData map[string]StudentAttributes `json:"demographic_data,omitempty"`

	// EquityStrategy controls the fitness aggregation strategy in the genetic engine.
	// "minimax"  → worst-team score drives fitness (penalizes garbage teams).
	// "" (empty) → mean(scores) (default, backward-compatible).
	EquityStrategy string `json:"equity_strategy"`
}

// CriterionConfig representa la configuración de un criterio individual
// CriterionConfig es un alias del tipo definido en el paquete core para uso local en este handler.
type CriterionConfig = core.CriterionConfig

// StudentAttributes representa los atributos de un estudiante
type StudentAttributes struct {
	NumericAttrs      map[string]float64 `json:"numeric_attrs,omitempty"`
	CategoricalAttrs  map[string]string  `json:"categorical_attrs,omitempty"`
	SocialConnections map[string]float64 `json:"social_connections,omitempty"`
}

// ProviderConfigRequest representa la configuración de providers seleccionados
type ProviderConfigRequest struct {
	// PersonalityTest: "enneagram", "mbti", o "" (ninguno)
	// Permite elegir qué test de personalidad usar
	PersonalityTest string `json:"personality_test"`

	// PersonalityWeight: peso del test de personalidad (0.0 - 1.0)
	PersonalityWeight float64 `json:"personality_weight"`

	// UseSociogram: true para incluir sociometría
	UseSociogram bool `json:"use_sociogram"`

	// SociogramWeight: peso de la sociometría (0.0 - 1.0)
	SociogramWeight float64 `json:"sociogram_weight"`

	// CustomProviders: configuración avanzada para usuarios que quieren control total
	// Ejemplo: {"enneagram": 0.3, "mbti": 0.2, "sociogram": 0.5}
	// Si se proporciona, anula PersonalityTest y UseSociogram
	CustomProviders map[string]float64 `json:"custom_providers,omitempty"`
}

// UpdateWeightsRequest representa la solicitud para actualizar pesos
type UpdateWeightsRequest struct {
	Weights map[string]float64 `json:"weights"`
}

// FormTeams procesa la solicitud de formación de equipos
// @Summary Formar equipos óptimos
// @Description Forma equipos usando los AffinityProviders configurados
// @Tags team-formation
// @Accept json
// @Produce json
// @Param request body FormTeamsRequest true "Solicitud de formación"
// @Success 200 {object} core.TeamFormationResult
// @Router /api/v1/team-formation/form [post]
func (h *TeamFormationHandler) FormTeams(c *fiber.Ctx) error {
	var req FormTeamsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   errInvalidRequest,
			"details": err.Error(),
		})
	}

	if req.SectionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "section_id is required"})
	}
	if len(req.Candidates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "candidates is required"})
	}
	if req.TeamSize < 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "team_size must be at least 2"})
	}

	if req.MinTeamSize == 0 {
		req.MinTeamSize = req.TeamSize - 1
	}
	if req.MaxTeamSize == 0 {
		req.MaxTeamSize = req.TeamSize + 1
	}

	candidates := make([]core.TeamMember, len(req.Candidates))
	for i, id := range req.Candidates {
		candidates[i] = core.TeamMember{ID: id}
	}

	formationRequest := core.TeamFormationRequest{
		SectionID:   req.SectionID,
		Candidates:  candidates,
		TeamSize:    req.TeamSize,
		MinTeamSize: req.MinTeamSize,
		MaxTeamSize: req.MaxTeamSize,
	}

	ctx := context.WithValue(c.Context(), core.ContextKeySectionID, req.SectionID)
	result, err := h.executeFormTeams(ctx, c, req, formationRequest, candidates)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to form teams",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{"success": true, "data": result})
}

// executeFormTeams runs the team formation via the plugin if available, otherwise via the engine.
func (h *TeamFormationHandler) executeFormTeams(ctx context.Context, c *fiber.Ctx, req FormTeamsRequest, formationRequest core.TeamFormationRequest, candidates []core.TeamMember) (*core.TeamFormationResult, error) {
	if h.tfPlugin == nil {
		return h.engine.FormTeams(ctx, formationRequest)
	}

	affinityMatrix := h.engine.PrecomputeAffinityMatrix(ctx, formationRequest.Candidates)
	providerInfo := h.engine.GetProviderInfo()
	providerIDs := make([]string, len(providerInfo))
	for i, pi := range providerInfo {
		providerIDs[i] = pi.ID
	}

	pluginReq := teamFormationPluginReq{
		Mode:           "affinity",
		SectionID:      req.SectionID,
		Candidates:     req.Candidates,
		TeamSize:       formationRequest.TeamSize,
		MinTeamSize:    formationRequest.MinTeamSize,
		MaxTeamSize:    formationRequest.MaxTeamSize,
		Algorithm:      req.Algorithm,
		AffinityMatrix: affinityMatrix,
		ProviderIDs:    providerIDs,
	}

	raw, pluginErr := h.tfPlugin.Execute(c.Context(), pluginReq)
	if pluginErr != nil {
		log.Printf("team_formation plugin error, falling back to engine: %v", pluginErr)
		return h.engine.FormTeams(ctx, formationRequest)
	}
	return marshalToTeamFormationResult(raw)
}

// FlexibleFormTeams procesa la solicitud flexible de formación de equipos
// Permite al usuario elegir dinámicamente qué providers usar
// @Summary Formar equipos con configuración flexible
// @Description Forma equipos permitiendo elegir qué test de personalidad y si usar sociometría, o usar criteria_parameters dinámicos
// @Tags team-formation
// @Accept json
// @Produce json
// @Param request body FlexibleFormTeamsRequest true "Solicitud de formación flexible"
// @Success 200 {object} core.TeamFormationResult
// @Router /api/v1/team-formation/form-flexible [post]
func (h *TeamFormationHandler) FlexibleFormTeams(c *fiber.Ctx) error {
	var req FlexibleFormTeamsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   errInvalidRequest,
			"details": err.Error(),
		})
	}

	if req.SectionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "section_id is required"})
	}
	if len(req.Candidates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "candidates is required"})
	}
	if req.TeamSize < 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "team_size must be at least 2"})
	}

	if req.MinTeamSize == 0 {
		req.MinTeamSize = req.TeamSize - 1
	}
	if req.MaxTeamSize == 0 {
		req.MaxTeamSize = req.TeamSize + 1
	}

	algorithm := req.Algorithm
	if algorithm == "" {
		algorithm = "genetic"
	}

	useLegacyMode := req.ProviderConfig != nil && len(req.CriteriaParameters) == 0
	useNewMode := len(req.CriteriaParameters) > 0

	if !useLegacyMode && !useNewMode {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Either provider_config or criteria_parameters must be provided",
		})
	}

	candidates := make([]core.TeamMember, len(req.Candidates))
	for i, id := range req.Candidates {
		candidates[i] = core.TeamMember{ID: id}
	}

	var result *core.TeamFormationResult
	var err error

	if useLegacyMode {
		result, err = h.runLegacyFormation(c, req, candidates, algorithm)
	} else {
		result, err = h.runCriteriaFormation(c, req, candidates, algorithm)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to form teams",
			"details": err.Error(),
		})
	}
	// Persist history record (best-effort, does not affect the response)
	go h.saveGroupHistory(req.SectionID, result)
	return c.JSON(fiber.Map{"success": true, "data": result})
}

// runLegacyFormation handles team formation via provider_config (legacy mode).
func (h *TeamFormationHandler) runLegacyFormation(c *fiber.Ctx, req FlexibleFormTeamsRequest, candidates []core.TeamMember, algorithm string) (*core.TeamFormationResult, error) {
	providerWeights, err := h.buildProviderWeights(req.ProviderConfig)
	if err != nil {
		return nil, fmt.Errorf("Invalid provider configuration: %w", err)
	}
	if len(providerWeights) == 0 {
		return nil, fmt.Errorf("At least one affinity provider must be selected")
	}

	formationRequest := core.TeamFormationRequest{
		SectionID:       req.SectionID,
		Candidates:      candidates,
		TeamSize:        req.TeamSize,
		MinTeamSize:     req.MinTeamSize,
		MaxTeamSize:     req.MaxTeamSize,
		Algorithm:       algorithm,
		ProviderWeights: providerWeights,
	}

	var result *core.TeamFormationResult
	if h.tfPlugin != nil {
		ctx := context.WithValue(c.Context(), core.ContextKeySectionID, req.SectionID)
		affinityMatrix := h.engine.PrecomputeAffinityMatrix(ctx, candidates)
		providerIDs := make([]string, 0, len(providerWeights))
		for id := range providerWeights {
			providerIDs = append(providerIDs, id)
		}
		pluginReq := teamFormationPluginReq{
			Mode: "affinity", SectionID: req.SectionID, Candidates: req.Candidates,
			TeamSize: req.TeamSize, MinTeamSize: req.MinTeamSize, MaxTeamSize: req.MaxTeamSize,
			Algorithm: algorithm, AffinityMatrix: affinityMatrix, ProviderIDs: providerIDs, ProviderWeights: providerWeights,
		}
		raw, pluginErr := h.tfPlugin.Execute(c.Context(), pluginReq)
		if pluginErr != nil {
			log.Printf("team_formation plugin error (legacy mode), falling back: %v", pluginErr)
			result, err = h.engine.FormTeamsWithConfig(c.Context(), formationRequest)
		} else {
			result, err = marshalToTeamFormationResult(raw)
		}
	} else {
		result, err = h.engine.FormTeamsWithConfig(c.Context(), formationRequest)
	}
	if err != nil {
		return nil, err
	}
	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata["provider_weights"] = providerWeights
	result.Metadata["algorithm_used"] = algorithm
	result.Metadata["mode"] = "legacy"
	return result, nil
}

// enrichWithOCEANData resolves the OCEAN vector for each candidate and injects
// the trait values into DemographicData.numeric_attrs so that NumericCriterion
// can evaluate Big Five traits instead of returning the 0.5 fallback.
//
// Resolution order (first source that returns data wins):
//  1. enneagram plugin  — get_student_eneatypes → get_ocean_from_type
//  2. mbti plugin       — get_student_ocean (if implemented)
//
// Only the traits that appear in criteria_parameters are injected;
// pre-existing numeric_attrs values are preserved.
// This is a best-effort operation: students with no data are skipped silently.
func (h *TeamFormationHandler) enrichWithOCEANData(ctx context.Context, req *FlexibleFormTeamsRequest) {
	if h.pm == nil {
		log.Printf("[ocean_enrich] no plugin manager available, skipping OCEAN enrichment")
		return
	}

	// Check whether any OCEAN trait is actually requested as a criterion.
	hasOCEANCriterion := false
	for name := range req.CriteriaParameters {
		if oceanTraits[name] {
			hasOCEANCriterion = true
			break
		}
	}
	if !hasOCEANCriterion {
		return
	}

	pluginCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Build lookup: studentID → OCEAN values.
	// Try enneagram first, then mbti as fallback.
	oceanByStudent, source := h.resolveOCEANFromPlugins(pluginCtx, req.SectionID)
	if len(oceanByStudent) == 0 {
		log.Printf("[ocean_enrich] no OCEAN data found for section %s (tried enneagram, mbti)", req.SectionID)
		return
	}

	log.Printf("[ocean_enrich] resolved OCEAN data from %q plugin for section %s (%d students)",
		source, req.SectionID, len(oceanByStudent))

	if req.DemographicData == nil {
		req.DemographicData = make(map[string]StudentAttributes)
	}

	enriched := 0
	for _, candidateID := range req.Candidates {
		ocean, found := oceanByStudent[candidateID]
		if !found {
			continue
		}

		attrs := req.DemographicData[candidateID]
		if attrs.NumericAttrs == nil {
			attrs.NumericAttrs = make(map[string]float64)
		}

		// Only inject the traits that are actually requested as criteria.
		if _, ok := req.CriteriaParameters["openness"]; ok {
			attrs.NumericAttrs["openness"] = ocean.Openness
		}
		if _, ok := req.CriteriaParameters["conscientiousness"]; ok {
			attrs.NumericAttrs["conscientiousness"] = ocean.Conscientiousness
		}
		if _, ok := req.CriteriaParameters["extraversion"]; ok {
			attrs.NumericAttrs["extraversion"] = ocean.Extraversion
		}
		if _, ok := req.CriteriaParameters["agreeableness"]; ok {
			attrs.NumericAttrs["agreeableness"] = ocean.Agreeableness
		}
		if _, ok := req.CriteriaParameters["neuroticism"]; ok {
			attrs.NumericAttrs["neuroticism"] = ocean.Neuroticism
		}

		req.DemographicData[candidateID] = attrs
		enriched++
	}

	log.Printf("[ocean_enrich] enriched %d/%d candidates with OCEAN data for section %s",
		enriched, len(req.Candidates), req.SectionID)
}

// oceanVector holds the five Big Five trait values resolved from a personality plugin.
type oceanVector struct {
	Openness          float64
	Conscientiousness float64
	Extraversion      float64
	Agreeableness     float64
	Neuroticism       float64
}

// resolveOCEANFromPlugins tries each personality plugin in order and returns the
// first map that contains at least one entry. The second return value is the
// plugin name that produced the data ("enneagram", "mbti", or "").
func (h *TeamFormationHandler) resolveOCEANFromPlugins(ctx context.Context, sectionID string) (map[string]oceanVector, string) {
	if data := h.oceanFromEnneagram(ctx, sectionID); len(data) > 0 {
		return data, "enneagram"
	}
	if data := h.oceanFromMBTI(ctx, sectionID); len(data) > 0 {
		return data, "mbti"
	}
	if data := h.oceanFromBFI44(ctx, sectionID); len(data) > 0 {
		return data, "bfi44"
	}
	return nil, ""
}

// oceanFromEnneagram calls the enneagram plugin with:
//  1. get_student_eneatypes  → map[studentID]dominantType
//  2. get_ocean_from_type    → OCEAN vector per type
//
// Types are deduplicated so the second call is made at most 9 times.
func (h *TeamFormationHandler) oceanFromEnneagram(ctx context.Context, sectionID string) map[string]oceanVector {
	plugin, err := h.pm.GetPlugin("enneagram")
	if err != nil {
		return nil
	}

	// Step 1: get dominant eneatype per student.
	raw, err := plugin.Execute(ctx, map[string]interface{}{
		"action":     "get_student_eneatypes",
		"section_id": sectionID,
	})
	if err != nil {
		log.Printf("[ocean_enrich/enneagram] get_student_eneatypes error: %v", err)
		return nil
	}

	eneatypes, ok := raw.(map[string]int)
	if !ok {
		log.Printf("[ocean_enrich/enneagram] unexpected type from get_student_eneatypes: %T", raw)
		return nil
	}
	if len(eneatypes) == 0 {
		return nil
	}

	// Step 2: resolve OCEAN per unique eneatype (cache to avoid redundant RPC calls).
	typeOCEAN := make(map[int]oceanVector)
	for _, t := range eneatypes {
		if _, cached := typeOCEAN[t]; cached {
			continue
		}
		rawOCEAN, err := plugin.Execute(ctx, map[string]interface{}{
			"action":   "get_ocean_from_type",
			"eneatype": float64(t),
		})
		if err != nil {
			log.Printf("[ocean_enrich/enneagram] get_ocean_from_type(%d) error: %v", t, err)
			continue
		}
		if vec, ok := toOceanVector(rawOCEAN); ok {
			typeOCEAN[t] = vec
		}
	}

	result := make(map[string]oceanVector, len(eneatypes))
	for studentID, t := range eneatypes {
		if vec, ok := typeOCEAN[t]; ok {
			result[studentID] = vec
		}
	}
	return result
}

// oceanFromMBTI attempts to resolve OCEAN vectors via the mbti plugin.
// The mbti plugin must support the "get_student_ocean" action that returns
// map[studentID]map[string]float64 with the five trait keys.
func (h *TeamFormationHandler) oceanFromMBTI(ctx context.Context, sectionID string) map[string]oceanVector {
	plugin, err := h.pm.GetPlugin("mbti")
	if err != nil {
		return nil
	}

	raw, err := plugin.Execute(ctx, map[string]interface{}{
		"action":     "get_student_ocean",
		"section_id": sectionID,
	})
	if err != nil {
		log.Printf("[ocean_enrich/mbti] get_student_ocean error: %v", err)
		return nil
	}

	rawMap, ok := raw.(map[string]interface{})
	if !ok {
		return nil
	}

	result := make(map[string]oceanVector, len(rawMap))
	for studentID, v := range rawMap {
		if vec, ok := toOceanVector(v); ok {
			result[studentID] = vec
		}
	}
	return result
}

// oceanFromBFI44 attempts to resolve OCEAN vectors via the bfi44 plugin.
// The bfi44 plugin must support the "get_student_ocean" action that returns
// map[studentID]map[string]float64 with the five trait keys.
func (h *TeamFormationHandler) oceanFromBFI44(ctx context.Context, sectionID string) map[string]oceanVector {
	plugin, err := h.pm.GetPlugin("bfi44")
	if err != nil {
		return nil
	}

	raw, err := plugin.Execute(ctx, map[string]interface{}{
		"action":     "get_student_ocean",
		"section_id": sectionID,
	})
	if err != nil {
		log.Printf("[ocean_enrich/bfi44] get_student_ocean error: %v", err)
		return nil
	}

	rawMap, ok := raw.(map[string]interface{})
	if !ok {
		return nil
	}

	result := make(map[string]oceanVector, len(rawMap))
	for studentID, v := range rawMap {
		if vec, ok := toOceanVector(v); ok {
			result[studentID] = vec
		}
	}
	return result
}

// toOceanVector converts the raw interface{} returned by a plugin Execute call
// into an oceanVector. The plugin may return map[string]interface{} or a struct
// that serialises to the same JSON keys.
func toOceanVector(raw interface{}) (oceanVector, bool) {
	b, err := json.Marshal(raw)
	if err != nil {
		return oceanVector{}, false
	}
	var m map[string]float64
	if err := json.Unmarshal(b, &m); err != nil {
		return oceanVector{}, false
	}
	return oceanVector{
		Openness:          m["openness"],
		Conscientiousness: m["conscientiousness"],
		Extraversion:      m["extraversion"],
		Agreeableness:     m["agreeableness"],
		Neuroticism:       m["neuroticism"],
	}, true
}

// runCriteriaFormation handles team formation using criteria_parameters (new mode).
func (h *TeamFormationHandler) runCriteriaFormation(c *fiber.Ctx, req FlexibleFormTeamsRequest, candidates []core.TeamMember, algorithm string) (*core.TeamFormationResult, error) {
	h.enrichWithOCEANData(c.Context(), &req)
	requestDTO := h.convertToDTO(req, candidates)

	log.Printf("[runCriteriaFormation] CriteriaParameters: %v", requestDTO.CriteriaParameters)
	log.Printf("[runCriteriaFormation] Has tfPlugin: %v", h.tfPlugin != nil)

	var (
		result *core.TeamFormationResult
		err    error
	)
	if h.tfPlugin != nil {
		pluginReq := teamFormationPluginReq{
			Mode: "criteria", TeamSize: req.TeamSize, MinTeamSize: req.MinTeamSize,
			MaxTeamSize: req.MaxTeamSize, Algorithm: algorithm, CriteriaRequest: requestDTO,
			EquityStrategy: req.EquityStrategy,
		}
		raw, pluginErr := h.tfPlugin.Execute(c.Context(), pluginReq)
		if pluginErr != nil {
			log.Printf("team_formation plugin error (criteria mode), falling back: %v", pluginErr)
			result, err = h.formTeamsWithDynamicCriteria(c.Context(), req, candidates)
		} else {
			result, err = marshalToTeamFormationResult(raw)
			if err == nil && result != nil {
				log.Printf("[runCriteriaFormation] Plugin result algorithm: %s", result.Algorithm)
				log.Printf("[runCriteriaFormation] Plugin result metadata mode: %v", result.Metadata["mode"])
				// The plugin result rarely carries CriteriaParameters or Goals because
				// those fields are owned by the orchestration layer, not the plugin.
				// Backfill them from the original request so the frontend always
				// receives the full config (including goal) and can forward it
				// verbatim to evaluate-groups without losing strategy information.
				if len(result.CriteriaParameters) == 0 {
					// Normalise weights the same way convertGeneticResultToTeamFormation does.
					totalWeight := 0.0
					for _, cfg := range req.CriteriaParameters {
						if cfg.Weight > 0 {
							totalWeight += cfg.Weight
						}
					}
					criteriaFull := make(map[string]core.CriterionConfig, len(req.CriteriaParameters))
					goalsFull := make(map[string]string, len(req.CriteriaParameters))
					for name, cfg := range req.CriteriaParameters {
						normalizedWeight := 0.0
						if totalWeight > 0 {
							normalizedWeight = cfg.Weight / totalWeight
						}
						criteriaFull[name] = core.CriterionConfig{
							Target: cfg.Target,
							Weight: normalizedWeight,
							Goal:   cfg.Goal,
						}
						if cfg.Goal != "" {
							goalsFull[name] = cfg.Goal
						}
					}
					result.CriteriaParameters = criteriaFull
					if len(result.Goals) == 0 && len(goalsFull) > 0 {
						result.Goals = goalsFull
					}
				}
			}
		}
	} else {
		result, err = h.formTeamsWithDynamicCriteria(c.Context(), req, candidates)
	}
	if err != nil {
		return nil, err
	}
	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata["criteria_parameters"] = req.CriteriaParameters
	result.Metadata["algorithm_used"] = algorithm
	result.Metadata["mode"] = "dynamic_criteria"
	return result, nil
}

// formTeamsWithDynamicCriteria forma equipos usando criteria_parameters dinámicos
func (h *TeamFormationHandler) formTeamsWithDynamicCriteria(ctx context.Context, req FlexibleFormTeamsRequest, candidates []core.TeamMember) (*core.TeamFormationResult, error) {
	log.Printf("[formTeamsWithDynamicCriteria] FALLBACK MODE - using scalar engine")
	requestDTO := h.convertToDTO(req, candidates)

	teamMapper := mapper.NewTeamFormationMapper()
	students, config, factory, evalContext, err := teamMapper.MapToDomain(requestDTO)
	if err != nil {
		return nil, fmt.Errorf("failed to map request to domain: %w", err)
	}

	// Propagate equity strategy from the HTTP request into the domain config so
	// CalculateFitness uses minimax aggregation during the genetic run.
	// The user's per-criterion targets are never modified here.
	log.Printf("[equity] equity_strategy received from request: %s", req.EquityStrategy)
	config.EquityStrategy = req.EquityStrategy

	bestChromosome, bestStats, err := runMultiRunGenetic(ctx, students, *config, factory, evalContext, requestDTO.TeamSize)
	if err != nil {
		return nil, err
	}

	return h.convertGeneticResultToTeamFormation(bestChromosome, bestStats, req.SectionID, req.CriteriaParameters, factory, config, evalContext), nil
}

// runMultiRunGenetic executes the genetic algorithm up to maxRuns times and returns the best result.
func runMultiRunGenetic(ctx context.Context, students []*domain.Student, config domain.Config, factory *genetic.CriterionFactory, evalContext domain.Context, targetGroupSize int) (*domain.Chromosome, *genetic.EvolutionStats, error) {
	// When EquityStrategy is "minimax" the fitness is the worst-team score (minimax).
	// The theoretical maximum equals the best individual-team score achievable
	// given the user's targets — typically well below 1.0 when targets create
	// inherent trade-offs (e.g. target=0 with leftover students from multiple
	// faculties).  We lower the stopping thresholds so the engine can early-exit
	// instead of running all generations in vain chasing an unreachable ceiling.
	targetFitness := 0.98
	acceptableFitness := 0.90
	if config.EquityStrategy == "minimax" {
		targetFitness = 0.75
		acceptableFitness = 0.60
	}

	baseConfig := genetic.EngineConfig{
		PopulationSize:  60,
		Generations:     40,
		ElitismRate:     0.15,
		MutationRate:    0.2,
		CrossoverRate:   0.8,
		TournamentSize:  5,
		TargetGroupSize: targetGroupSize,
		MaxStagnation:   15,
		TargetFitness:   targetFitness,
	}

	const maxRuns = 1
	var bestChromosome *domain.Chromosome
	var bestStats *genetic.EvolutionStats
	bestFitness := -1.0

	for run := 1; run <= maxRuns; run++ {
		cfg := baseConfig
		cfg.Seed = time.Now().UnixNano() + int64(run*1000)
		engine := genetic.NewEngine(cfg, factory, config, evalContext)
		chromosome, stats, err := engine.Run(ctx, students)
		if err != nil {
			return nil, nil, fmt.Errorf("genetic algorithm execution failed on run %d: %w", run, err)
		}
		if chromosome != nil && chromosome.Fitness > bestFitness {
			bestFitness = chromosome.Fitness
			bestChromosome = chromosome
			bestStats = stats
		}
		if bestFitness >= acceptableFitness {
			markEarlyExit(bestChromosome, run)
			break
		}
		if run < maxRuns {
			log.Printf("Run %d: fitness %.4f below threshold %.4f, retrying...", run, bestFitness, acceptableFitness)
		}
	}

	if bestChromosome != nil {
		if bestChromosome.Stats == nil {
			bestChromosome.Stats = make(map[string]interface{})
		}
		if _, exists := bestChromosome.Stats["runs_executed"]; !exists {
			bestChromosome.Stats["runs_executed"] = maxRuns
			bestChromosome.Stats["early_exit"] = false
		}
		bestChromosome.Stats["final_fitness"] = bestFitness
	}
	return bestChromosome, bestStats, nil
}

// markEarlyExit stamps early-exit metadata on the chromosome.
func markEarlyExit(chromosome *domain.Chromosome, run int) {
	if chromosome == nil {
		return
	}
	if chromosome.Stats == nil {
		chromosome.Stats = make(map[string]interface{})
	}
	chromosome.Stats["runs_executed"] = run
	chromosome.Stats["early_exit"] = true
}

// convertToDTO convierte el FlexibleFormTeamsRequest HTTP a un DTO estándar
func (h *TeamFormationHandler) convertToDTO(req FlexibleFormTeamsRequest, candidates []core.TeamMember) *dto.TeamFormationRequestDTO {
	// Extraer IDs de candidatos
	candidateIDs := make([]string, len(req.Candidates))
	for i, id := range req.Candidates {
		candidateIDs[i] = id
	}

	// Convertir CriteriaParameters
	criteriaParams := make(map[string]dto.CriterionParameterDTO)
	for name, config := range req.CriteriaParameters {
		criteriaParams[name] = dto.CriterionParameterDTO{
			Target: config.Target,
			Weight: config.Weight,
			Goal:   config.Goal,
			Type:   config.Type,
			Label:  config.Label,
		}
	}

	// Convertir DemographicData
	demographicData := make(map[string]dto.StudentAttributesDTO)
	for studentID, attrs := range req.DemographicData {
		demographicData[studentID] = dto.StudentAttributesDTO{
			NumericAttrs:      attrs.NumericAttrs,
			CategoricalAttrs:  attrs.CategoricalAttrs,
			SocialConnections: attrs.SocialConnections,
		}
	}

	return &dto.TeamFormationRequestDTO{
		SectionID:          req.SectionID,
		Candidates:         candidateIDs,
		TeamSize:           req.TeamSize,
		MinTeamSize:        req.MinTeamSize,
		MaxTeamSize:        req.MaxTeamSize,
		Algorithm:          req.Algorithm,
		CriteriaParameters: criteriaParams,
		DemographicData:    demographicData,
		EquityStrategy:     req.EquityStrategy,
	}
}

// convertGeneticResultToTeamFormation convierte el resultado del algoritmo genético al formato de la API
func (h *TeamFormationHandler) convertGeneticResultToTeamFormation(bestChromosome *domain.Chromosome, stats *genetic.EvolutionStats, sectionID string, criteriaParams map[string]CriterionConfig, factory *genetic.CriterionFactory, config *domain.Config, evalContext domain.Context) *core.TeamFormationResult { //nolint:cyclop
	if bestChromosome == nil {
		return &core.TeamFormationResult{
			Teams: []core.Team{}, GlobalScore: 0, Algorithm: "genetic_dynamic",
			ProvidersUsed: []string{}, Weights: map[string]float64{},
			Targets: map[string]float64{}, ExecutionTimeMs: 0,
			Metadata: map[string]interface{}{"error": "no valid solution found"},
		}
	}

	// Normalize weights so they sum to 1.0 — matches what the genetic engine actually used.
	totalWeight := 0.0
	for _, cfg := range criteriaParams {
		if cfg.Weight > 0 {
			totalWeight += cfg.Weight
		}
	}
	weights := make(map[string]float64, len(criteriaParams))
	targets := make(map[string]float64, len(criteriaParams))
	goals := make(map[string]string, len(criteriaParams))
	criteriaFull := make(map[string]core.CriterionConfig, len(criteriaParams))
	for criterionName, cfg := range criteriaParams {
		normalizedWeight := 0.0
		if totalWeight > 0 {
			normalizedWeight = cfg.Weight / totalWeight
		}
		weights[criterionName] = normalizedWeight
		targets[criterionName] = cfg.Target
		if cfg.Goal != "" {
			goals[criterionName] = cfg.Goal
		}
		criteriaFull[criterionName] = core.CriterionConfig{
			Target: cfg.Target,
			Weight: normalizedWeight,
			Goal:   cfg.Goal,
		}
	}

	teams := make([]core.Team, 0, len(bestChromosome.Groups))
	totalScore := 0.0
	for i, group := range bestChromosome.Groups {
		members := make([]core.TeamMember, 0, len(group))
		for _, student := range group {
			members = append(members, core.TeamMember{ID: student.ID})
		}
		groupCopy := make(domain.Group, len(group))
		copy(groupCopy, group)
		teamScore := genetic.CalculateFitnessForGroup(&groupCopy, *config, factory, evalContext)
		totalScore += teamScore

		// Calculate per-criterion scores and build both ProviderScores (flat map for
		// backward compat) and CriterionBreakdown (detailed desglose for the UI / DB).
		providerScores := make(map[string]float64, len(criteriaParams))
		breakdown := make(map[string]core.CriterionBreakdownEntry, len(criteriaParams))
		for criterionName, cfg := range criteriaParams {
			criterion := factory.Get(criterionName)
			if criterion == nil {
				continue
			}
			score := criterion.Calculate(&groupCopy, evalContext)
			providerScores[criterionName] = score

			// RealValue: raw underlying metric independent of fitness transformation.
			// - CategoricalCriterion → normalised Blau index via CalculateRealValue.
			// - NumericCriterion     → mean (target/minimize/maximize) or variance
			//                          (homogeneity/heterogeneity) via CalculateRealValue.
			// - Other types          → fall back to the score value.
			realValue := score
			if catCrit, ok := criterion.(*genetic.CategoricalCriterion); ok {
				realValue = catCrit.CalculateRealValue(&groupCopy, evalContext)
			} else if numCrit, ok := criterion.(*genetic.NumericCriterion); ok {
				realValue = numCrit.CalculateRealValue(&groupCopy, evalContext)
			}

			breakdown[criterionName] = core.CriterionBreakdownEntry{
				Score:     score,
				RealValue: realValue,
				Target:    cfg.Target,
				Weight:    weights[criterionName],
				Goal:      cfg.Goal,
			}
		}

		teams = append(teams, core.Team{
			ID: fmt.Sprintf("team_%d", i+1), Members: members, AffinityScore: teamScore,
			Metadata: core.TeamMetadata{
				ProviderScores:     providerScores,
				TotalComparisons:   len(members) * (len(members) - 1) / 2,
				IsFeasible:         true,
				CriterionBreakdown: breakdown,
			},
		})
	}

	// globalScore: when EquityStrategy is "minimax" we keep the chromosome's own
	// fitness value, which was computed by CalculateFitness with the variance
	// penalty applied.  Recomputing a plain average here would silently discard
	// the equity adjustment and report a misleadingly high score to the caller.
	// With EquityStrategy="" (default) we still use the plain average for backward compatibility.
	globalScore := bestChromosome.Fitness
	if len(teams) > 0 && config.EquityStrategy == "" {
		globalScore = totalScore / float64(len(teams))
	}

	var executionTimeMs int64
	if stats != nil {
		executionTimeMs = stats.TotalTime.Milliseconds()
	}

	metadata := buildGeneticResultMetadata(stats, bestChromosome)
	enrichMetadataWithEquityStats(metadata, bestChromosome, config, factory, evalContext)

	providersUsed := make([]string, 0, len(criteriaParams))
	for criterionName := range criteriaParams {
		providersUsed = append(providersUsed, criterionName)
	}
	if len(providersUsed) == 0 {
		providersUsed = []string{"dynamic_criteria"}
	}

	return &core.TeamFormationResult{
		Teams: teams, GlobalScore: globalScore, Algorithm: "genetic_dynamic",
		ProvidersUsed: providersUsed, Weights: weights, Targets: targets, Goals: goals,
		CriteriaParameters: criteriaFull,
		EquityStrategy:     config.EquityStrategy,
		ExecutionTimeMs:    executionTimeMs, Metadata: metadata,
	}
}

// buildGeneticResultMetadata assembles the metadata map for a genetic algorithm result.
func buildGeneticResultMetadata(stats *genetic.EvolutionStats, chromosome *domain.Chromosome) map[string]interface{} {
	metadata := map[string]interface{}{
		"best_fitness": chromosome.Fitness,
		"algorithm":    "genetic_dynamic",
	}
	if stats != nil {
		metadata["generations_run"] = len(stats.Generations)
		metadata["total_time_ms"] = stats.TotalTime.Milliseconds()
		metadata["final_diversity"] = getFinalDiversity(stats)
		metadata["converged"] = stats.Converged
		metadata["stop_reason"] = stats.StopReason
	}
	if chromosome.Stats != nil {
		metadata["chromosome_stats"] = chromosome.Stats
		if runs, ok := chromosome.Stats["runs_executed"]; ok {
			metadata["runs_executed"] = runs
		}
		if earlyExit, ok := chromosome.Stats["early_exit"]; ok {
			metadata["early_exit"] = earlyExit
		}
		if finalFitness, ok := chromosome.Stats["final_fitness"]; ok {
			metadata["final_fitness"] = finalFitness
		}
	}
	return metadata
}

// enrichMetadataWithEquityStats appends equity audit fields to the metadata map.
// It must be called after the chromosome has been fully evaluated so that per-team
// scores are deterministic (same weights/factory/context used during the run).
func enrichMetadataWithEquityStats(metadata map[string]interface{}, chromosome *domain.Chromosome, config *domain.Config, factory *genetic.CriterionFactory, ctx domain.Context) {
	if chromosome == nil || config == nil || factory == nil {
		return
	}
	eq := genetic.ComputeEquityStats(chromosome, *config, factory, ctx)
	metadata["equity_enforced"] = eq.EquityEnforced
	metadata["lowest_team_score"] = eq.LowestTeamScore
	metadata["highest_team_score"] = eq.HighestTeamScore
	metadata["score_std_dev"] = eq.ScoreStdDev
	metadata["score_variance"] = eq.ScoreVariance
}

// getFinalDiversity obtiene la diversidad final de las estadísticas
func getFinalDiversity(stats *genetic.EvolutionStats) float64 {
	if stats == nil || len(stats.Generations) == 0 {
		return 0.0
	}
	lastGen := stats.Generations[len(stats.Generations)-1]
	return lastGen.Diversity
}

// buildProviderWeights construye el mapa de pesos basado en la configuración del usuario
func (h *TeamFormationHandler) buildProviderWeights(config *ProviderConfigRequest) (map[string]float64, error) {
	if config == nil {
		return map[string]float64{"enneagram": 0.3, "sociogram": 0.7}, nil
	}
	if len(config.CustomProviders) > 0 {
		return h.buildFromCustomProviders(config.CustomProviders)
	}
	weights := make(map[string]float64)
	if err := h.applyPersonalityWeight(config, weights); err != nil {
		return nil, err
	}
	if err := h.applySociogramWeight(config, weights); err != nil {
		return nil, err
	}
	return weights, nil
}

// buildFromCustomProviders validates and returns weights from a custom map.
func (h *TeamFormationHandler) buildFromCustomProviders(custom map[string]float64) (map[string]float64, error) {
	weights := make(map[string]float64)
	for provider, weight := range custom {
		if weight < 0 || weight > 1 {
			return nil, fmt.Errorf("weight for provider '%s' must be between 0 and 1", provider)
		}
		if !h.registry.HasProvider(provider) {
			return nil, fmt.Errorf("unknown provider: %s", provider)
		}
		if weight > 0 {
			weights[provider] = weight
		}
	}
	return weights, nil
}

// applyPersonalityWeight adds a personality test weight to weights if configured.
func (h *TeamFormationHandler) applyPersonalityWeight(config *ProviderConfigRequest, weights map[string]float64) error {
	if config.PersonalityTest == "" {
		return nil
	}
	validTests := map[string]bool{"enneagram": true, "mbti": true}
	if !validTests[config.PersonalityTest] {
		return fmt.Errorf("invalid personality test: %s (valid: enneagram, mbti)", config.PersonalityTest)
	}
	if !h.registry.HasProvider(config.PersonalityTest) {
		return fmt.Errorf("personality test '%s' is not available", config.PersonalityTest)
	}
	weight := config.PersonalityWeight
	if weight <= 0 {
		weight = 0.5
	}
	if weight > 1 {
		return fmt.Errorf("personality_weight must be between 0 and 1")
	}
	weights[config.PersonalityTest] = weight
	return nil
}

// applySociogramWeight adds the sociogram weight to weights if configured.
func (h *TeamFormationHandler) applySociogramWeight(config *ProviderConfigRequest, weights map[string]float64) error {
	if !config.UseSociogram {
		return nil
	}
	if !h.registry.HasProvider("sociogram") {
		return fmt.Errorf("sociogram provider is not available")
	}
	weight := config.SociogramWeight
	if weight <= 0 {
		weight = 0.5
	}
	if weight > 1 {
		return fmt.Errorf("sociogram_weight must be between 0 and 1")
	}
	weights["sociogram"] = weight
	return nil
}

// GetProviders retorna información de los providers disponibles
// @Summary Obtener providers disponibles
// @Description Lista todos los AffinityProviders registrados y sus pesos
// @Tags team-formation
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/team-formation/providers [get]
func (h *TeamFormationHandler) GetProviders(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data":    h.registry.Summary(),
	})
}

// UpdateWeights actualiza los pesos de los providers
// @Summary Actualizar pesos de providers
// @Description Modifica los pesos de ponderación para la formación de equipos
// @Tags team-formation
// @Accept json
// @Produce json
// @Param request body UpdateWeightsRequest true "Nuevos pesos"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/team-formation/weights [put]
func (h *TeamFormationHandler) UpdateWeights(c *fiber.Ctx) error {
	var req UpdateWeightsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   errInvalidRequest,
			"details": err.Error(),
		})
	}

	// Validar que los pesos estén en rango
	var total float64
	for _, w := range req.Weights {
		if w < 0 || w > 1 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Weights must be between 0 and 1",
			})
		}
		total += w
	}

	// Actualizar pesos en el engine
	h.engine.UpdateWeights(req.Weights)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Weights updated successfully",
		"data": fiber.Map{
			"weights":      req.Weights,
			"total_weight": total,
		},
	})
}

// CalculateAffinity calcula la afinidad entre dos personas
// @Summary Calcular afinidad entre dos personas
// @Description Retorna la afinidad combinada ponderada entre dos personas
// @Tags team-formation
// @Produce json
// @Param section_id query string true "ID de la sección"
// @Param person_a query string true "ID de persona A"
// @Param person_b query string true "ID de persona B"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/team-formation/affinity [get]
func (h *TeamFormationHandler) CalculateAffinity(c *fiber.Ctx) error {
	sectionID := c.Query("section_id")
	personA := c.Query("person_a")
	personB := c.Query("person_b")

	if sectionID == "" || personA == "" || personB == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "section_id, person_a and person_b are required",
		})
	}

	// Crear contexto con section_id
	ctx := context.WithValue(c.Context(), core.ContextKeySectionID, sectionID)

	score, providerScores, err := h.engine.CalculateCombinedAffinity(ctx, personA, personB)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to calculate affinity",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"person_a":        personA,
			"person_b":        personB,
			"combined_score":  score,
			"provider_scores": providerScores,
		},
	})
}

// GetAvailableOptions retorna las opciones disponibles para formación de equipos
// @Summary Obtener opciones disponibles
// @Description Lista los tests de personalidad y otras opciones disponibles para formar equipos
// @Tags team-formation
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/team-formation/options [get]
func (h *TeamFormationHandler) GetAvailableOptions(c *fiber.Ctx) error {
	// Obtener providers disponibles
	providers := h.registry.Summary()

	// Construir lista de tests de personalidad disponibles
	personalityTests := make([]map[string]interface{}, 0)
	hasSociogram := false

	if eligible, ok := providers["eligible_providers"].([]map[string]interface{}); ok {
		for _, p := range eligible {
			providerID := p["id"].(string)
			category := ""
			if cat, ok := p["category"].(string); ok {
				category = cat
			}

			if category == "personality" {
				personalityTests = append(personalityTests, map[string]interface{}{
					"id":          providerID,
					"name":        p["name"],
					"description": getPersonalityTestDescription(providerID),
				})
			} else if providerID == "sociogram" {
				hasSociogram = true
			}
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"personality_tests":   personalityTests,
			"sociogram_available": hasSociogram,
			"algorithms": []map[string]interface{}{
				{
					"id":          "genetic",
					"name":        "Algoritmo Genético",
					"description": "Optimización multi-objetivo usando evolución (recomendado)",
					"default":     true,
				},
				{
					"id":          "greedy",
					"name":        "Algoritmo Voraz",
					"description": "Formación rápida basada en mejor siguiente opción",
					"default":     false,
				},
			},
			"default_config": fiber.Map{
				"personality_test":   "enneagram",
				"personality_weight": 0.3,
				"use_sociogram":      true,
				"sociogram_weight":   0.7,
				"algorithm":          "genetic",
			},
		},
	})
}

// EvaluateGroupsRequest representa una solicitud de evaluación de grupos ya formados.
// No ejecuta el algoritmo genético; solo calcula los scores con los criterios dados.
//
// SectionCandidates es opcional: cuando se proporciona, se usa como población completa
// para calcular CategoryCounts (kGlobal del índice de Blau), replicando exactamente
// el comportamiento de form-flexible que siempre opera sobre toda la sección.
// Si se omite, se usa el conjunto de miembros de los grupos enviados, lo que puede
// producir un kGlobal menor y scores categóricos distintos a los de form-flexible.
type EvaluateGroupsRequest struct {
	SectionID          string                       `json:"section_id"`
	Groups             []EvaluateGroupItem          `json:"groups"`
	CriteriaParameters map[string]CriterionConfig   `json:"criteria_parameters"`
	DemographicData    map[string]StudentAttributes `json:"demographic_data"`
	// SectionCandidates lista todos los IDs de candidatos de la sección (no sólo los agrupados).
	// Úsalo para que el kGlobal del índice de Blau coincida con el calculado por form-flexible.
	SectionCandidates []string `json:"section_candidates,omitempty"`
	// EquityStrategy controla la agregación del GlobalScore: "minimax" = min(team_scores), "" = mean(team_scores).
	EquityStrategy string `json:"equity_strategy,omitempty"`
}

type EvaluateGroupItem struct {
	GroupID   string   `json:"group_id"`
	MemberIDs []string `json:"member_ids"`
}

type EvaluateGroupsResponse struct {
	Groups      []EvaluatedGroup `json:"groups"`
	GlobalScore float64          `json:"global_score"`
}

type EvaluatedGroup struct {
	GroupID            string                                  `json:"group_id"`
	AffinityScore      float64                                 `json:"affinity_score"`
	ProviderScores     map[string]float64                      `json:"provider_scores"`
	CriterionBreakdown map[string]core.CriterionBreakdownEntry `json:"criterion_breakdown,omitempty"`
}

// EvaluateGroups recalcula los scores de grupos ya formados sin ejecutar el algoritmo genético
// y sin persistir ningún resultado. Es un endpoint de solo cómputo.
func (h *TeamFormationHandler) EvaluateGroups(c *fiber.Ctx) error {
	var req EvaluateGroupsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": fmt.Errorf("invalid request body: %w", err).Error(),
		})
	}

	if req.SectionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "section_id is required",
		})
	}
	if len(req.Groups) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "at least one group is required",
		})
	}
	if len(req.CriteriaParameters) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "criteria_parameters is required",
		})
	}

	// Recopilar todos los candidatos de todos los grupos para construir el DTO del mapper.
	allCandidates := make([]string, 0)
	for _, g := range req.Groups {
		allCandidates = append(allCandidates, g.MemberIDs...)
	}

	// Convertir CriterionConfig → dto.CriterionParameterDTO
	criteriaDTO := make(map[string]dto.CriterionParameterDTO, len(req.CriteriaParameters))
	for name, cfg := range req.CriteriaParameters {
		criteriaDTO[name] = dto.CriterionParameterDTO{
			Target: cfg.Target,
			Weight: cfg.Weight,
			Goal:   cfg.Goal,
			Type:   cfg.Type,
			Label:  cfg.Label,
		}
	}

	// Convertir StudentAttributes → dto.StudentAttributesDTO
	demographicDTO := make(map[string]dto.StudentAttributesDTO, len(req.DemographicData))
	for studentID, attrs := range req.DemographicData {
		demographicDTO[studentID] = dto.StudentAttributesDTO{
			NumericAttrs:      attrs.NumericAttrs,
			CategoricalAttrs:  attrs.CategoricalAttrs,
			SocialConnections: attrs.SocialConnections,
		}
	}

	// Población completa para calcular CategoryCounts (kGlobal del índice de Blau).
	// Si el caller envía section_candidates usamos esa lista; en caso contrario caemos
	// al conjunto de miembros agrupados, que puede ser incompleto y producir un kGlobal
	// menor al que usó form-flexible, causando divergencias en scores categóricos.
	populationCandidates := allCandidates
	if len(req.SectionCandidates) > 0 {
		populationCandidates = req.SectionCandidates
	}

	requestDTO := &dto.TeamFormationRequestDTO{
		SectionID:          req.SectionID,
		Candidates:         populationCandidates,
		TeamSize:           2, // valor mínimo requerido por Validate(); no afecta la evaluación
		CriteriaParameters: criteriaDTO,
		DemographicData:    demographicDTO,
	}

	teamMapper := mapper.NewTeamFormationMapper()
	_, config, factory, evalContext, err := teamMapper.MapToDomain(requestDTO)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": fmt.Errorf("failed to build evaluation context: %w", err).Error(),
		})
	}

	// Construir mapa de estudiantes por ID para armar los domain.Group de cada equipo.
	studentByID := make(map[string]*domain.Student, len(allCandidates))
	for candidateID := range demographicDTO {
		attrs := demographicDTO[candidateID]
		s := &domain.Student{
			ID:               candidateID,
			NumericAttrs:     attrs.NumericAttrs,
			CategoricalAttrs: attrs.CategoricalAttrs,
		}
		if s.NumericAttrs == nil {
			s.NumericAttrs = make(map[string]float64)
		}
		if s.CategoricalAttrs == nil {
			s.CategoricalAttrs = make(map[string]string)
		}
		studentByID[candidateID] = s
	}

	// Normalizar pesos para usarlos en el CriterionBreakdown.
	evalTotalWeight := 0.0
	for _, cfg := range req.CriteriaParameters {
		if cfg.Weight > 0 {
			evalTotalWeight += cfg.Weight
		}
	}
	evalWeights := make(map[string]float64, len(req.CriteriaParameters))
	for name, cfg := range req.CriteriaParameters {
		if evalTotalWeight > 0 {
			evalWeights[name] = cfg.Weight / evalTotalWeight
		}
	}

	// Evaluar cada grupo individualmente usando la función de fitness del motor genético.
	evaluatedGroups := make([]EvaluatedGroup, 0, len(req.Groups))
	totalScore := 0.0

	for _, g := range req.Groups {
		group := make(domain.Group, 0, len(g.MemberIDs))
		for _, memberID := range g.MemberIDs {
			if s, ok := studentByID[memberID]; ok {
				group = append(group, s)
			} else {
				// Estudiante sin datos demográficos: incluirlo con atributos vacíos.
				group = append(group, &domain.Student{
					ID:               memberID,
					NumericAttrs:     make(map[string]float64),
					CategoricalAttrs: make(map[string]string),
				})
			}
		}

		groupScore := genetic.CalculateFitnessForGroup(&group, *config, factory, evalContext)

		// Calcular score por criterio individual para mostrar detalle en el frontend.
		providerScores := make(map[string]float64, len(req.CriteriaParameters))
		breakdown := make(map[string]core.CriterionBreakdownEntry, len(req.CriteriaParameters))
		for criterionName, cfg := range req.CriteriaParameters {
			criterion := factory.Get(criterionName)
			if criterion == nil {
				continue
			}
			score := criterion.Calculate(&group, evalContext)
			providerScores[criterionName] = score

			// RealValue: raw underlying metric independent of fitness transformation.
			// - CategoricalCriterion → normalised Blau index via CalculateRealValue.
			// - NumericCriterion     → mean (target/minimize/maximize) or variance
			//                          (homogeneity/heterogeneity) via CalculateRealValue.
			// - Other types          → fall back to the score value.
			realValue := score
			if catCrit, ok := criterion.(*genetic.CategoricalCriterion); ok {
				realValue = catCrit.CalculateRealValue(&group, evalContext)
			} else if numCrit, ok := criterion.(*genetic.NumericCriterion); ok {
				realValue = numCrit.CalculateRealValue(&group, evalContext)
			}

			breakdown[criterionName] = core.CriterionBreakdownEntry{
				Score:     score,
				RealValue: realValue,
				Target:    cfg.Target,
				Weight:    evalWeights[criterionName],
				Goal:      cfg.Goal,
			}
		}

		totalScore += groupScore
		evaluatedGroups = append(evaluatedGroups, EvaluatedGroup{
			GroupID:            g.GroupID,
			AffinityScore:      groupScore,
			ProviderScores:     providerScores,
			CriterionBreakdown: breakdown,
		})
	}

	// GlobalScore: replicar la misma estrategia de agregación que usa form-flexible.
	// - EquityStrategy="minimax" → minimax: score del peor equipo (igual que CalculateFitness).
	// - EquityStrategy=""        → media aritmética de los scores individuales.
	globalScore := 0.0
	if len(evaluatedGroups) > 0 {
		if req.EquityStrategy == "minimax" {
			worst := evaluatedGroups[0].AffinityScore
			for _, eg := range evaluatedGroups[1:] {
				if eg.AffinityScore < worst {
					worst = eg.AffinityScore
				}
			}
			globalScore = worst
		} else {
			globalScore = totalScore / float64(len(evaluatedGroups))
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": EvaluateGroupsResponse{
			Groups:      evaluatedGroups,
			GlobalScore: globalScore,
		},
	})
}

// getPersonalityTestDescription retorna una descripción para cada test de personalidad
func getPersonalityTestDescription(testID string) string {
	descriptions := map[string]string{
		"enneagram": "Sistema de 9 tipos de personalidad basado en motivaciones y miedos",
		"mbti":      "Myers-Briggs Type Indicator - 16 tipos de personalidad basados en preferencias cognitivas",
	}
	if desc, ok := descriptions[testID]; ok {
		return desc
	}
	return "Test de personalidad"
}
