// Team Formation subprocess plugin for ms-core.
// Compile with:
//
//	go build -o ../../plugins_dist/team_formation .
//
// Launched as a subprocess by the host via hashicorp/go-plugin.
// Receives pre-computed affinity data from the kernel (no RPC back to other plugins).
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"ms-core/internal/adapters/dto"
	"ms-core/internal/adapters/mapper"
	"ms-core/internal/core"
	"ms-core/internal/core/domain"
	"ms-core/internal/core/genetic"
)

// ---------------------------------------------------------------------------
// Plugin entry point
// ---------------------------------------------------------------------------

func main() {
	core.ServePlugin(&TeamFormationPlugin{})
}

// ---------------------------------------------------------------------------
// Plugin implementation
// ---------------------------------------------------------------------------

// TeamFormationPlugin implementa core.Plugin y ejecuta el motor de optimización
// de formación de equipos como subproceso aislado.
// Soporta dos modos:
//   - "affinity": recibe una matriz de afinidad pre-calculada y ejecuta el algoritmo.
//   - "criteria": recibe datos demográficos + criterios dinámicos y corre el algoritmo genético.
type TeamFormationPlugin struct{}

func (p *TeamFormationPlugin) Name() string    { return "team_formation" }
func (p *TeamFormationPlugin) Version() string { return "1.0.0" }
func (p *TeamFormationPlugin) Initialize() error {
	log.Println("team_formation plugin initialized")
	return nil
}
func (p *TeamFormationPlugin) Shutdown() error { return nil }

// Execute recibe la solicitud de formación desde el kernel, ejecuta el algoritmo
// y devuelve el resultado. El kernel es responsable de recopilar los datos de
// afinidad; este plugin sólo realiza la optimización.
func (p *TeamFormationPlugin) Execute(ctx context.Context, data interface{}) (interface{}, error) {
	// data llega como map[string]interface{} (JSON deserializado por la capa RPC).
	// Re-serializamos para obtener una struct tipada.
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("team_formation: failed to marshal request: %w", err)
	}

	var req TeamFormationPluginRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, fmt.Errorf("team_formation: failed to unmarshal request: %w", err)
	}

	switch req.Mode {
	case "affinity":
		return p.handleAffinityMode(ctx, req)
	case "criteria":
		return p.handleCriteriaMode(ctx, req)
	default:
		return nil, fmt.Errorf("team_formation: unknown mode %q (expected 'affinity' or 'criteria')", req.Mode)
	}
}

// ---------------------------------------------------------------------------
// Mode: affinity — optimización sobre matriz pre-calculada
// ---------------------------------------------------------------------------

// handleAffinityMode computes relational cost matrix C'_ij from providers array
// and runs the optimization algorithm.
//
// According to the thesis formula:
//
//	C'_ij = Σ_p ω_p * (1 - S_p * T_p(i,j)) / 2
//	S_rel(G) = min(1, Σ_ij C'_ij / (|G|(|G|-1)/2))
//	f_2(x) = (1/m) * Σ_G S_rel(G)
func (p *TeamFormationPlugin) handleAffinityMode(ctx context.Context, req TeamFormationPluginRequest) (*core.TeamFormationResult, error) {
	if len(req.Candidates) == 0 {
		return nil, fmt.Errorf("team_formation affinity mode: candidates is required")
	}
	if len(req.Providers) == 0 && len(req.AffinityMatrix) == 0 {
		return nil, fmt.Errorf("team_formation affinity mode: either providers or affinity_matrix is required")
	}

	// Convertir IDs de candidatos a TeamMember
	candidates := make([]core.TeamMember, len(req.Candidates))
	for i, id := range req.Candidates {
		candidates[i] = core.TeamMember{ID: id}
	}

	// Compute C'_ij matrix from providers array using thesis formula
	costMatrix := computeRelationalCostMatrix(candidates, req.Providers)

	// Construir config del motor
	config := &core.TeamFormationConfig{
		Algorithm:     req.Algorithm,
		DefaultWeight: 1.0,
		Weights:       make(map[string]float64),
		GeneticParams: &core.GeneticAlgorithmParams{
			PopulationSize: req.GeneticPopSize,
			Generations:    req.GeneticGens,
			MutationRate:   req.GeneticMutation,
			ElitismRate:    0.1,
		},
	}
	if config.Algorithm == "" {
		config.Algorithm = "genetic"
	}
	if config.GeneticParams.PopulationSize == 0 {
		config.GeneticParams.PopulationSize = 100
	}
	if config.GeneticParams.Generations == 0 {
		config.GeneticParams.Generations = 50
	}
	if config.GeneticParams.MutationRate == 0 {
		config.GeneticParams.MutationRate = 0.1
	}

	request := core.TeamFormationRequest{
		SectionID:       req.SectionID,
		Candidates:      candidates,
		TeamSize:        req.TeamSize,
		MinTeamSize:     req.MinTeamSize,
		MaxTeamSize:     req.MaxTeamSize,
		Algorithm:       req.Algorithm,
		ProviderWeights: req.ProviderWeights,
	}

	// Añadir section_id al contexto
	ctx = context.WithValue(ctx, core.ContextKeySectionID, req.SectionID)

	result, err := core.OptimizeWithMatrix(ctx, config, request, costMatrix)
	if err != nil {
		return nil, fmt.Errorf("team_formation affinity mode: optimization failed: %w", err)
	}

	// Enriquecer metadata con info de providers
	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata["plugin"] = "team_formation"
	result.Metadata["mode"] = "affinity"
	if len(req.ProviderIDs) > 0 {
		result.ProvidersUsed = req.ProviderIDs
		result.Metadata["provider_weights"] = req.ProviderWeights
	}

	return result, nil
}

