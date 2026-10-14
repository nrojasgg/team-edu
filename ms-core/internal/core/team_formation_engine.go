package core

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"ms-core/internal/core/domain"
	"ms-core/internal/core/genetic"
	"sort"
	"sync"
	"time"
)

// TeamFormationEngine es el motor de optimización agnóstico para formar equipos.
// No conoce el origen de los datos de afinidad, solo trabaja con AffinityProviders.

// teamIDFmt is the format string used to create team identifiers.
const teamIDFmt = "team-%d"

type TeamFormationEngine struct {
	providers []WeightedProvider
	config    *TeamFormationConfig
	mu        sync.RWMutex
}

// TeamFormationConfig contiene la configuración del motor
type TeamFormationConfig struct {
	// Weights mapea provider_id -> peso (0.0 a 1.0)
	Weights map[string]float64 `json:"weights"`

	// Algorithm define el algoritmo a usar: "greedy", "genetic", "simulated_annealing"
	// Por defecto: "genetic"
	Algorithm string `json:"algorithm"`

	// GeneticParams parámetros para algoritmo genético
	GeneticParams *GeneticAlgorithmParams `json:"genetic_params,omitempty"`

	// DefaultWeight peso por defecto si no se especifica para un provider
	DefaultWeight float64 `json:"default_weight"`
}

// GeneticAlgorithmParams parámetros del algoritmo genético
type GeneticAlgorithmParams struct {
	PopulationSize int     `json:"population_size"`
	Generations    int     `json:"generations"`
	MutationRate   float64 `json:"mutation_rate"`
	ElitismRate    float64 `json:"elitism_rate"`
}

// DefaultTeamFormationConfig retorna configuración por defecto
func DefaultTeamFormationConfig() *TeamFormationConfig {
	return &TeamFormationConfig{
		Weights:       make(map[string]float64),
		Algorithm:     "genetic",
		DefaultWeight: 1.0,
		GeneticParams: &GeneticAlgorithmParams{
			PopulationSize: 100,
			Generations:    50,
			MutationRate:   0.1,
			ElitismRate:    0.1,
		},
	}
}

// NewTeamFormationEngine crea una nueva instancia del motor
func NewTeamFormationEngine(config *TeamFormationConfig) *TeamFormationEngine {
	if config == nil {
		config = DefaultTeamFormationConfig()
	}
	return &TeamFormationEngine{
		providers: make([]WeightedProvider, 0),
		config:    config,
	}
}

// RegisterProvider registra un AffinityProvider con su peso configurado
func (e *TeamFormationEngine) RegisterProvider(provider AffinityProvider) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	providerID := provider.ProviderID()

	// Verificar que no esté duplicado
	for _, wp := range e.providers {
		if wp.Provider.ProviderID() == providerID {
			return fmt.Errorf("provider already registered: %s", providerID)
		}
	}

	// Obtener peso de la configuración o usar el default
	weight := e.config.DefaultWeight
	if w, ok := e.config.Weights[providerID]; ok {
		weight = w
	}

	e.providers = append(e.providers, WeightedProvider{
		Provider: provider,
		Weight:   weight,
	})

	log.Printf("AffinityProvider registered: %s (weight: %.2f)", providerID, weight)
	return nil
}

// UpdateWeights actualiza los pesos de los providers dinámicamente
func (e *TeamFormationEngine) UpdateWeights(weights map[string]float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.config.Weights = weights

	// Actualizar los providers ya registrados
	for i, wp := range e.providers {
		if weight, ok := weights[wp.Provider.ProviderID()]; ok {
			e.providers[i].Weight = weight
		}
	}
}

// GetProviderInfo retorna información de los providers registrados
func (e *TeamFormationEngine) GetProviderInfo() []ProviderMetadata {
	e.mu.RLock()
	defer e.mu.RUnlock()

	info := make([]ProviderMetadata, len(e.providers))
	for i, wp := range e.providers {
		info[i] = wp.Provider.GetMetadata()
	}
	return info
}

// CalculateCombinedAffinity calcula la afinidad combinada entre dos personas
// ponderando todos los providers registrados
func (e *TeamFormationEngine) CalculateCombinedAffinity(ctx context.Context, personA, personB string) (float64, map[string]float64, error) {
	e.mu.RLock()
	providers := make([]WeightedProvider, len(e.providers))
	copy(providers, e.providers)
	e.mu.RUnlock()

	if len(providers) == 0 {
		log.Printf("WARNING: No affinity providers registered!")
		return 0.5, nil, fmt.Errorf("no affinity providers registered")
	}

	var combinedScore float64
	var totalWeight float64
	providerScores := make(map[string]float64)
	var activeProviders int

	for _, wp := range providers {
		// Ignorar providers con peso 0
		if wp.Weight == 0 {
			continue
		}

		// Verificar si el provider soporta el contexto actual
		if !wp.Provider.SupportsContext(ctx) {
			log.Printf("Provider %s does not support current context", wp.Provider.ProviderID())
			continue
		}

		activeProviders++
		score, err := wp.Provider.CalculatePairwiseAffinity(ctx, personA, personB)
		if err != nil {
			log.Printf("Warning: Provider %s failed for pair (%s, %s): %v",
				wp.Provider.ProviderID(), personA, personB, err)
			continue
		}

		providerScores[wp.Provider.ProviderID()] = score
		combinedScore += score * wp.Weight
		totalWeight += wp.Weight
	}

	if totalWeight == 0 {
		// Si no hay providers activos, retornar 0.5 (neutral) en lugar de 0
		log.Printf("WARNING: No active providers with weight > 0 for pair (%s, %s). Providers registered: %d, active: %d",
			personA, personB, len(providers), activeProviders)
		return 0.5, providerScores, nil
	}

	// Normalizar por el peso total
	normalizedScore := combinedScore / totalWeight
	return normalizedScore, providerScores, nil
}

// FormTeams ejecuta el algoritmo de formación de equipos
func (e *TeamFormationEngine) FormTeams(ctx context.Context, request TeamFormationRequest) (*TeamFormationResult, error) {
	startTime := time.Now()

	if len(request.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates provided")
	}

	if request.TeamSize <= 0 {
		return nil, fmt.Errorf("team size must be positive")
	}

	// Agregar section_id al contexto
	ctx = context.WithValue(ctx, ContextKeySectionID, request.SectionID)

	// Usar el algoritmo inteligente que prueba diferentes configuraciones
	result, err := e.formTeamsIntelligent(ctx, request)

	if err != nil {
		return nil, err
	}

	result.ExecutionTimeMs = time.Since(startTime).Milliseconds()
	return result, nil
}

// FormTeamsWithConfig ejecuta el algoritmo de formación con configuración personalizada
// Permite especificar providers y pesos diferentes a los configurados globalmente
func (e *TeamFormationEngine) FormTeamsWithConfig(ctx context.Context, request TeamFormationRequest) (*TeamFormationResult, error) {
	startTime := time.Now()

	if len(request.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates provided")
	}

	if request.TeamSize <= 0 {
		return nil, fmt.Errorf("team size must be positive")
	}

	// Agregar section_id al contexto
	ctx = context.WithValue(ctx, ContextKeySectionID, request.SectionID)

	// Si hay pesos personalizados, usarlos temporalmente
	var originalWeights map[string]float64
	if len(request.ProviderWeights) > 0 {
		originalWeights = e.getProviderWeights()
		e.applyCustomWeights(request.ProviderWeights)
		defer e.restoreWeights(originalWeights)

		// Log de providers activos después de aplicar pesos
		log.Printf("FormTeamsWithConfig: Applied custom weights for providers")
		for providerID, weight := range request.ProviderWeights {
			log.Printf("  Provider %s: weight=%.2f", providerID, weight)
		}
	}

	// Siempre usar el algoritmo inteligente para mejor distribución de equipos
	result, err := e.formTeamsIntelligent(ctx, request)

	if err != nil {
		return nil, err
	}

	// Inicializar metadata si es nil
	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}

	result.ExecutionTimeMs = time.Since(startTime).Milliseconds()
	return result, nil
}

