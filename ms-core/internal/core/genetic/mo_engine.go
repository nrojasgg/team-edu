package genetic

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"ms-core/internal/core/domain"
)

// MOIndividual wraps a domain.Chromosome with NSGA-II multi-objective attributes.
type MOIndividual struct {
	Chromosome *domain.Chromosome
	Fitness    FitnessResult

	dominationCount int
	dominatedBy     []*MOIndividual
}

// MOEngineConfig holds configuration for the multi-objective NSGA-II engine.
type MOEngineConfig struct {
	PopulationSize  int
	Generations     int
	ElitismRate     float64
	MutationRate    float64
	CrossoverRate   float64
	TournamentSize  int
	TargetGroupSize int
	Seed            int64
	MaxStagnation   int
}

// DefaultMOEngineConfig returns sensible defaults for the MO engine.
func DefaultMOEngineConfig() MOEngineConfig {
	return MOEngineConfig{
		PopulationSize:  150,
		Generations:     100,
		ElitismRate:     0.15,
		MutationRate:    0.2,
		CrossoverRate:   0.8,
		TournamentSize:  2,
		TargetGroupSize: 4,
		Seed:            time.Now().UnixNano(),
		MaxStagnation:   40,
	}
}

// MOResult holds the result of a multi-objective optimization run.
type MOResult struct {
	ParetoFront     []*MOIndividual
	AllFronts       [][]*MOIndividual
	Stats           *EvolutionStats
	ExecutionTimeMs int64
}

// MOEngine runs NSGA-II with criteria-based multi-objective evaluation.
type MOEngine struct {
	config     MOEngineConfig
	factory    *CriterionFactory
	moConfig   MultiObjectiveConfig
	domainConf domain.Config
	ctx        domain.Context
	rng        *rand.Rand
}

// NewMOEngine creates a new multi-objective NSGA-II engine.
func NewMOEngine(config MOEngineConfig, factory *CriterionFactory, moConfig MultiObjectiveConfig, domainConf domain.Config, ctx domain.Context) *MOEngine {
	return &MOEngine{
		config:     config,
		factory:    factory,
		moConfig:   moConfig,
		domainConf: domainConf,
		ctx:        ctx,
		rng:        rand.New(rand.NewSource(config.Seed)),
	}
}

// Run executes the NSGA-II algorithm and returns the Pareto front.
func (e *MOEngine) Run(goCtx context.Context, students []*domain.Student) (*MOResult, error) {
	if len(students) == 0 {
		return nil, fmt.Errorf("cannot run with zero students")
	}

	startTime := time.Now()
	stats := &EvolutionStats{
		Generations: make([]GenerationStats, 0, e.config.Generations),
	}

	population, err := e.initMOPopulation(students)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MO population: %w", err)
	}

	e.evaluateMOPopulation(population)

	stagnationCount := 0
	bestHypervolume := 0.0

	for gen := 0; gen < e.config.Generations; gen++ {
		select {
		case <-goCtx.Done():
			stats.StoppedEarly = true
			stats.StopReason = "context cancelled"
			stats.TotalTime = time.Since(startTime)
			fronts := e.fastNonDominatedSortMO(population)
			return &MOResult{
				ParetoFront:     fronts[0],
				AllFronts:       fronts,
				Stats:           stats,
				ExecutionTimeMs: stats.TotalTime.Milliseconds(),
			}, goCtx.Err()
		default:
		}

		fronts := e.fastNonDominatedSortMO(population)
		e.assignCrowdingDistance(fronts)

		currentHV := e.calculateHypervolume(fronts[0])
		if currentHV > bestHypervolume {
			bestHypervolume = currentHV
			stagnationCount = 0
		} else {
			stagnationCount++
		}

		genStats := GenerationStats{
			Generation:  gen,
			BestFitness: currentHV,
			ElapsedTime: time.Since(startTime),
		}
		stats.Generations = append(stats.Generations, genStats)

		if e.config.MaxStagnation > 0 && stagnationCount >= e.config.MaxStagnation {
			stats.StoppedEarly = true
			stats.StopReason = fmt.Sprintf("stagnation for %d generations", e.config.MaxStagnation)
			break
		}

		if gen == e.config.Generations-1 {
			break
		}

		offspring := e.evolveMOGeneration(population, students)
		e.evaluateMOPopulation(offspring)

		combined := append(population, offspring...)
		combinedFronts := e.fastNonDominatedSortMO(combined)
		e.assignCrowdingDistance(combinedFronts)

		population = e.selectSurvivors(combinedFronts)
	}

	stats.TotalTime = time.Since(startTime)
	stats.FinalBestFitness = bestHypervolume

	finalFronts := e.fastNonDominatedSortMO(population)
	e.assignCrowdingDistance(finalFronts)

	paretoFront := finalFronts[0]
	sort.Slice(paretoFront, func(i, j int) bool {
		return paretoFront[i].Fitness.PsyScore > paretoFront[j].Fitness.PsyScore
	})

	return &MOResult{
		ParetoFront:     paretoFront,
		AllFronts:       finalFronts,
		Stats:           stats,
		ExecutionTimeMs: stats.TotalTime.Milliseconds(),
	}, nil
}

