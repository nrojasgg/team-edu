package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"ms-core/internal/adapters/dto"
	"ms-core/internal/adapters/mapper"
	"ms-core/internal/core"
	"ms-core/internal/core/domain"
	"ms-core/internal/core/genetic"
)

// runLegacyFormationCtx is the context-pure equivalent of runLegacyFormation.
// It is used by the async worker path where no *fiber.Ctx is available.
func (h *TeamFormationHandler) runLegacyFormationCtx(ctx context.Context, req FlexibleFormTeamsRequest, candidates []core.TeamMember, algorithm string) (*core.TeamFormationResult, error) {
	providerWeights, err := h.buildProviderWeights(req.ProviderConfig)
	if err != nil {
		return nil, fmt.Errorf("invalid provider configuration: %w", err)
	}
	if len(providerWeights) == 0 {
		return nil, fmt.Errorf("at least one affinity provider must be selected")
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
		enrichedCtx := context.WithValue(ctx, core.ContextKeySectionID, req.SectionID)
		affinityMatrix := h.engine.PrecomputeAffinityMatrix(enrichedCtx, candidates)
		providerIDs := make([]string, 0, len(providerWeights))
		for id := range providerWeights {
			providerIDs = append(providerIDs, id)
		}
		pluginReq := teamFormationPluginReq{
			Mode:            "affinity",
			SectionID:       req.SectionID,
			Candidates:      req.Candidates,
			TeamSize:        req.TeamSize,
			MinTeamSize:     req.MinTeamSize,
			MaxTeamSize:     req.MaxTeamSize,
			Algorithm:       algorithm,
			AffinityMatrix:  affinityMatrix,
			ProviderIDs:     providerIDs,
			ProviderWeights: providerWeights,
		}
		raw, pluginErr := h.tfPlugin.Execute(ctx, pluginReq)
		if pluginErr != nil {
			log.Printf("team_formation plugin error (legacy ctx mode), falling back: %v", pluginErr)
			result, err = h.engine.FormTeamsWithConfig(ctx, formationRequest)
		} else {
			result, err = marshalToTeamFormationResult(raw)
		}
	} else {
		result, err = h.engine.FormTeamsWithConfig(ctx, formationRequest)
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

// runCriteriaFormationCtx is the context-pure equivalent of runCriteriaFormation.
// It is used by the async worker path where no *fiber.Ctx is available.
func (h *TeamFormationHandler) runCriteriaFormationCtx(ctx context.Context, req FlexibleFormTeamsRequest, candidates []core.TeamMember, algorithm string) (*core.TeamFormationResult, error) {
	h.enrichWithOCEANData(ctx, &req)
	requestDTO := h.convertToDTO(req, candidates)

	var (
		result *core.TeamFormationResult
		err    error
	)

	if h.tfPlugin != nil {
		pluginReq := teamFormationPluginReq{
			Mode:            "criteria",
			TeamSize:        req.TeamSize,
			MinTeamSize:     req.MinTeamSize,
			MaxTeamSize:     req.MaxTeamSize,
			Algorithm:       algorithm,
			CriteriaRequest: requestDTO,
			EquityStrategy:  req.EquityStrategy,
		}
		raw, pluginErr := h.tfPlugin.Execute(ctx, pluginReq)
		if pluginErr != nil {
			log.Printf("team_formation plugin error (criteria ctx mode), falling back: %v", pluginErr)
			result, err = h.formTeamsWithDynamicCriteria(ctx, req, candidates)
		} else {
			result, err = marshalToTeamFormationResult(raw)
			if err == nil && result != nil && len(result.CriteriaParameters) == 0 {
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
	} else {
		result, err = h.formTeamsWithDynamicCriteria(ctx, req, candidates)
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

// evaluateGroupsCore contains all the EvaluateGroups business logic without any
// dependency on *fiber.Ctx. It is shared by both the sync HTTP handler and the
// async worker.
func (h *TeamFormationHandler) evaluateGroupsCore(ctx context.Context, req EvaluateGroupsRequest) (*EvaluateGroupsResponse, error) {
	allCandidates := make([]string, 0)
	for _, g := range req.Groups {
		allCandidates = append(allCandidates, g.MemberIDs...)
	}

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

	demographicDTO := make(map[string]dto.StudentAttributesDTO, len(req.DemographicData))
	for studentID, attrs := range req.DemographicData {
		demographicDTO[studentID] = dto.StudentAttributesDTO{
			NumericAttrs:      attrs.NumericAttrs,
			CategoricalAttrs:  attrs.CategoricalAttrs,
			SocialConnections: attrs.SocialConnections,
		}
	}

	populationCandidates := allCandidates
	if len(req.SectionCandidates) > 0 {
		populationCandidates = req.SectionCandidates
	}

	requestDTO := &dto.TeamFormationRequestDTO{
		SectionID:          req.SectionID,
		Candidates:         populationCandidates,
		TeamSize:           2,
		CriteriaParameters: criteriaDTO,
		DemographicData:    demographicDTO,
	}

	teamMapper := mapper.NewTeamFormationMapper()
	_, config, factory, evalContext, err := teamMapper.MapToDomain(requestDTO)
	if err != nil {
		return nil, fmt.Errorf("failed to build evaluation context: %w", err)
	}

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

	evaluatedGroups := make([]EvaluatedGroup, 0, len(req.Groups))
	totalScore := 0.0

	for _, g := range req.Groups {
		group := make(domain.Group, 0, len(g.MemberIDs))
		for _, memberID := range g.MemberIDs {
			if s, ok := studentByID[memberID]; ok {
				group = append(group, s)
			} else {
				group = append(group, &domain.Student{
					ID:               memberID,
					NumericAttrs:     make(map[string]float64),
					CategoricalAttrs: make(map[string]string),
				})
			}
		}

		groupScore := genetic.CalculateFitnessForGroup(&group, *config, factory, evalContext)

		providerScores := make(map[string]float64, len(req.CriteriaParameters))
		breakdown := make(map[string]core.CriterionBreakdownEntry, len(req.CriteriaParameters))
		for criterionName, cfg := range req.CriteriaParameters {
			criterion := factory.Get(criterionName)
			if criterion == nil {
				continue
			}
			score := criterion.Calculate(&group, evalContext)
			providerScores[criterionName] = score

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

	// Suppress unused-import warning: ctx is intentionally accepted for future
	// deadline propagation into the mapper / genetic engine.
	_ = ctx

	return &EvaluateGroupsResponse{
		Groups:      evaluatedGroups,
		GlobalScore: globalScore,
	}, nil
}

// extractTeamFormationResult performs a safe type-assertion from the polymorphic
// interface{} returned by runLegacy/runCriteriaFormationCtx back to the concrete
// *core.TeamFormationResult used by saveGroupHistory.
func extractTeamFormationResult(v interface{}) *core.TeamFormationResult {
	if v == nil {
		return nil
	}
	if r, ok := v.(*core.TeamFormationResult); ok {
		return r
	}
	// Last resort: round-trip through JSON to handle any intermediate wrapper.
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var r core.TeamFormationResult
	if err := json.Unmarshal(b, &r); err != nil {
		return nil
	}
	return &r
}