// getProviderWeights obtiene los pesos actuales de los providers
func (e *TeamFormationEngine) getProviderWeights() map[string]float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	weights := make(map[string]float64)
	for _, wp := range e.providers {
		weights[wp.Provider.ProviderID()] = wp.Weight
	}
	return weights
}

// applyCustomWeights aplica pesos personalizados a los providers
func (e *TeamFormationEngine) applyCustomWeights(weights map[string]float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for i, wp := range e.providers {
		providerID := wp.Provider.ProviderID()
		if weight, ok := weights[providerID]; ok {
			e.providers[i].Weight = weight
		} else {
			// Si el provider no está en los pesos personalizados, peso = 0
			e.providers[i].Weight = 0
		}
	}
}

// restoreWeights restaura los pesos originales de los providers
func (e *TeamFormationEngine) restoreWeights(weights map[string]float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for i, wp := range e.providers {
		if weight, ok := weights[wp.Provider.ProviderID()]; ok {
			e.providers[i].Weight = weight
		}
	}
}

// buildAffinityMatrix computes pairwise affinities for all candidates.
// Any error from CalculateCombinedAffinity is logged and treated as 0.
func (e *TeamFormationEngine) buildAffinityMatrix(ctx context.Context, candidates []TeamMember) map[string]map[string]float64 {
	matrix := make(map[string]map[string]float64, len(candidates))
	for _, c := range candidates {
		matrix[c.ID] = make(map[string]float64)
	}
	for i, c1 := range candidates {
		for j := i + 1; j < len(candidates); j++ {
			c2 := candidates[j]
			score, _, err := e.CalculateCombinedAffinity(ctx, c1.ID, c2.ID)
			if err != nil {
				log.Printf("Warning: could not calculate affinity for (%s, %s): %v", c1.ID, c2.ID, err)
				score = 0
			}
			matrix[c1.ID][c2.ID] = score
			matrix[c2.ID][c1.ID] = score
		}
	}
	return matrix
}

// buildRelationalCostMatrix computes the relational cost matrix C'_ij according to thesis formula:
// C'_ij = Σ_p ω_p * (1 - S_p * T_p(i,j)) / 2
// Returns a matrix of relational costs along with provider intentionalities for later use.
//
// If matrixContainsCosts is true, the input matrix already contains C'_ij values (e.g., from
// affinity mode) and ComputeRelationalCost is skipped - values are used directly.
func (e *TeamFormationEngine) buildRelationalCostMatrix(ctx context.Context, candidates []TeamMember, matrixContainsCosts bool) (map[string]map[string]float64, []ProviderIntentionality) {
	matrix := make(map[string]map[string]float64, len(candidates))
	for _, c := range candidates {
		matrix[c.ID] = make(map[string]float64)
	}

	// Collect provider intentionalities from registered providers
	providers := e.collectProviderIntentionalities()

	for i, c1 := range candidates {
		for j := i + 1; j < len(candidates); j++ {
			c2 := candidates[j]
			_, providerScores, err := e.CalculateCombinedAffinity(ctx, c1.ID, c2.ID)
			if err != nil {
				log.Printf("Warning: could not calculate affinity for (%s, %s): %v", c1.ID, c2.ID, err)
				providerScores = make(map[string]float64)
			}

			var cost float64
			if matrixContainsCosts {
				// Matrix already contains C'_ij values (e.g., from affinity mode)
				cost = providerScores["precomputed"] // mapAffinityProvider stores C'_ij here
				if cost == 0 {
					cost = 0.5 // Default for uncomputed pairs
				}
			} else {
				// Compute C'_ij using thesis formula from raw T_p values
				cost = ComputeRelationalCost(providerScores, providers)
			}
			matrix[c1.ID][c2.ID] = cost
			matrix[c2.ID][c1.ID] = cost
		}
	}
	return matrix, providers
}

// collectProviderIntentionalities gathers S_p and ω_p from all registered providers.
func (e *TeamFormationEngine) collectProviderIntentionalities() []ProviderIntentionality {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]ProviderIntentionality, 0, len(e.providers))
	for _, wp := range e.providers {
		metadata := wp.Provider.GetMetadata()
		intentionality := 1 // default: harmony
		if metadata.SocialIntentionality != 0 {
			intentionality = metadata.SocialIntentionality
		}
		result = append(result, ProviderIntentionality{
			ProviderID:           wp.Provider.ProviderID(),
			Weight:               wp.Weight,
			SocialIntentionality: intentionality,
		})
	}
	return result
}

// sortCandidatesByConnectivity reorders candidates so the most-connected ones come first.
func sortCandidatesByConnectivity(candidates []TeamMember, affinityMatrix map[string]map[string]float64) []TeamMember {
	type candidateScore struct {
		member   TeamMember
		sumScore float64
	}
	scored := make([]candidateScore, len(candidates))
	for i, c := range candidates {
		var sum float64
		for _, score := range affinityMatrix[c.ID] {
			sum += score
		}
		scored[i] = candidateScore{member: c, sumScore: sum}
	}
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].sumScore > scored[j].sumScore
	})
	result := make([]TeamMember, len(candidates))
	for i, cs := range scored {
		result[i] = cs.member
	}
	return result
}

// buildAssignedSet returns a set of all member IDs present in the given teams.
func buildAssignedSet(teams []Team) map[string]bool {
	assigned := make(map[string]bool)
	for _, t := range teams {
		for _, m := range t.Members {
			assigned[m.ID] = true
		}
	}
	return assigned
}

// findBestCandidateForTeam returns the index in remaining of the candidate
// with the highest average affinity to the current team members, or -1 if empty.
func findBestCandidateForTeam(team []TeamMember, remaining []TeamMember, affinityMatrix map[string]map[string]float64) int {
	bestIdx := -1
	bestScore := -2.0 // Mínimo posible es -1.0
	for i, candidate := range remaining {
		var avgAffinity float64
		for _, member := range team {
			avgAffinity += affinityMatrix[candidate.ID][member.ID]
		}
		avgAffinity /= float64(len(team))
		if avgAffinity > bestScore {
			bestScore = avgAffinity
			bestIdx = i
		}
	}
	return bestIdx
}