func (e *MOEngine) initMOPopulation(students []*domain.Student) ([]*MOIndividual, error) {
	chromosomes, err := InitPopulation(students, e.config.TargetGroupSize, e.config.PopulationSize, e.config.Seed)
	if err != nil {
		return nil, err
	}

	population := make([]*MOIndividual, len(chromosomes))
	for i, chr := range chromosomes {
		population[i] = &MOIndividual{
			Chromosome: chr,
			Fitness:    FitnessResult{IsFeasible: true},
		}
	}
	return population, nil
}

func (e *MOEngine) evaluateMOPopulation(population []*MOIndividual) {
	for _, ind := range population {
		ind.Fitness = CalculateMultiObjectiveFitness(
			ind.Chromosome, e.moConfig, e.factory, e.ctx, e.domainConf.UseQuadraticPenalty,
		)
	}
}

func (e *MOEngine) fastNonDominatedSortMO(population []*MOIndividual) [][]*MOIndividual {
	fronts := make([][]*MOIndividual, 0)
	front1 := make([]*MOIndividual, 0)

	for _, p := range population {
		p.dominationCount = 0
		p.dominatedBy = make([]*MOIndividual, 0)
	}

	for _, p := range population {
		for _, q := range population {
			if p == q {
				continue
			}
			if e.moDominates(p, q) {
				p.dominatedBy = append(p.dominatedBy, q)
			} else if e.moDominates(q, p) {
				p.dominationCount++
			}
		}
		if p.dominationCount == 0 {
			p.Fitness.Rank = 1
			front1 = append(front1, p)
		}
	}

	fronts = append(fronts, front1)

	for i := 0; i < len(fronts); i++ {
		nextFront := make([]*MOIndividual, 0)
		for _, p := range fronts[i] {
			for _, q := range p.dominatedBy {
				q.dominationCount--
				if q.dominationCount == 0 {
					q.Fitness.Rank = i + 2
					nextFront = append(nextFront, q)
				}
			}
		}
		if len(nextFront) == 0 {
			break
		}
		fronts = append(fronts, nextFront)
	}

	return fronts
}

func (e *MOEngine) moDominates(a, b *MOIndividual) bool {
	if !a.Fitness.IsFeasible && b.Fitness.IsFeasible {
		return false
	}
	if a.Fitness.IsFeasible && !b.Fitness.IsFeasible {
		return true
	}

	betterInOne := (a.Fitness.PsyScore > b.Fitness.PsyScore) || (a.Fitness.SocScore > b.Fitness.SocScore)
	notWorseInAny := (a.Fitness.PsyScore >= b.Fitness.PsyScore) && (a.Fitness.SocScore >= b.Fitness.SocScore)

	return betterInOne && notWorseInAny
}

func (e *MOEngine) assignCrowdingDistance(fronts [][]*MOIndividual) {
	for _, front := range fronts {
		l := len(front)
		if l == 0 {
			continue
		}

		for _, ind := range front {
			ind.Fitness.Distance = 0
		}

		sort.Slice(front, func(i, j int) bool {
			return front[i].Fitness.PsyScore < front[j].Fitness.PsyScore
		})
		front[0].Fitness.Distance = 1e9
		front[l-1].Fitness.Distance = 1e9

		rangePsy := front[l-1].Fitness.PsyScore - front[0].Fitness.PsyScore
		if rangePsy > 0 {
			for i := 1; i < l-1; i++ {
				front[i].Fitness.Distance += (front[i+1].Fitness.PsyScore - front[i-1].Fitness.PsyScore) / rangePsy
			}
		}

		sort.Slice(front, func(i, j int) bool {
			return front[i].Fitness.SocScore < front[j].Fitness.SocScore
		})
		front[0].Fitness.Distance = 1e9
		front[l-1].Fitness.Distance = 1e9

		rangeSoc := front[l-1].Fitness.SocScore - front[0].Fitness.SocScore
		if rangeSoc > 0 {
			for i := 1; i < l-1; i++ {
				front[i].Fitness.Distance += (front[i+1].Fitness.SocScore - front[i-1].Fitness.SocScore) / rangeSoc
			}
		}
	}
}

func (e *MOEngine) calculateHypervolume(front []*MOIndividual) float64 {
	if len(front) == 0 {
		return 0.0
	}

	sorted := make([]*MOIndividual, len(front))
	copy(sorted, front)
	// Sort ascending by PsyScore for correct dominated hypervolume in maximization
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Fitness.PsyScore < sorted[j].Fitness.PsyScore
	})

	hv := 0.0
	prevPsy := 0.0 // Reference point (nadir) for maximization
	for _, ind := range sorted {
		psy := ind.Fitness.PsyScore
		soc := ind.Fitness.SocScore
		width := psy - prevPsy
		hv += width * soc
		prevPsy = psy
	}

	return hv
}