// computeRelationalCostMatrix computes C'_ij matrix from providers array.
// C'_ij = Σ_p ω_p * (1 - S_p * T_p(i,j)) / 2
//
// Example: If you have two providers:
//   - Eneagrama: T₁ with S₁=1 (harmony), ω₁=0.4
//   - Test Social: T₂ with S₂=-1 (controlled friction), ω₂=0.6
//
// Then C'_ij = 0.4*(1-1*T₁)/2 + 0.6*(1-(-1)*T₂)/2
func computeRelationalCostMatrix(candidates []core.TeamMember, providers []ProviderData) map[string]map[string]float64 {
	n := len(candidates)
	matrix := make(map[string]map[string]float64, n)
	for _, c := range candidates {
		matrix[c.ID] = make(map[string]float64, n)
	}

	// Compute total weight for normalization
	totalWeight := 0.0
	for _, p := range providers {
		if p.Weight > 0 {
			totalWeight += p.Weight
		}
	}
	if totalWeight == 0 {
		totalWeight = 1.0 // Avoid division by zero
	}

	// For each pair (i, j), compute C'_ij
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			idI := candidates[i].ID
			idJ := candidates[j].ID

			var totalCost float64
			for _, provider := range providers {
				if provider.Weight <= 0 {
					continue
				}
				// Get T_p(i,j) from provider's matrix
				Tp := getMatrixValue(provider.TpMatrix, idI, idJ, 0.0)
				// Clamp T_p to [-1, 1] as per thesis
				if Tp < -1 {
					Tp = -1
				}
				if Tp > 1 {
					Tp = 1
				}

				Sp := float64(provider.Intentionality) // S_p ∈ {-1, 1}
				Wp := provider.Weight / totalWeight    // Normalized weight

				// C'_ij component = ω_p * (1 - S_p * T_p(i,j)) / 2
				contribution := Wp * (1.0 - Sp*Tp) / 2.0
				totalCost += contribution
			}

			matrix[idI][idJ] = totalCost
			matrix[idJ][idI] = totalCost
		}
	}

	return matrix
}

// getMatrixValue safely retrieves a value from a matrix, returning defaultVal if not found.
func getMatrixValue(matrix map[string]map[string]float64, key1, key2 string, defaultVal float64) float64 {
	if row, ok := matrix[key1]; ok {
		if val, ok2 := row[key2]; ok2 {
			return val
		}
	}
	return defaultVal
}

// ---------------------------------------------------------------------------
// Mode: criteria — algoritmo genético con criterios dinámicos
// ---------------------------------------------------------------------------