// formTeamsGreedy implementa un algoritmo voraz para formar equipos
func (e *TeamFormationEngine) formTeamsGreedy(ctx context.Context, request TeamFormationRequest) (*TeamFormationResult, error) {
	candidates := make([]TeamMember, len(request.Candidates))
	copy(candidates, request.Candidates)

	// Calcular matriz de afinidad y ordenar por conectividad
	affinityMatrix := e.buildAffinityMatrix(ctx, candidates)
	candidates = sortCandidatesByConnectivity(candidates, affinityMatrix)

	// Formar equipos greedily
	teams := make([]Team, 0)
	assigned := make(map[string]bool)
	teamCounter := 1

	for len(getUnassigned(candidates, assigned)) >= request.TeamSize {
		team := e.formBestTeam(candidates, assigned, affinityMatrix, request.TeamSize)
		if len(team.Members) == 0 {
			break
		}
		team.ID = fmt.Sprintf(teamIDFmt, teamCounter)
		teams = append(teams, team)
		teamCounter++

		for _, m := range team.Members {
			assigned[m.ID] = true
		}
	}

	// Manejar candidatos restantes
	unassigned := getUnassigned(candidates, assigned)

	// Si hay suficientes para un equipo mínimo, formar uno
	if len(unassigned) >= request.MinTeamSize && request.MinTeamSize > 0 {
		team := Team{
			ID:      fmt.Sprintf(teamIDFmt, teamCounter),
			Members: unassigned,
		}
		team.AffinityScore = e.calculateTeamScore(ctx, team.Members, affinityMatrix)
		teams = append(teams, team)
		unassigned = nil
	}

	// Calcular score global
	var globalScore float64
	for _, t := range teams {
		globalScore += t.AffinityScore
	}
	if len(teams) > 0 {
		globalScore /= float64(len(teams))
	}

	providersUsed, weights := e.collectProvidersInfo()

	return &TeamFormationResult{
		Teams:             teams,
		UnassignedMembers: unassigned,
		GlobalScore:       globalScore,
		Algorithm:         "greedy",
		ProvidersUsed:     providersUsed,
		Weights:           weights,
	}, nil
}

// formBestTeam forma el mejor equipo posible con los candidatos disponibles
func (e *TeamFormationEngine) formBestTeam(
	candidates []TeamMember,
	assigned map[string]bool,
	affinityMatrix map[string]map[string]float64,
	teamSize int,
) Team {
	available := getUnassigned(candidates, assigned)
	if len(available) < teamSize {
		return Team{}
	}

	// Empezar con el primer candidato disponible (el de mayor conectividad)
	team := []TeamMember{available[0]}
	remaining := available[1:]

	for len(team) < teamSize && len(remaining) > 0 {
		bestCandidate := findBestCandidateForTeam(team, remaining, affinityMatrix)
		if bestCandidate >= 0 {
			team = append(team, remaining[bestCandidate])
			remaining = append(remaining[:bestCandidate], remaining[bestCandidate+1:]...)
		} else {
			break
		}
	}

	// Calcular score del equipo
	var teamScore float64
	comparisons := 0
	for i := 0; i < len(team); i++ {
		for j := i + 1; j < len(team); j++ {
			teamScore += affinityMatrix[team[i].ID][team[j].ID]
			comparisons++
		}
	}
	if comparisons > 0 {
		teamScore /= float64(comparisons)
	}

	return Team{
		Members:       team,
		AffinityScore: teamScore,
		Metadata: TeamMetadata{
			TotalComparisons: comparisons,
		},
	}
}

// Nsga2Evolve runs the NSGA-II evolution loop starting from initPopulation
// and returns the final sorted Pareto fronts. The evalFn callback evaluates
// a chromosome and returns its FitnessResult.
func (e *TeamFormationEngine) Nsga2Evolve(
	population []*genetic.Individual,
	params *GeneticAlgorithmParams,
	rng *rand.Rand,
	evalFn func([]int) genetic.FitnessResult,
	onConvergence func(int, float64),
) [][]*genetic.Individual {
	for gen := 0; gen < params.Generations; gen++ {
		fronts := genetic.FastNonDominatedSort(population)
		for _, f := range fronts {
			genetic.CalculateCrowdingDistance(f)
		}

		if onConvergence != nil && len(fronts) > 0 && len(fronts[0]) > 0 {
			hv := genetic.Hypervolume2D(fronts[0], 0.0, 0.0)
			onConvergence(gen+1, hv)
		}

		offspring := make([]*genetic.Individual, 0, len(population))
		for len(offspring) < len(population) {
			p1 := e.tournamentSelectNSGA2(population, rng)
			p2 := e.tournamentSelectNSGA2(population, rng)
			childChrom := e.orderCrossover(p1.Chromosome, p2.Chromosome, rng)
			if rng.Float64() < params.MutationRate {
				e.mutate(childChrom, rng)
			}
			offspring = append(offspring, &genetic.Individual{
				Chromosome: childChrom,
				Fitness:    evalFn(childChrom),
			})
		}

		combined := append(population, offspring...)
		fronts = genetic.FastNonDominatedSort(combined)

		newPopulation := make([]*genetic.Individual, 0, len(population))
		target := len(population)
		for _, front := range fronts {
			genetic.CalculateCrowdingDistance(front)
			if len(newPopulation)+len(front) <= target {
				newPopulation = append(newPopulation, front...)
			} else {
				sort.Slice(front, func(i, j int) bool {
					return front[i].Fitness.Distance > front[j].Fitness.Distance
				})
				newPopulation = append(newPopulation, front[:target-len(newPopulation)]...)
				break
			}
		}
		population = newPopulation
	}
	return genetic.FastNonDominatedSort(population)
}

func (e *TeamFormationEngine) Nsga2EvolveNoCrowding(
	population []*genetic.Individual,
	params *GeneticAlgorithmParams,
	rng *rand.Rand,
	evalFn func([]int) genetic.FitnessResult,
	onConvergence func(int, float64),
) [][]*genetic.Individual {
	for gen := 0; gen < params.Generations; gen++ {
		fronts := genetic.FastNonDominatedSort(population)
		for _, f := range fronts {
			genetic.CalculateCrowdingDistance(f)
		}

		if onConvergence != nil && len(fronts) > 0 && len(fronts[0]) > 0 {
			hv := genetic.Hypervolume2D(fronts[0], 0.0, 0.0)
			onConvergence(gen+1, hv)
		}

		offspring := make([]*genetic.Individual, 0, len(population))
		for len(offspring) < len(population) {
			p1 := e.tournamentSelectNoCrowding(population, rng)
			p2 := e.tournamentSelectNoCrowding(population, rng)
			childChrom := e.orderCrossover(p1.Chromosome, p2.Chromosome, rng)
			if rng.Float64() < params.MutationRate {
				e.mutate(childChrom, rng)
			}
			offspring = append(offspring, &genetic.Individual{
				Chromosome: childChrom,
				Fitness:    evalFn(childChrom),
			})
		}

		combined := append(population, offspring...)
		fronts = genetic.FastNonDominatedSort(combined)

		newPopulation := make([]*genetic.Individual, 0, len(population))
		target := len(population)
		for _, front := range fronts {
			genetic.CalculateCrowdingDistance(front)
			if len(newPopulation)+len(front) <= target {
				newPopulation = append(newPopulation, front...)
			} else {
				sort.Slice(front, func(i, j int) bool {
					return front[i].Fitness.Distance > front[j].Fitness.Distance
				})
				newPopulation = append(newPopulation, front[:target-len(newPopulation)]...)
				break
			}
		}
		population = newPopulation
	}
	return genetic.FastNonDominatedSort(population)
}