func (e *MOEngine) crowdedTournament(population []*MOIndividual) *MOIndividual {
	a := population[e.rng.Intn(len(population))]
	b := population[e.rng.Intn(len(population))]

	if a.Fitness.Rank < b.Fitness.Rank {
		return a
	}
	if a.Fitness.Rank > b.Fitness.Rank {
		return b
	}
	if a.Fitness.Distance > b.Fitness.Distance {
		return a
	}
	return b
}

func (e *MOEngine) evolveMOGeneration(population []*MOIndividual, students []*domain.Student) []*MOIndividual {
	nextGen := make([]*MOIndividual, 0, e.config.PopulationSize)

	eliteCount := int(float64(e.config.PopulationSize) * e.config.ElitismRate)
	if eliteCount < 1 {
		eliteCount = 1
	}

	sorted := make([]*MOIndividual, len(population))
	copy(sorted, population)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Fitness.Rank != sorted[j].Fitness.Rank {
			return sorted[i].Fitness.Rank < sorted[j].Fitness.Rank
		}
		return sorted[i].Fitness.Distance > sorted[j].Fitness.Distance
	})

	for i := 0; i < eliteCount && i < len(sorted); i++ {
		nextGen = append(nextGen, &MOIndividual{
			Chromosome: deepCopyChromosome(sorted[i].Chromosome),
			Fitness:    sorted[i].Fitness,
		})
	}

	for len(nextGen) < e.config.PopulationSize {
		parent1 := e.crowdedTournament(population)
		parent2 := e.crowdedTournament(population)

		var offspring *domain.Chromosome
		if e.rng.Float64() < e.config.CrossoverRate {
			offspring = e.moCrossover(parent1.Chromosome, parent2.Chromosome, students)
		} else {
			offspring = deepCopyChromosome(parent1.Chromosome)
		}

		if e.rng.Float64() < e.config.MutationRate {
			e.moMutation(offspring)
		}

		nextGen = append(nextGen, &MOIndividual{
			Chromosome: offspring,
			Fitness:    FitnessResult{IsFeasible: true},
		})
	}

	if len(nextGen) > e.config.PopulationSize {
		nextGen = nextGen[:e.config.PopulationSize]
	}

	return nextGen
}

func (e *MOEngine) moCrossover(parent1, parent2 *domain.Chromosome, students []*domain.Student) *domain.Chromosome {
	offspring := &domain.Chromosome{
		Groups: make([]domain.Group, len(parent1.Groups)),
		Stats:  make(map[string]interface{}),
	}

	parent2Order := make([]*domain.Student, 0, len(students))
	for _, group := range parent2.Groups {
		parent2Order = append(parent2Order, group...)
	}

	if e.rng.Float64() < 0.3 {
		for i := len(parent2Order) - 1; i > 0; i-- {
			j := e.rng.Intn(i + 1)
			parent2Order[i], parent2Order[j] = parent2Order[j], parent2Order[i]
		}
	}

	index := 0
	for i, group := range parent1.Groups {
		offspring.Groups[i] = make(domain.Group, len(group))
		for j := 0; j < len(group); j++ {
			if index < len(parent2Order) {
				offspring.Groups[i][j] = parent2Order[index]
				index++
			}
		}
	}

	return offspring
}

func (e *MOEngine) moMutation(chromosome *domain.Chromosome) {
	if len(chromosome.Groups) < 2 {
		return
	}

	mutationType := e.rng.Intn(2)

	switch mutationType {
	case 0:
		g1 := e.rng.Intn(len(chromosome.Groups))
		g2 := e.rng.Intn(len(chromosome.Groups))
		if g1 == g2 {
			g2 = (g2 + 1) % len(chromosome.Groups)
		}
		if len(chromosome.Groups[g1]) > 0 && len(chromosome.Groups[g2]) > 0 {
			i1 := e.rng.Intn(len(chromosome.Groups[g1]))
			i2 := e.rng.Intn(len(chromosome.Groups[g2]))
			chromosome.Groups[g1][i1], chromosome.Groups[g2][i2] = chromosome.Groups[g2][i2], chromosome.Groups[g1][i1]
		}

	case 1:
		g := e.rng.Intn(len(chromosome.Groups))
		if len(chromosome.Groups[g]) >= 2 {
			i1 := e.rng.Intn(len(chromosome.Groups[g]))
			i2 := e.rng.Intn(len(chromosome.Groups[g]))
			if i1 != i2 {
				chromosome.Groups[g][i1], chromosome.Groups[g][i2] = chromosome.Groups[g][i2], chromosome.Groups[g][i1]
			}
		}
	}
}

func (e *MOEngine) selectSurvivors(fronts [][]*MOIndividual) []*MOIndividual {
	survivors := make([]*MOIndividual, 0, e.config.PopulationSize)

	for _, front := range fronts {
		if len(survivors)+len(front) <= e.config.PopulationSize {
			survivors = append(survivors, front...)
		} else {
			remaining := e.config.PopulationSize - len(survivors)
			sort.Slice(front, func(i, j int) bool {
				return front[i].Fitness.Distance > front[j].Fitness.Distance
			})
			survivors = append(survivors, front[:remaining]...)
			break
		}
	}

	return survivors
}