// runGeneticWithSeed executes a single genetic run with the given seed and returns the best chromosome and stats.
func runGeneticWithSeed(ctx context.Context, seed int64, base genetic.EngineConfig, factory *genetic.CriterionFactory, config domain.Config, evalContext domain.Context, students []*domain.Student) (*domain.Chromosome, *genetic.EvolutionStats, error) {
	cfg := base
	cfg.Seed = seed
	engine := genetic.NewEngine(cfg, factory, config, evalContext)
	return engine.Run(ctx, students)
}

// markRunStats attaches number-of-runs metadata to the best chromosome.
func markRunStats(chromosome *domain.Chromosome, runsExecuted int, earlyExit bool, fitness float64) {
	if chromosome == nil {
		return
	}
	if chromosome.Stats == nil {
		chromosome.Stats = make(map[string]interface{})
	}
	if _, exists := chromosome.Stats["runs_executed"]; !exists {
		chromosome.Stats["runs_executed"] = runsExecuted
		chromosome.Stats["early_exit"] = earlyExit
	}
	chromosome.Stats["final_fitness"] = fitness
}

func (p *TeamFormationPlugin) handleCriteriaMode(ctx context.Context, req TeamFormationPluginRequest) (*core.TeamFormationResult, error) {
	if req.CriteriaRequest == nil {
		return nil, fmt.Errorf("team_formation criteria mode: criteria_request is required")
	}

	requestDTO := req.CriteriaRequest

	teamMapper := mapper.NewTeamFormationMapper()
	students, config, factory, evalContext, err := teamMapper.MapToDomain(requestDTO)
	if err != nil {
		return nil, fmt.Errorf("team_formation criteria mode: mapping failed: %w", err)
	}

	log.Printf("[CRITERIA] Received criteria_parameters: %v", requestDTO.CriteriaParameters)
	log.Printf("[CRITERIA] Config.Weights after mapping: %v", config.Weights)

	socialCriteria := teamMapper.GetSocialCriteriaNames()
	moConfig := genetic.SplitCriteriaByCategory(config.Weights, socialCriteria)

	log.Printf("[CRITERIA] SplitCriteriaByCategory result: HasPsy=%v, HasSoc=%v, PsyWeights=%v, SocWeights=%v",
		moConfig.HasPsy, moConfig.HasSoc, moConfig.PsyWeights, moConfig.SocWeights)

	if moConfig.HasPsy && moConfig.HasSoc {
		log.Printf("[MO] Detected mixed criteria: psychometric=%v, social=%v — routing to NSGA-II",
			moConfig.PsyWeights, moConfig.SocWeights)
		return p.handleMultiObjectiveCriteriaMode(ctx, req, students, config, factory, evalContext, moConfig)
	}

	log.Printf("[SCALAR] Single-objective criteria: hasPsy=%v, hasSoc=%v", moConfig.HasPsy, moConfig.HasSoc)

	// Propagate equity setting from the plugin request into the domain config so
	// CalculateFitness uses minimax aggregation during the genetic run.
	// requestDTO.EquityStrategy is the canonical source (set by convertToDTO in the
	// handler), but we also honour req.EquityStrategy as a belt-and-suspenders
	// fallback so the field is applied regardless of which path populated it.
	if req.EquityStrategy == "minimax" || requestDTO.EquityStrategy == "minimax" {
		config.EquityStrategy = "minimax"
	}
	log.Printf("[equity] plugin equity_strategy=%s (req=%s dto=%s)", config.EquityStrategy, req.EquityStrategy, requestDTO.EquityStrategy)

	// Lower the TargetFitness threshold when equity mode is active: minimax
	// fitness is bounded by the worst team score, which is typically well below
	// 1.0 when per-criterion targets create inherent trade-offs across groups.
	targetFitness := 0.98
	acceptableFitness := 0.90
	if config.EquityStrategy == "minimax" {
		targetFitness = 0.75
		acceptableFitness = 0.60
	}

	baseConfig := genetic.EngineConfig{
		PopulationSize:  150,
		Generations:     100,
		ElitismRate:     0.15,
		MutationRate:    0.2,
		CrossoverRate:   0.8,
		TournamentSize:  5,
		TargetGroupSize: requestDTO.TeamSize,
		MaxStagnation:   40,
		TargetFitness:   targetFitness,
	}

	const maxRuns = 3
	var bestChromosome *domain.Chromosome
	var bestStats *genetic.EvolutionStats
	bestFitness := -1.0

	for run := 1; run <= maxRuns; run++ {
		seed := time.Now().UnixNano() + int64(run*1000)
		chromosome, stats, runErr := runGeneticWithSeed(ctx, seed, baseConfig, factory, *config, evalContext, students)
		if runErr != nil {
			return nil, fmt.Errorf("team_formation criteria mode: genetic algorithm failed on run %d: %w", run, runErr)
		}
		if chromosome != nil && chromosome.Fitness > bestFitness {
			bestFitness = chromosome.Fitness
			bestChromosome = chromosome
			bestStats = stats
		}
		if bestFitness >= acceptableFitness {
			markRunStats(bestChromosome, run, true, bestFitness)
			break
		}
		if run < maxRuns {
			log.Printf("team_formation: run %d fitness %.4f below %.4f, retrying...", run, bestFitness, acceptableFitness)
		}
	}
	markRunStats(bestChromosome, maxRuns, false, bestFitness)
	return convertCriteriaResult(bestChromosome, bestStats, requestDTO, factory, config, evalContext), nil
}