func (e *TeamFormationEngine) WeightedSumEvolve(
	population []*genetic.Individual,
	params *GeneticAlgorithmParams,
	rng *rand.Rand,
	evalFn func([]int) genetic.FitnessResult,
	onConvergence func(int, float64),
) []*genetic.Individual {
	for gen := 0; gen < params.Generations; gen++ {
		fronts := genetic.FastNonDominatedSort(population)
		for _, f := range fronts {
			genetic.CalculateCrowdingDistance(f)
		}

		if onConvergence != nil && len(fronts) > 0 && len(fronts[0]) > 0 {
			hv := genetic.Hypervolume2D(fronts[0], 0.0, 0.0)
			onConvergence(gen+1, hv)
		}

		offspring := make([]*genetic.Individual, 0, len(population))
		for len(offspring) < len(population) {
			p1 := e.tournamentSelectWeightedSum(population, rng)
			p2 := e.tournamentSelectWeightedSum(population, rng)
			childChrom := e.orderCrossover(p1.Chromosome, p2.Chromosome, rng)
			if rng.Float64() < params.MutationRate {
				e.mutate(childChrom, rng)
			}
			offspring = append(offspring, &genetic.Individual{
				Chromosome: childChrom,
				Fitness:    evalFn(childChrom),
			})
		}

		combined := append(population, offspring...)

		sort.Slice(combined, func(i, j int) bool {
			wi := combined[i].Fitness.PsyScore*0.5 + combined[i].Fitness.SocScore*0.5
			wj := combined[j].Fitness.PsyScore*0.5 + combined[j].Fitness.SocScore*0.5
			return wi > wj // Descending: best first (maximization)
		})

		eliteCount := int(float64(len(population)) * params.ElitismRate)
		if eliteCount < 1 {
			eliteCount = 1
		}

		newPopulation := make([]*genetic.Individual, 0, len(population))
		newPopulation = append(newPopulation, combined[:eliteCount]...)

		for len(newPopulation) < len(population) {
			newPopulation = append(newPopulation, combined[len(newPopulation)])
		}

		population = newPopulation[:len(population)]
	}

	sort.Slice(population, func(i, j int) bool {
		wi := population[i].Fitness.PsyScore*0.5 + population[i].Fitness.SocScore*0.5
		wj := population[j].Fitness.PsyScore*0.5 + population[j].Fitness.SocScore*0.5
		return wi > wj // Descending: best first (maximization)
	})

	return population
}

// resolveGeneticParams returns params if non-nil, otherwise default genetic params.
func resolveGeneticParams(params *GeneticAlgorithmParams) *GeneticAlgorithmParams {
	if params != nil {
		return params
	}
	return &GeneticAlgorithmParams{
		PopulationSize: 100,
		Generations:    50,
		MutationRate:   0.1,
		ElitismRate:    0.1,
	}
}

// collectProvidersInfo collects provider IDs and weights under the read lock.
func (e *TeamFormationEngine) collectProvidersInfo() ([]string, map[string]float64) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	ids := make([]string, 0, len(e.providers))
	weights := make(map[string]float64, len(e.providers))
	for _, wp := range e.providers {
		ids = append(ids, wp.Provider.ProviderID())
		weights[wp.Provider.ProviderID()] = wp.Weight
	}
	return ids, weights
}

// initializePopulation creates the initial NSGA-II population.
func (e *TeamFormationEngine) initializePopulation(
	size int,
	candidates []TeamMember,
	affinityMatrix map[string]map[string]float64,
	teamSize int,
	rng *rand.Rand,
) []*genetic.Individual {
	population := make([]*genetic.Individual, size)
	for i := range population {
		chromosome := rng.Perm(len(candidates))
		population[i] = &genetic.Individual{
			Chromosome: chromosome,
			Fitness:    e.EvaluateChromosome(chromosome, candidates, affinityMatrix, teamSize),
		}
	}
	return population
}

// scoreTeams computes and assigns AffinityScore, ID and Metadata for each team.
func (e *TeamFormationEngine) scoreTeams(ctx context.Context, teams []Team, affinityMatrix map[string]map[string]float64) {
	for i := range teams {
		oceanVectors := getOCEANVectors(teams[i].Members)
		psyScore := domain.CalculateGroupFitness(oceanVectors)
		socScore := e.calculateTeamScore(ctx, teams[i].Members, affinityMatrix)
		teams[i].AffinityScore = (psyScore + socScore) / 2.0
		teams[i].ID = fmt.Sprintf(teamIDFmt, i+1)
		teams[i].Metadata = TeamMetadata{
			PsychometricScore: psyScore,
			SociometricScore:  socScore,
		}
	}
}

// formTeamsGenetic implementa un algoritmo genético NSGA-II para formar equipos
func (e *TeamFormationEngine) formTeamsGenetic(ctx context.Context, request TeamFormationRequest) (*TeamFormationResult, error) {
	params := resolveGeneticParams(e.config.GeneticParams)

	candidates := request.Candidates
	numTeams := len(candidates) / request.TeamSize

	if numTeams == 0 {
		return nil, fmt.Errorf("not enough candidates to form teams")
	}

	log.Printf("formTeamsGenetic (NSGA-II): Processing %d candidates for section %s", len(candidates), request.SectionID)

	// Precompute affinity matrix using shared helper
	affinityMatrix := e.buildAffinityMatrix(ctx, candidates)

	// Initialize Population
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	population := e.initializePopulation(params.PopulationSize, candidates, affinityMatrix, request.TeamSize, rng)

	// Evolve
	evalFn := func(c []int) genetic.FitnessResult {
		return e.EvaluateChromosome(c, candidates, affinityMatrix, request.TeamSize)
	}
	finalFronts := e.Nsga2Evolve(population, params, rng, evalFn, nil)

	if len(finalFronts) == 0 {
		return nil, fmt.Errorf("genetic algorithm failed")
	}
	paretoFront := finalFronts[0]
	bestInd := e.selectKneePoint(paretoFront)
	if bestInd == nil {
		bestInd = paretoFront[0]
	}

	// Decode and score teams
	teams := e.decodeChromosome(bestInd.Chromosome, candidates, request.TeamSize)
	e.scoreTeams(ctx, teams, affinityMatrix)

	// Collect unassigned candidates
	var unassigned []TeamMember
	if len(candidates)%request.TeamSize != 0 {
		unassigned = getUnassigned(candidates, buildAssignedSet(teams))
	}

	providersUsed, weights := e.collectProvidersInfo()

	return &TeamFormationResult{
		Teams:             teams,
		UnassignedMembers: unassigned,
		GlobalScore:       (bestInd.Fitness.PsyScore + bestInd.Fitness.SocScore) / 2.0,
		Algorithm:         "nsga-2",
		ProvidersUsed:     providersUsed,
		Weights:           weights,
		Metadata: map[string]interface{}{
			"psy_score":   bestInd.Fitness.PsyScore,
			"soc_score":   bestInd.Fitness.SocScore,
			"pareto_rank": bestInd.Fitness.Rank,
		},
	}, nil
}

// EvaluateChromosome evalúa un cromosoma y retorna el resultado de fitness
func (e *TeamFormationEngine) EvaluateChromosome(
	chromosome []int,
	candidates []TeamMember,
	affinityMatrix map[string]map[string]float64,
	teamSize int,
) genetic.FitnessResult {
	teams := e.decodeChromosome(chromosome, candidates, teamSize)

	if len(teams) == 0 {
		return genetic.FitnessResult{IsFeasible: false}
	}

	var totalPsyScore float64
	var totalSocScore float64

	for _, team := range teams {
		oceanVectors := getOCEANVectors(team.Members)
		// CalculateGroupFitness returns float64 (0-1)
		psy := domain.CalculateGroupFitness(oceanVectors)
		totalPsyScore += psy

		soc := e.calculateTeamScoreFromMatrix(team.Members, affinityMatrix)
		totalSocScore += soc
	}

	avgPsy := totalPsyScore / float64(len(teams))
	avgSoc := totalSocScore / float64(len(teams))

	return genetic.FitnessResult{
		PsyScore:   avgPsy,
		SocScore:   avgSoc,
		IsFeasible: true,
	}
}

// EvaluateChromosomeWithAggregation evaluates a chromosome and returns a scalar value
// using the specified aggregation strategy: "weighted_sum" or "minimax"
func (e *TeamFormationEngine) EvaluateChromosomeWithAggregation(
	chromosome []int,
	candidates []TeamMember,
	affinityMatrix map[string]map[string]float64,
	teamSize int,
	strategy string,
) float64 {
	result := e.EvaluateChromosome(chromosome, candidates, affinityMatrix, teamSize)

	if !result.IsFeasible {
		return -1000.0
	}

	avgPsy := result.PsyScore
	avgSoc := result.SocScore

	switch strategy {
	case "weighted_sum":
		return (avgPsy + avgSoc) / 2.0
	case "minimax":
		if avgPsy < avgSoc {
			return avgPsy
		}
		return avgSoc
	default:
		return (avgPsy + avgSoc) / 2.0
	}
}

// formTeamsWeightedSum implements a simple genetic algorithm with single-objective fitness
// using the specified aggregation strategy ("weighted_sum" or "minimax").
// Returns just the best solution found, not the Pareto front.
func (e *TeamFormationEngine) formTeamsWeightedSum(ctx context.Context, request TeamFormationRequest, strategy string) (*TeamFormationResult, error) {
	params := resolveGeneticParams(e.config.GeneticParams)

	candidates := request.Candidates
	numTeams := len(candidates) / request.TeamSize

	if numTeams == 0 {
		return nil, fmt.Errorf("not enough candidates to form teams")
	}

	log.Printf("formTeamsWeightedSum (%s): Processing %d candidates for section %s", strategy, len(candidates), request.SectionID)

	affinityMatrix := e.buildAffinityMatrix(ctx, candidates)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	population := e.initializePopulationWeightedSum(params.PopulationSize, candidates, affinityMatrix, request.TeamSize, rng, strategy)

	bestFitness := -1.0
	bestChromosome := make([]int, len(population[0].Chromosome))
	copy(bestChromosome, population[0].Chromosome)

	for gen := 0; gen < params.Generations; gen++ {
		offspring := make([]*genetic.Individual, 0, len(population))
		for len(offspring) < len(population) {
			p1 := e.tournamentSelectWeightedSum(population, rng)
			p2 := e.tournamentSelectWeightedSum(population, rng)
			childChrom := e.orderCrossover(p1.Chromosome, p2.Chromosome, rng)
			if rng.Float64() < params.MutationRate {
				e.mutate(childChrom, rng)
			}
			fitness := e.EvaluateChromosomeWithAggregation(childChrom, candidates, affinityMatrix, request.TeamSize, strategy)
			offspring = append(offspring, &genetic.Individual{
				Chromosome: childChrom,
				Fitness:    genetic.FitnessResult{PsyScore: fitness, SocScore: fitness, IsFeasible: fitness > -1000},
			})
		}

		combined := append(population, offspring...)
		sort.Slice(combined, func(i, j int) bool {
			return combined[i].Fitness.PsyScore > combined[j].Fitness.PsyScore
		})
		population = combined[:params.PopulationSize]

		if len(population) > 0 && population[0].Fitness.PsyScore > bestFitness {
			bestFitness = population[0].Fitness.PsyScore
			copy(bestChromosome, population[0].Chromosome)
		}
	}

	teams := e.decodeChromosome(bestChromosome, candidates, request.TeamSize)
	e.scoreTeams(ctx, teams, affinityMatrix)

	var unassigned []TeamMember
	if len(candidates)%request.TeamSize != 0 {
		unassigned = getUnassigned(candidates, buildAssignedSet(teams))
	}

	providersUsed, weights := e.collectProvidersInfo()

	return &TeamFormationResult{
		Teams:             teams,
		UnassignedMembers: unassigned,
		GlobalScore:       bestFitness,
		Algorithm:         "genetic_weighted_sum_" + strategy,
		ProvidersUsed:     providersUsed,
		Weights:           weights,
		Metadata: map[string]interface{}{
			"strategy": strategy,
		},
	}, nil
}

// initializePopulationWeightedSum creates the initial population for weighted sum GA
func (e *TeamFormationEngine) initializePopulationWeightedSum(
	size int,
	candidates []TeamMember,
	affinityMatrix map[string]map[string]float64,
	teamSize int,
	rng *rand.Rand,
	strategy string,
) []*genetic.Individual {
	population := make([]*genetic.Individual, size)
	for i := range population {
		chromosome := rng.Perm(len(candidates))
		fitness := e.EvaluateChromosomeWithAggregation(chromosome, candidates, affinityMatrix, teamSize, strategy)
		population[i] = &genetic.Individual{
			Chromosome: chromosome,
			Fitness:    genetic.FitnessResult{PsyScore: fitness, SocScore: fitness, IsFeasible: fitness > -1000},
		}
	}
	return population
}

// tournamentSelectWeightedSum performs tournament selection based on scalar fitness
func (e *TeamFormationEngine) tournamentSelectWeightedSum(population []*genetic.Individual, rng *rand.Rand) *genetic.Individual {
	a := population[rng.Intn(len(population))]
	b := population[rng.Intn(len(population))]
	if a.Fitness.PsyScore > b.Fitness.PsyScore {
		return a
	}
	return b
}

func (e *TeamFormationEngine) decodeChromosome(chromosome []int, candidates []TeamMember, teamSize int) []Team {
	numTeams := len(candidates) / teamSize
	teams := make([]Team, numTeams)

	for i := 0; i < numTeams; i++ {
		start := i * teamSize
		end := start + teamSize
		members := make([]TeamMember, 0, teamSize)
		for j := start; j < end && j < len(chromosome); j++ {
			members = append(members, candidates[chromosome[j]])
		}
		teams[i] = Team{Members: members}
	}

	return teams
}

func (e *TeamFormationEngine) calculateTeamScore(_ context.Context, members []TeamMember, affinityMatrix map[string]map[string]float64) float64 {
	return e.calculateTeamRelationalScoreFromMatrix(members, affinityMatrix)
}

// calculateTeamRelationalScoreFromMatrix computes S_rel(G) according to thesis formula:
// S_rel(G) = min(1, Σ_ij C'_ij / (|G|(|G|-1)/2))
// The affinityMatrix contains pre-computed C'_ij values using ComputeRelationalCost.
func (e *TeamFormationEngine) calculateTeamRelationalScoreFromMatrix(members []TeamMember, affinityMatrix map[string]map[string]float64) float64 {
	if len(members) <= 1 {
		return 0.0
	}

	// Collect all C'_ij values for pairs in this group
	relationalCosts := make([]float64, 0)
	for i := 0; i < len(members); i++ {
		for j := i + 1; j < len(members); j++ {
			cost := affinityMatrix[members[i].ID][members[j].ID]
			relationalCosts = append(relationalCosts, cost)
		}
	}

	if len(relationalCosts) == 0 {
		return 0.0
	}

	return ComputeGroupRelationalCohesion(relationalCosts, len(members))
}

// calculateTeamScoreFromMatrixLegacy provides backward compatibility for existing tests.
// Uses the old threshold-based approach instead of the thesis formula.
func (e *TeamFormationEngine) calculateTeamScoreFromMatrix(members []TeamMember, affinityMatrix map[string]map[string]float64) float64 {
	if len(members) <= 1 {
		return 0.5
	}

	totalScore, compatiblePairs, incompatiblePairs, comparisons := classifyPairs(members, affinityMatrix)

	if comparisons == 0 {
		return 0.5
	}

	avgScore := totalScore / float64(comparisons)
	return applyCompatibilityAdjustment(avgScore, compatiblePairs, incompatiblePairs, comparisons)
}