func (p *TeamFormationPlugin) handleMultiObjectiveCriteriaMode(
	ctx context.Context,
	req TeamFormationPluginRequest,
	students []*domain.Student,
	config *domain.Config,
	factory *genetic.CriterionFactory,
	evalContext domain.Context,
	moConfig genetic.MultiObjectiveConfig,
) (*core.TeamFormationResult, error) {
	requestDTO := req.CriteriaRequest

	moEngineConfig := genetic.DefaultMOEngineConfig()
	moEngineConfig.TargetGroupSize = requestDTO.TeamSize
	moEngineConfig.Seed = time.Now().UnixNano()

	moEngine := genetic.NewMOEngine(moEngineConfig, factory, moConfig, *config, evalContext)
	result, err := moEngine.Run(ctx, students)
	if err != nil {
		return nil, fmt.Errorf("team_formation MO mode: NSGA-II failed: %w", err)
	}

	totalWeight := 0.0
	for _, cp := range requestDTO.CriteriaParameters {
		if cp.Weight > 0 {
			totalWeight += cp.Weight
		}
	}
	weights := make(map[string]float64, len(requestDTO.CriteriaParameters))
	targets := make(map[string]float64, len(requestDTO.CriteriaParameters))
	resultCriteriaParams := make(map[string]core.CriterionConfig, len(requestDTO.CriteriaParameters))
	for name, cp := range requestDTO.CriteriaParameters {
		if totalWeight > 0 {
			weights[name] = cp.Weight / totalWeight
		}
		targets[name] = cp.Target
		resultCriteriaParams[name] = core.CriterionConfig{
			Target: cp.Target,
			Weight: cp.Weight,
			Goal:   cp.Goal,
			Type:   cp.Type,
			Label:  cp.Label,
		}
	}

	paretoSolutions := make([]map[string]interface{}, 0, len(result.ParetoFront))

	for idx, ind := range result.ParetoFront {
		teams := make([]core.Team, 0, len(ind.Chromosome.Groups))
		totalScore := 0.0

		for i, group := range ind.Chromosome.Groups {
			members := make([]core.TeamMember, 0, len(group))
			for _, student := range group {
				members = append(members, core.TeamMember{ID: student.ID})
			}

			groupCopy := make(domain.Group, len(group))
			copy(groupCopy, group)
			teamScore := genetic.CalculateFitnessForGroup(&groupCopy, *config, factory, evalContext)
			totalScore += teamScore

			providerScores := make(map[string]float64, len(requestDTO.CriteriaParameters))
			breakdown := make(map[string]core.CriterionBreakdownEntry, len(requestDTO.CriteriaParameters))

			for criterionName, cp := range requestDTO.CriteriaParameters {
				criterion := factory.Get(criterionName)
				if criterion == nil {
					continue
				}
				score := criterion.Calculate(&groupCopy, evalContext)
				providerScores[criterionName] = score

				realValue := score
				if catCrit, ok := criterion.(*genetic.CategoricalCriterion); ok {
					realValue = catCrit.CalculateRealValue(&groupCopy, evalContext)
				}

				breakdown[criterionName] = core.CriterionBreakdownEntry{
					Score:     score,
					RealValue: realValue,
					Target:    cp.Target,
					Weight:    weights[criterionName],
				}
			}

			teams = append(teams, core.Team{
				ID:            fmt.Sprintf("team_%d", i+1),
				Members:       members,
				AffinityScore: teamScore,
				Metadata: core.TeamMetadata{
					ProviderScores:     providerScores,
					TotalComparisons:   len(members) * (len(members) - 1) / 2,
					IsFeasible:         true,
					CriterionBreakdown: breakdown,
				},
			})
		}

		globalScore := 0.0
		if len(teams) > 0 {
			globalScore = totalScore / float64(len(teams))
		}

		paretoSolutions = append(paretoSolutions, map[string]interface{}{
			"id":             idx,
			"teams":          teams,
			"global_score":   globalScore,
			"psy_score":      ind.Fitness.PsyScore,
			"soc_score":      ind.Fitness.SocScore,
			"rank":           ind.Fitness.Rank,
			"crowding":       ind.Fitness.Distance,
		})
	}

	kneeIdx := 0
	kneeDist := 1e9
	for idx, ind := range result.ParetoFront {
		psy := ind.Fitness.PsyScore
		soc := ind.Fitness.SocScore
		dist := math.Sqrt((1-psy)*(1-psy) + (1-soc)*(1-soc))
		if dist < kneeDist {
			kneeDist = dist
			kneeIdx = idx
		}
	}

	bestInd := result.ParetoFront[kneeIdx]
	bestTeams := paretoSolutions[kneeIdx]["teams"].([]core.Team)
	bestGlobalScore := paretoSolutions[kneeIdx]["global_score"].(float64)

	providersUsed := make([]string, 0, len(requestDTO.CriteriaParameters))
	for name := range requestDTO.CriteriaParameters {
		providersUsed = append(providersUsed, name)
	}
	sort.Strings(providersUsed)

	metadata := map[string]interface{}{
		"algorithm":       "nsga2_criteria",
		"plugin":          "team_formation",
		"mode":            "multi_objective",
		"pareto_front":    paretoSolutions,
		"knee_index":      kneeIdx,
		"pareto_size":     len(result.ParetoFront),
		"total_fronts":    len(result.AllFronts),
		"execution_time_ms": result.ExecutionTimeMs,
		"psy_weights":     moConfig.PsyWeights,
		"soc_weights":     moConfig.SocWeights,
	}
	if result.Stats != nil {
		metadata["generations_run"] = len(result.Stats.Generations)
		metadata["stop_reason"] = result.Stats.StopReason
		metadata["converged"] = result.Stats.Converged
	}

	_ = bestInd

	return &core.TeamFormationResult{
		Teams:           bestTeams,
		GlobalScore:     bestGlobalScore,
		Algorithm:       "nsga2_criteria",
		ProvidersUsed:   providersUsed,
		Weights:         weights,
		Targets:         targets,
		CriteriaParameters: resultCriteriaParams,
		EquityStrategy:  config.EquityStrategy,
		ExecutionTimeMs: result.ExecutionTimeMs,
		Metadata:        metadata,
	}, nil
}