// classifyPairs iterates all member pairs and returns score totals and compatibility counts.
// DEPRECATED: Used only for backward compatibility. Use relational cost matrix instead.
func classifyPairs(members []TeamMember, affinityMatrix map[string]map[string]float64) (totalScore float64, compatible, incompatible, comparisons int) {
	for i := 0; i < len(members); i++ {
		for j := i + 1; j < len(members); j++ {
			pairScore := affinityMatrix[members[i].ID][members[j].ID]
			totalScore += pairScore
			comparisons++
			switch {
			case pairScore >= 0.9:
				compatible++
			case pairScore <= 0.1:
				incompatible++
			}
		}
	}
	return
}

// applyCompatibilityAdjustment applies penalties or bonuses based on pair classifications.
// DEPRECATED: Used only for backward compatibility (legacy scoring).
func applyCompatibilityAdjustment(avgScore float64, compatiblePairs, incompatiblePairs, comparisons int) float64 {
	if incompatiblePairs > 0 {
		incompatibilityRatio := float64(incompatiblePairs) / float64(comparisons)
		penalty := incompatibilityRatio * 2.0
		if penalty > 0.95 {
			penalty = 0.95
		}
		avgScore = avgScore*(1.0-penalty) - 0.2
		if avgScore < 0 {
			return 0
		}
		return avgScore
	}
	if compatiblePairs == comparisons && compatiblePairs > 0 {
		avgScore = avgScore * 1.15
		if avgScore > 1.0 {
			return 1.0
		}
	}
	return avgScore
}

func (e *TeamFormationEngine) orderCrossover(parent1, parent2 []int, rng *rand.Rand) []int {
	n := len(parent1)
	child := make([]int, n)
	for i := range child {
		child[i] = -1
	}

	// Seleccionar segmento aleatorio
	start := rng.Intn(n)
	end := start + rng.Intn(n-start)

	// Copiar segmento del parent1
	inChild := make(map[int]bool)
	for i := start; i <= end && i < n; i++ {
		child[i] = parent1[i]
		inChild[parent1[i]] = true
	}

	// Rellenar con parent2
	pos := (end + 1) % n
	for _, gene := range parent2 {
		if !inChild[gene] {
			child[pos] = gene
			pos = (pos + 1) % n
		}
	}

	return child
}

func (e *TeamFormationEngine) mutate(chromosome []int, rng *rand.Rand) {
	// Swap mutation
	i := rng.Intn(len(chromosome))
	j := rng.Intn(len(chromosome))
	chromosome[i], chromosome[j] = chromosome[j], chromosome[i]
}

// Helpers
func getUnassigned(candidates []TeamMember, assigned map[string]bool) []TeamMember {
	var unassigned []TeamMember
	for _, c := range candidates {
		if !assigned[c.ID] {
			unassigned = append(unassigned, c)
		}
	}
	return unassigned
}

// TeamSizeDistribution representa una posible distribución de tamaños de grupos
type TeamSizeDistribution struct {
	TeamSizes []int   // Tamaños de cada grupo
	Score     float64 // Score de compatibilidad global
}

// formTeamsIntelligent es el algoritmo principal que prueba diferentes configuraciones
// de tamaños de grupo y selecciona la que maximiza la compatibilidad global.
// Por ejemplo, con 5 estudiantes y tamaño deseado 3:
// - Prueba 3-2 (un grupo de 3 y uno de 2)
// - Prueba 2-2-1 (dos grupos de 2 y uno de 1)
// - Selecciona la configuración con mejor compatibilidad
func (e *TeamFormationEngine) formTeamsIntelligent(ctx context.Context, request TeamFormationRequest) (*TeamFormationResult, error) {
	numCandidates := len(request.Candidates)
	targetSize := request.TeamSize

	// Establecer límites de tamaño de grupo
	minSize := request.MinTeamSize
	if minSize <= 0 {
		minSize = 1 // Mínimo por defecto: 1 persona
	}
	maxSize := request.MaxTeamSize
	if maxSize <= 0 {
		maxSize = targetSize + 1 // Máximo por defecto: target + 1
	}

	// Si el número exacto de candidatos es divisible por el tamaño deseado,
	// usar directamente el algoritmo estándar
	if numCandidates%targetSize == 0 {
		return e.formTeamsWithDistribution(ctx, request, nil)
	}

	// Generar todas las posibles distribuciones de tamaños de grupo
	distributions := generateDistributions(numCandidates, targetSize, minSize, maxSize)

	if len(distributions) == 0 {
		// Si no hay distribuciones válidas, usar el algoritmo estándar
		return e.formTeamsWithDistribution(ctx, request, nil)
	}

	log.Printf("Evaluating %d different team size distributions for %d candidates (target size: %d)",
		len(distributions), numCandidates, targetSize)

	// Precalcular matriz de afinidad (se reutilizará para todas las distribuciones)
	affinityMatrix := e.precomputeAffinityMatrix(ctx, request.Candidates)

	// Evaluar cada distribución y encontrar la mejor
	var bestResult *TeamFormationResult
	var bestScore float64 = -2.0 // Mínimo posible
	var bestDistribution []int

	for _, dist := range distributions {
		result, err := e.evaluateDistribution(ctx, request, dist, affinityMatrix)
		if err != nil {
			log.Printf("Warning: Failed to evaluate distribution %v: %v", dist, err)
			continue
		}

		// Log detallado con scores de cada equipo
		teamScores := make([]string, len(result.Teams))
		for i, team := range result.Teams {
			teamScores[i] = fmt.Sprintf("T%d(%.2f)", i+1, team.AffinityScore)
		}
		log.Printf("Distribution %v -> Global: %.4f | Teams: %v", dist, result.GlobalScore, teamScores)

		if result.GlobalScore > bestScore {
			bestScore = result.GlobalScore
			bestResult = result
			bestDistribution = dist
		}
	}

	if bestResult == nil {
		// Fallback al algoritmo estándar
		return e.formTeamsWithDistribution(ctx, request, nil)
	}

	// Añadir metadata sobre la distribución elegida
	if bestResult.Metadata == nil {
		bestResult.Metadata = make(map[string]interface{})
	}
	bestResult.Metadata["chosen_distribution"] = bestDistribution
	bestResult.Metadata["distributions_evaluated"] = len(distributions)
	bestResult.Metadata["team_details"] = buildTeamDetails(bestResult.Teams, affinityMatrix)

	log.Printf("Best distribution: %v with score %.4f", bestDistribution, bestScore)

	return bestResult, nil
}

// buildTeamDetails returns per-team compatibility metadata for the intelligent result.
func buildTeamDetails(teams []Team, affinityMatrix map[string]map[string]float64) []map[string]interface{} {
	details := make([]map[string]interface{}, len(teams))
	for i, team := range teams {
		compatible, incompatible := countCompatiblePairs(team.Members, affinityMatrix)
		details[i] = map[string]interface{}{
			"team_id":            team.ID,
			"size":               len(team.Members),
			"affinity_score":     team.AffinityScore,
			"compatible_pairs":   compatible,
			"incompatible_pairs": incompatible,
		}
	}
	return details
}

// countCompatiblePairs counts pairs with affinity >= 0.5 and pairs below that threshold.
func countCompatiblePairs(members []TeamMember, affinityMatrix map[string]map[string]float64) (compatible, incompatible int) {
	for x := 0; x < len(members); x++ {
		for y := x + 1; y < len(members); y++ {
			if affinityMatrix[members[x].ID][members[y].ID] >= 0.5 {
				compatible++
			} else {
				incompatible++
			}
		}
	}
	return
}

// generateDistributions genera todas las posibles distribuciones de tamaños de grupo
// que sumen exactamente numCandidates, donde cada grupo tiene tamaño entre minSize y maxSize.
func generateDistributions(numCandidates, targetSize, minSize, maxSize int) [][]int {
	result := collectDistributions(numCandidates, targetSize, minSize, maxSize)

	sort.Slice(result, func(i, j int) bool {
		return compareDistributionPreference(result[i], result[j], targetSize)
	})

	const maxDistributions = 10
	if len(result) > maxDistributions {
		result = result[:maxDistributions]
	}
	return result
}