// buildCriteriaMetadata assembles the metadata map for a criteria-mode result.
func buildCriteriaMetadata(stats *genetic.EvolutionStats, bestChromosome *domain.Chromosome) map[string]interface{} {
	metadata := map[string]interface{}{
		"algorithm": "genetic_dynamic",
		"plugin":    "team_formation",
		"mode":      "criteria",
	}
	if bestChromosome != nil {
		metadata["best_fitness"] = bestChromosome.Fitness
	}
	if stats != nil {
		metadata["generations_run"] = len(stats.Generations)
		metadata["total_time_ms"] = stats.TotalTime.Milliseconds()
		metadata["converged"] = stats.Converged
		metadata["stop_reason"] = stats.StopReason
		if len(stats.Generations) > 0 {
			metadata["final_diversity"] = stats.Generations[len(stats.Generations)-1].Diversity
		}
	}
	if bestChromosome != nil && bestChromosome.Stats != nil {
		metadata["chromosome_stats"] = bestChromosome.Stats
		for _, key := range []string{"runs_executed", "early_exit", "final_fitness"} {
			if v, ok := bestChromosome.Stats[key]; ok {
				metadata[key] = v
			}
		}
	}
	return metadata
}

// convertCriteriaResult convierte el resultado del algoritmo genético al formato de la API.
// Réplica de TeamFormationHandler.convertGeneticResultToTeamFormation.
func convertCriteriaResult(
	bestChromosome *domain.Chromosome,
	stats *genetic.EvolutionStats,
	requestDTO *dto.TeamFormationRequestDTO,
	factory *genetic.CriterionFactory,
	config *domain.Config,
	evalContext domain.Context,
) *core.TeamFormationResult {
	if bestChromosome == nil {
		return &core.TeamFormationResult{
			Teams:           []core.Team{},
			GlobalScore:     0,
			Algorithm:       "genetic_dynamic",
			ProvidersUsed:   []string{},
			Weights:         map[string]float64{},
			Targets:         map[string]float64{},
			ExecutionTimeMs: 0,
			Metadata:        map[string]interface{}{"error": "no valid solution found"},
		}
	}

	// Normalize weights so they sum to 1.0 — matches what the genetic engine actually used.
	totalWeight := 0.0
	for _, cp := range requestDTO.CriteriaParameters {
		if cp.Weight > 0 {
			totalWeight += cp.Weight
		}
	}
	weights := make(map[string]float64, len(requestDTO.CriteriaParameters))
	targets := make(map[string]float64, len(requestDTO.CriteriaParameters))
	resultCriteriaParams := make(map[string]core.CriterionConfig, len(requestDTO.CriteriaParameters))
	for name, cp := range requestDTO.CriteriaParameters {
		if totalWeight > 0 {
			weights[name] = cp.Weight / totalWeight
		} else {
			weights[name] = 0
		}
		targets[name] = cp.Target
		resultCriteriaParams[name] = core.CriterionConfig{
			Target: cp.Target,
			Weight: cp.Weight,
			Goal:   cp.Goal,
			Type:   cp.Type,
			Label:  cp.Label,
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
		providerScores := make(map[string]float64, len(requestDTO.CriteriaParameters))
		breakdown := make(map[string]core.CriterionBreakdownEntry, len(requestDTO.CriteriaParameters))

		for criterionName, cp := range requestDTO.CriteriaParameters {
			criterion := factory.Get(criterionName)
			if criterion == nil {
				continue
			}
			// Score uses the absolute-error formula: 1 - |Target - NormalizedBlau|
			score := criterion.Calculate(&groupCopy, evalContext)
			providerScores[criterionName] = score

			// RealValue is the raw normalised diversity metric (e.g. Blau index),
			// independent of the target. For CategoricalCriterion we call
			// CalculateRealValue directly; for other criterion types we fall back
			// to the score value since they don't expose a separate raw metric.
			realValue := score
			if catCrit, ok := criterion.(*genetic.CategoricalCriterion); ok {
				realValue = catCrit.CalculateRealValue(&groupCopy, evalContext)
			}

			breakdown[criterionName] = core.CriterionBreakdownEntry{
				Score:     score,
				RealValue: realValue,
				Target:    cp.Target,
				Weight:    weights[criterionName],
			}
		}

		teams = append(teams, core.Team{
			ID:            fmt.Sprintf("team_%d", i+1),
			Members:       members,
			AffinityScore: teamScore,
			Metadata: core.TeamMetadata{
				ProviderScores:     providerScores,
				TotalComparisons:   len(members) * (len(members) - 1) / 2,
				IsFeasible:         true,
				CriterionBreakdown: breakdown,
			},
		})
	}

	// globalScore: when EquityStrategy is active (minimax or leximin), keep the
	// chromosome's own fitness value (worst-team score). Recomputing a plain average
	// here would silently discard the equity adjustment.
	globalScore := bestChromosome.Fitness
	if len(teams) > 0 && config.EquityStrategy == "" {
		globalScore = totalScore / float64(len(teams))
	}

	var executionTimeMs int64
	if stats != nil {
		executionTimeMs = stats.TotalTime.Milliseconds()
	}

	providersUsed := make([]string, 0, len(requestDTO.CriteriaParameters))
	for name := range requestDTO.CriteriaParameters {
		providersUsed = append(providersUsed, name)
	}
	sort.Strings(providersUsed)
	if len(providersUsed) == 0 {
		providersUsed = []string{"dynamic_criteria"}
	}

	metadata := buildCriteriaMetadata(stats, bestChromosome)
	// Enrich metadata with equity audit fields so callers can audit the result.
	eq := genetic.ComputeEquityStats(bestChromosome, *config, factory, evalContext)
	metadata["equity_enforced"] = eq.EquityEnforced
	metadata["lowest_team_score"] = eq.LowestTeamScore
	metadata["highest_team_score"] = eq.HighestTeamScore
	metadata["score_std_dev"] = eq.ScoreStdDev
	metadata["score_variance"] = eq.ScoreVariance

	return &core.TeamFormationResult{
		Teams:           teams,
		GlobalScore:     globalScore,
		Algorithm:       "genetic_dynamic",
		ProvidersUsed:   providersUsed,
		Weights:         weights,
		Targets:            targets,
		CriteriaParameters: resultCriteriaParams,
		EquityStrategy:     config.EquityStrategy,
		ExecutionTimeMs: executionTimeMs,
		Metadata:        metadata,
	}
}

// ---------------------------------------------------------------------------
// Wire types
// ---------------------------------------------------------------------------

// TeamFormationPluginRequest es la estructura que el kernel envía al plugin.
// El campo Mode determina cuál de los dos modos de optimización se ejecuta.
type TeamFormationPluginRequest struct {
	// Mode selecciona el algoritmo de entrada:
	//   "affinity" — usa AffinityMatrix pre-calculada por el kernel.
	//   "criteria" — usa CriteriaRequest con datos demográficos y criterios dinámicos.
	Mode string `json:"mode"`

	// ── Campos comunes ──────────────────────────────────────────────────────
	TeamSize    int    `json:"team_size"`
	MinTeamSize int    `json:"min_team_size"`
	MaxTeamSize int    `json:"max_team_size"`
	Algorithm   string `json:"algorithm,omitempty"` // "greedy" | "genetic"

	// ── Modo "affinity" ─────────────────────────────────────────────────────
	SectionID      string                        `json:"section_id,omitempty"`
	Candidates     []string                      `json:"candidates,omitempty"`
	AffinityMatrix map[string]map[string]float64 `json:"affinity_matrix,omitempty"`
	// Providers es un array de matrices T_p con sus pesos e intencionalidad.
	// Cada provider representa una tabla de compatibilidad separada (ej. eneagrama, test social).
	// El plugin computa C'_ij = Σ_p ω_p * (1 - S_p * T_p(i,j)) / 2 según la tesis.
	Providers       []ProviderData     `json:"providers,omitempty"`
	ProviderIDs     []string           `json:"provider_ids,omitempty"`
	ProviderWeights map[string]float64 `json:"provider_weights,omitempty"`
	// Parámetros del algoritmo genético (opcionales; se usan defaults si son 0)
	GeneticPopSize  int     `json:"genetic_pop_size,omitempty"`
	GeneticGens     int     `json:"genetic_gens,omitempty"`
	GeneticMutation float64 `json:"genetic_mutation,omitempty"`

	// ── Modo "criteria" ─────────────────────────────────────────────────────
	CriteriaRequest *dto.TeamFormationRequestDTO `json:"criteria_request,omitempty"`
	// EquityStrategy activa la agregación por equidad: "minimax" o "leximin".
	// Propagated from the HTTP request into the domain config before the genetic run.
	EquityStrategy string `json:"equity_strategy,omitempty"`
}

// ProviderData representa una tabla de compatibilidad T_p con su peso e intencionalidad.
// Según la tesis: C'_ij = Σ_p ω_p * (1 - S_p * T_p(i,j)) / 2
type ProviderData struct {
	ProviderID     string                        `json:"provider_id"`
	TpMatrix       map[string]map[string]float64 `json:"tp_matrix"`      // T_p(i,j) ∈ [-1, 1]
	Weight         float64                       `json:"weight"`         // ω_p (sum should be 1.0)
	Intentionality int                           `json:"intentionality"` // S_p: 1=harmony, -1=controlled friction
}