// collectDistributions enumerates all valid size distributions summing to numCandidates.
func collectDistributions(numCandidates, targetSize, minSize, maxSize int) [][]int {
	var result [][]int
	var current []int

	var generate func(remaining int)
	generate = func(remaining int) {
		if remaining == 0 {
			dist := make([]int, len(current))
			copy(dist, current)
			result = append(result, dist)
			return
		}
		for _, size := range generateSizePriorities(remaining, targetSize, minSize, maxSize) {
			if size > remaining || size < minSize || size > maxSize {
				continue
			}
			if len(current) > 0 && size > current[len(current)-1] {
				continue
			}
			current = append(current, size)
			generate(remaining - size)
			current = current[:len(current)-1]
		}
	}
	generate(numCandidates)
	return result
}

// compareDistributionPreference returns true if distribution a is preferred over b.
func compareDistributionPreference(a, b []int, targetSize int) bool {
	singletonsA := countSingletons(a)
	singletonsB := countSingletons(b)
	if singletonsA != singletonsB {
		return singletonsA < singletonsB
	}
	varianceA := calculateSizeVariance(a)
	varianceB := calculateSizeVariance(b)
	if varianceA != varianceB {
		return varianceA < varianceB
	}
	return calculateTargetDistance(a, targetSize) < calculateTargetDistance(b, targetSize)
}

// countSingletons cuenta el número de grupos de 1 persona
func countSingletons(dist []int) int {
	count := 0
	for _, size := range dist {
		if size == 1 {
			count++
		}
	}
	return count
}

// calculateSizeVariance calcula la varianza de los tamaños de grupo
func calculateSizeVariance(dist []int) float64 {
	if len(dist) <= 1 {
		return 0
	}
	var sum float64
	for _, size := range dist {
		sum += float64(size)
	}
	mean := sum / float64(len(dist))

	var variance float64
	for _, size := range dist {
		diff := float64(size) - mean
		variance += diff * diff
	}
	return variance / float64(len(dist))
}

// calculateTargetDistance calcula la distancia promedio de los tamaños al target
func calculateTargetDistance(dist []int, target int) float64 {
	if len(dist) == 0 {
		return 0
	}
	var totalDist float64
	for _, size := range dist {
		diff := float64(size - target)
		if diff < 0 {
			diff = -diff
		}
		totalDist += diff
	}
	return totalDist / float64(len(dist))
}

// generateSizePriorities genera los tamaños a probar en orden de prioridad
// (primero los más cercanos al target)
func generateSizePriorities(remaining, targetSize, minSize, maxSize int) []int {
	var sizes []int
	seen := make(map[int]bool)

	// Primero el tamaño target
	if targetSize >= minSize && targetSize <= maxSize && targetSize <= remaining {
		sizes = append(sizes, targetSize)
		seen[targetSize] = true
	}

	// Luego tamaños cercanos al target
	for delta := 1; delta <= maxSize; delta++ {
		// Target - delta
		smaller := targetSize - delta
		if smaller >= minSize && smaller <= remaining && !seen[smaller] {
			sizes = append(sizes, smaller)
			seen[smaller] = true
		}
		// Target + delta
		larger := targetSize + delta
		if larger <= maxSize && larger <= remaining && !seen[larger] {
			sizes = append(sizes, larger)
			seen[larger] = true
		}
	}

	return sizes
}

// precomputeAffinityMatrix precalcula la matriz de afinidad para todos los pares de candidatos
func (e *TeamFormationEngine) precomputeAffinityMatrix(ctx context.Context, candidates []TeamMember) map[string]map[string]float64 {
	affinityMatrix := make(map[string]map[string]float64)
	for _, c := range candidates {
		affinityMatrix[c.ID] = make(map[string]float64)
	}

	log.Printf("=== Computing Affinity Matrix for %d candidates ===", len(candidates))
	compatiblePairs := 0
	incompatiblePairs := 0

	for i, c1 := range candidates {
		for j := i + 1; j < len(candidates); j++ {
			c2 := candidates[j]
			score, providerScores, err := e.CalculateCombinedAffinity(ctx, c1.ID, c2.ID)
			if err != nil {
				log.Printf("Warning: could not calculate affinity for (%s, %s): %v", c1.ID, c2.ID, err)
				score = 0
			}
			affinityMatrix[c1.ID][c2.ID] = score
			affinityMatrix[c2.ID][c1.ID] = score

			// Log detallado de cada par
			if score >= 0.5 {
				compatiblePairs++
				log.Printf("  Pair (%s, %s): %.2f [COMPATIBLE] %v", c1.ID, c2.ID, score, providerScores)
			} else {
				incompatiblePairs++
				log.Printf("  Pair (%s, %s): %.2f [INCOMPATIBLE] %v", c1.ID, c2.ID, score, providerScores)
			}
		}
	}

	log.Printf("=== Affinity Matrix Summary: %d compatible pairs, %d incompatible pairs ===",
		compatiblePairs, incompatiblePairs)

	return affinityMatrix
}

// evaluateDistribution evalúa una distribución específica de tamaños de grupo
// usando el algoritmo genético para encontrar la mejor asignación
func (e *TeamFormationEngine) evaluateDistribution(
	_ context.Context,
	request TeamFormationRequest,
	distribution []int,
	affinityMatrix map[string]map[string]float64,
) (*TeamFormationResult, error) {
	candidates := request.Candidates
	params := resolveGeneticParams(e.config.GeneticParams)

	// Inicializar población
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	population := make([]*genetic.Individual, params.PopulationSize)
	for i := range population {
		chromosome := rng.Perm(len(candidates))
		population[i] = &genetic.Individual{
			Chromosome: chromosome,
			Fitness:    e.evaluateChromosomeWithDistribution(chromosome, candidates, affinityMatrix, distribution),
		}
	}

	evalFn := func(c []int) genetic.FitnessResult {
		return e.evaluateChromosomeWithDistribution(c, candidates, affinityMatrix, distribution)
	}
	finalFronts := e.Nsga2Evolve(population, params, rng, evalFn, nil)

	if len(finalFronts) == 0 {
		return nil, fmt.Errorf("genetic algorithm failed")
	}
	paretoFront := finalFronts[0]
	bestInd := e.selectKneePoint(paretoFront)
	if bestInd == nil {
		bestInd = paretoFront[0]
	}

	// Decodificar solución con la distribución específica
	teams := e.decodeChromosomeWithDistribution(bestInd.Chromosome, candidates, distribution)
	isFeasible := bestInd.Fitness.IsFeasible

	// Calcular scores de cada equipo
	for i := range teams {
		oceanVectors := getOCEANVectors(teams[i].Members)
		psy := domain.CalculateGroupFitness(oceanVectors)
		soc := e.calculateTeamScoreFromMatrix(teams[i].Members, affinityMatrix)

		teams[i].AffinityScore = (psy + soc) / 2.0
		teams[i].ID = fmt.Sprintf(teamIDFmt, i+1)
		teams[i].Metadata = TeamMetadata{
			TotalComparisons:  len(teams[i].Members) * (len(teams[i].Members) - 1) / 2,
			PsychometricScore: psy,
			SociometricScore:  soc,
			IsFeasible:        true,
		}
	}

	providersUsed, weights := e.collectProvidersInfo()

	finalGlobalScore := (bestInd.Fitness.PsyScore + bestInd.Fitness.SocScore) / 2.0
	if !isFeasible {
		finalGlobalScore = -1000.0
	}

	return &TeamFormationResult{
		Teams:         teams,
		GlobalScore:   finalGlobalScore,
		Algorithm:     "genetic_intelligent_nsga2",
		ProvidersUsed: providersUsed,
		Weights:       weights,
		Metadata: map[string]interface{}{
			"pareto_front_size": len(paretoFront),
			"is_feasible":       isFeasible,
			"fitness_psy":       bestInd.Fitness.PsyScore,
			"fitness_soc":       bestInd.Fitness.SocScore,
		},
	}, nil
}

// evaluateChromosomeWithDistribution evalúa un cromosoma usando una distribución específica de tamaños
func (e *TeamFormationEngine) evaluateChromosomeWithDistribution(
	chromosome []int,
	candidates []TeamMember,
	affinityMatrix map[string]map[string]float64,
	distribution []int,
) genetic.FitnessResult {
	teams := e.decodeChromosomeWithDistribution(chromosome, candidates, distribution)

	if len(teams) == 0 {
		return genetic.FitnessResult{IsFeasible: false}
	}

	var totalPsyScore float64
	var totalSocScore float64

	for _, team := range teams {
		oceanVectors := getOCEANVectors(team.Members)
		psy := domain.CalculateGroupFitness(oceanVectors)
		totalPsyScore += psy

		soc := e.calculateTeamScoreFromMatrix(team.Members, affinityMatrix)
		totalSocScore += soc
	}

	avgPsy := totalPsyScore / float64(len(teams))
	avgSoc := totalSocScore / float64(len(teams))

	return genetic.FitnessResult{
		PsyScore:   avgPsy,
		SocScore:   avgSoc,
		IsFeasible: true,
	}
}

// decodeChromosomeWithDistribution decodifica un cromosoma usando una distribución específica
func (e *TeamFormationEngine) decodeChromosomeWithDistribution(
	chromosome []int,
	candidates []TeamMember,
	distribution []int,
) []Team {
	teams := make([]Team, len(distribution))
	pos := 0

	for i, teamSize := range distribution {
		members := make([]TeamMember, 0, teamSize)
		for j := 0; j < teamSize && pos < len(chromosome); j++ {
			if chromosome[pos] < len(candidates) {
				members = append(members, candidates[chromosome[pos]])
			}
			pos++
		}
		teams[i] = Team{Members: members}
	}

	return teams
}

// formTeamsWithDistribution forma equipos usando una distribución específica o el algoritmo estándar
func (e *TeamFormationEngine) formTeamsWithDistribution(
	ctx context.Context,
	request TeamFormationRequest,
	distribution []int,
) (*TeamFormationResult, error) {
	if distribution == nil {
		// Usar algoritmo estándar
		switch e.config.Algorithm {
		case "genetic":
			return e.formTeamsGenetic(ctx, request)
		case "greedy":
			fallthrough
		default:
			return e.formTeamsGreedy(ctx, request)
		}
	}

	// Usar algoritmo con distribución específica
	affinityMatrix := e.precomputeAffinityMatrix(ctx, request.Candidates)
	return e.evaluateDistribution(ctx, request, distribution, affinityMatrix)
}

// Helpers used by NSGA-II

func getOCEANVectors(members []TeamMember) []domain.OCEANVector {
	vectors := make([]domain.OCEANVector, 0, len(members))
	for _, m := range members {
		if m.Psychometrics != nil {
			vectors = append(vectors, *m.Psychometrics)
		}
	}
	return vectors
}

func (e *TeamFormationEngine) tournamentSelectNSGA2(population []*genetic.Individual, rng *rand.Rand) *genetic.Individual {
	a := population[rng.Intn(len(population))]
	b := population[rng.Intn(len(population))]
	if genetic.CrowdedComparison(a.Fitness, b.Fitness) {
		return a
	}
	return b
}

func (e *TeamFormationEngine) tournamentSelectNoCrowding(population []*genetic.Individual, rng *rand.Rand) *genetic.Individual {
	a := population[rng.Intn(len(population))]
	b := population[rng.Intn(len(population))]

	if a.Fitness.Rank < b.Fitness.Rank {
		return a
	}
	if b.Fitness.Rank < a.Fitness.Rank {
		return b
	}

	if rng.Float64() < 0.5 {
		return a
	}
	return b
}

func (e *TeamFormationEngine) selectKneePoint(front []*genetic.Individual) *genetic.Individual {
	if len(front) == 0 {
		return nil
	}
	var bestDist float64 = math.MaxFloat64
	var bestInd *genetic.Individual

	// Ideal point is (1.0, 1.0) for maximization
	for _, ind := range front {
		d := math.Sqrt(math.Pow(ind.Fitness.PsyScore-1.0, 2) + math.Pow(ind.Fitness.SocScore-1.0, 2))
		if d < bestDist {
			bestDist = d
			bestInd = ind
		}
	}
	return bestInd
}

// Added back tournamentSelectStandard for use in evaluateDistribution (standard GA)
func (e *TeamFormationEngine) tournamentSelectStandard(fitness []float64, rng *rand.Rand) int {
	a := rng.Intn(len(fitness))
	b := rng.Intn(len(fitness))
	if fitness[a] > fitness[b] {
		return a
	}
	return b
}

// ---------------------------------------------------------------------------
// Plugin support — pre-computed affinity matrix
// ---------------------------------------------------------------------------

// mapAffinityProvider wraps a pre-computed affinity matrix as an AffinityProvider.
// Used by the team_formation plugin to run optimization without calling back to the kernel.
type mapAffinityProvider struct {
	matrix map[string]map[string]float64
}

func (p *mapAffinityProvider) ProviderID() string   { return "precomputed" }
func (p *mapAffinityProvider) ProviderName() string { return "Pre-computed Affinity" }
func (p *mapAffinityProvider) CalculatePairwiseAffinity(_ context.Context, a, b string) (float64, error) {
	if row, ok := p.matrix[a]; ok {
		if score, ok2 := row[b]; ok2 {
			return score, nil
		}
	}
	return 0.5, nil
}
func (p *mapAffinityProvider) SupportsContext(_ context.Context) bool { return true }
func (p *mapAffinityProvider) GetMetadata() ProviderMetadata {
	return ProviderMetadata{ID: "precomputed", Name: "Pre-computed Affinity", Category: "computed"}
}

// PrecomputeAffinityMatrix computes all pairwise affinity scores for the given candidates.
// Returns a symmetric map[studentA][studentB]float64. Safe to call concurrently.
// The result can be passed to the team_formation plugin, avoiding repeated RPC calls
// during optimization.
func (e *TeamFormationEngine) PrecomputeAffinityMatrix(ctx context.Context, candidates []TeamMember) map[string]map[string]float64 {
	return e.precomputeAffinityMatrix(ctx, candidates)
}

// OptimizeWithMatrix runs the team formation algorithm on a pre-computed affinity matrix.
// This is the entry point called by the team_formation plugin subprocess.
// It bypasses AffinityProvider RPC entirely — the matrix must already contain all scores.
func OptimizeWithMatrix(ctx context.Context, config *TeamFormationConfig, request TeamFormationRequest, matrix map[string]map[string]float64) (*TeamFormationResult, error) {
	if config == nil {
		config = DefaultTeamFormationConfig()
	}
	e := &TeamFormationEngine{
		config: config,
		providers: []WeightedProvider{{
			Provider: &mapAffinityProvider{matrix: matrix},
			Weight:   1.0,
		}},
	}
	return e.formTeamsIntelligent(ctx, request)
}
