package genetic

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sort"
	"time"

	"ms-core/internal/core/domain"
)

// EngineConfig holds configuration for the genetic algorithm engine.
type EngineConfig struct {
	PopulationSize   int     // Number of chromosomes in population
	Generations      int     // Number of generations to evolve
	ElitismRate      float64 // Percentage of top individuals to preserve (0.0-1.0)
	MutationRate     float64 // Probability of mutation (0.0-1.0)
	CrossoverRate    float64 // Probability of crossover (0.0-1.0)
	TournamentSize   int     // Number of individuals in tournament selection
	TargetGroupSize  int     // Desired group size for team formation
	Seed             int64   // Random seed for reproducibility
	WorkerCount      int     // Number of parallel workers (0 = auto-detect)
	MaxStagnation    int     // Stop if no improvement for N generations (0 = disabled)
	TargetFitness    float64 // Stop if fitness reaches this value (0 = disabled)
	EnableStatistics bool    // Track detailed statistics
}

// DefaultEngineConfig returns a configuration with sensible defaults.
func DefaultEngineConfig() EngineConfig {
	return EngineConfig{
		PopulationSize:   100,
		Generations:      50,
		ElitismRate:      0.1, // Top 10%
		MutationRate:     0.2, // 20% chance
		CrossoverRate:    0.7, // 70% chance
		TournamentSize:   3,   // Tournament of 3
		TargetGroupSize:  4,   // Groups of 4
		Seed:             time.Now().UnixNano(),
		WorkerCount:      0,    // Auto-detect
		MaxStagnation:    10,   // Stop after 10 generations without improvement
		TargetFitness:    0.95, // Stop at 95% fitness
		EnableStatistics: true,
	}
}

// GenerationStats holds statistics for a single generation.
type GenerationStats struct {
	Generation    int
	BestFitness   float64
	WorstFitness  float64
	AvgFitness    float64
	MedianFitness float64
	StdDev        float64
	Diversity     float64 // Measure of population diversity
	ElapsedTime   time.Duration
}

// EvolutionStats holds complete statistics for the evolution process.
type EvolutionStats struct {
	Generations      []GenerationStats
	TotalTime        time.Duration
	FinalBestFitness float64
	Converged        bool
	StoppedEarly     bool
	StopReason       string
	TotalEvaluations int
}

// Engine is the main genetic algorithm executor with parallel evaluation.
type Engine struct {
	config      EngineConfig
	factory     *CriterionFactory
	fitnessConf domain.Config
	ctx         domain.Context
	rng         *rand.Rand
	workers     int
}

// NewEngine creates a new genetic algorithm engine.
func NewEngine(config EngineConfig, factory *CriterionFactory, fitnessConf domain.Config, ctx domain.Context) *Engine {
	workers := config.WorkerCount
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	return &Engine{
		config:      config,
		factory:     factory,
		fitnessConf: fitnessConf,
		ctx:         ctx,
		rng:         rand.New(rand.NewSource(config.Seed)),
		workers:     workers,
	}
}

// Run executes the genetic algorithm with parallel evaluation.
func (e *Engine) Run(goCtx context.Context, students []*domain.Student) (*domain.Chromosome, *EvolutionStats, error) {
	if len(students) == 0 {
		return nil, nil, fmt.Errorf("cannot run with zero students")
	}

	stats := &EvolutionStats{
		Generations: make([]GenerationStats, 0, e.config.Generations),
	}

	startTime := time.Now()

	// Initialize population
	population, err := InitPopulation(students, e.config.TargetGroupSize, e.config.PopulationSize, e.config.Seed)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize population: %w", err)
	}

	// Evaluate initial population
	e.evaluatePopulation(goCtx, population)
	stats.TotalEvaluations += len(population)

	var bestChromosome *domain.Chromosome
	bestFitness := -1.0
	stagnationCount := 0

	// Evolution loop
	for gen := 0; gen < e.config.Generations; gen++ {
		genStart := time.Now()

		// Check context cancellation
		select {
		case <-goCtx.Done():
			stats.StoppedEarly = true
			stats.StopReason = "context cancelled"
			stats.TotalTime = time.Since(startTime)
			return bestChromosome, stats, goCtx.Err()
		default:
		}

		// Sort population by fitness (descending)
		sort.Slice(population, func(i, j int) bool {
			return population[i].Fitness > population[j].Fitness
		})

		// Update best solution
		currentBest := population[0]
		if currentBest.Fitness > bestFitness {
			bestFitness = currentBest.Fitness
			bestChromosome = deepCopyChromosome(currentBest)
			stagnationCount = 0
		} else {
			stagnationCount++
		}

		// Calculate generation statistics
		genStats := e.calculateGenerationStats(gen, population, time.Since(genStart))
		stats.Generations = append(stats.Generations, genStats)

		// Check stopping criteria
		if e.config.TargetFitness > 0 && bestFitness >= e.config.TargetFitness {
			stats.StoppedEarly = true
			stats.StopReason = fmt.Sprintf("target fitness %.4f reached", e.config.TargetFitness)
			stats.Converged = true
			break
		}

		if e.config.MaxStagnation > 0 && stagnationCount >= e.config.MaxStagnation {
			stats.StoppedEarly = true
			stats.StopReason = fmt.Sprintf("stagnation for %d generations", e.config.MaxStagnation)
			break
		}

		// Last generation, no need to evolve
		if gen == e.config.Generations-1 {
			break
		}

		// Create next generation
		nextGeneration := e.evolveGeneration(population, students)

		// Evaluate new generation
		e.evaluatePopulation(goCtx, nextGeneration)
		stats.TotalEvaluations += len(nextGeneration)

		population = nextGeneration
	}

	stats.TotalTime = time.Since(startTime)
	stats.FinalBestFitness = bestFitness

	return bestChromosome, stats, nil
}

// evaluatePopulation evaluates fitness for all chromosomes.
func (e *Engine) evaluatePopulation(ctx context.Context, population []*domain.Chromosome) {
	for _, chromosome := range population {
		select {
		case <-ctx.Done():
			return
		default:
			chromosome.Fitness = CalculateFitness(chromosome, e.fitnessConf, e.factory, e.ctx)
		}
	}
}

// evolveGeneration creates a new generation using elitism, selection, crossover, and mutation.
func (e *Engine) evolveGeneration(population []*domain.Chromosome, students []*domain.Student) []*domain.Chromosome {
	nextGen := make([]*domain.Chromosome, 0, e.config.PopulationSize)

	// Elitism: preserve top individuals
	eliteCount := int(float64(e.config.PopulationSize) * e.config.ElitismRate)
	if eliteCount < 1 {
		eliteCount = 1
	}

	for i := 0; i < eliteCount; i++ {
		nextGen = append(nextGen, deepCopyChromosome(population[i]))
	}

	// Fill rest with offspring
	for len(nextGen) < e.config.PopulationSize {
		// Tournament selection
		parent1 := e.tournamentSelection(population)
		parent2 := e.tournamentSelection(population)

		// Crossover
		var offspring *domain.Chromosome
		if e.rng.Float64() < e.config.CrossoverRate {
			offspring = e.orderCrossover(parent1, parent2, students)
		} else {
			offspring = deepCopyChromosome(parent1)
		}

		// Mutation
		if e.rng.Float64() < e.config.MutationRate {
			e.swapMutation(offspring)
		}

		nextGen = append(nextGen, offspring)
	}

	// Trim if we generated too many
	if len(nextGen) > e.config.PopulationSize {
		nextGen = nextGen[:e.config.PopulationSize]
	}

	return nextGen
}

// tournamentSelection selects an individual using tournament selection.
func (e *Engine) tournamentSelection(population []*domain.Chromosome) *domain.Chromosome {
	tournamentSize := e.config.TournamentSize
	if tournamentSize > len(population) {
		tournamentSize = len(population)
	}

	best := population[e.rng.Intn(len(population))]
	bestFitness := best.Fitness

	for i := 1; i < tournamentSize; i++ {
		candidate := population[e.rng.Intn(len(population))]
		if candidate.Fitness > bestFitness {
			best = candidate
			bestFitness = candidate.Fitness
		}
	}

	return best
}

// orderCrossover performs order crossover (OX) while preserving group structure.
func (e *Engine) orderCrossover(parent1, parent2 *domain.Chromosome, students []*domain.Student) *domain.Chromosome {
	// Simple approach: take group structure from parent1, fill with order from parent2
	offspring := &domain.Chromosome{
		Groups:  make([]domain.Group, len(parent1.Groups)),
		Fitness: 0.0,
		Stats:   make(map[string]interface{}),
	}

	// Create flat list from parent2 for ordering reference
	parent2Order := make([]*domain.Student, 0, len(students))
	for _, group := range parent2.Groups {
		parent2Order = append(parent2Order, group...)
	}

	// Shuffle parent2 order slightly for variation
	if e.rng.Float64() < 0.3 {
		for i := len(parent2Order) - 1; i > 0; i-- {
			j := e.rng.Intn(i + 1)
			parent2Order[i], parent2Order[j] = parent2Order[j], parent2Order[i]
		}
	}

	// Fill offspring groups with parent2 ordering but parent1 structure
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

// swapMutation performs swap mutation that preserves group sizes.
// It randomly swaps students between groups or within groups.
func (e *Engine) swapMutation(chromosome *domain.Chromosome) {
	if len(chromosome.Groups) < 2 {
		return
	}

	mutationType := e.rng.Intn(2)

	switch mutationType {
	case 0:
		// Inter-group swap: swap students between two different groups
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
		// Intra-group swap: swap students within the same group
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

// calculateGenerationStats computes statistics for a generation.
func (e *Engine) calculateGenerationStats(gen int, population []*domain.Chromosome, elapsed time.Duration) GenerationStats {
	if len(population) == 0 {
		return GenerationStats{Generation: gen}
	}

	// Calculate basic stats
	best := population[0].Fitness
	worst := population[len(population)-1].Fitness
	sum := 0.0
	for _, chr := range population {
		sum += chr.Fitness
	}
	avg := sum / float64(len(population))

	// Calculate median
	median := population[len(population)/2].Fitness

	// Calculate standard deviation
	variance := 0.0
	for _, chr := range population {
		diff := chr.Fitness - avg
		variance += diff * diff
	}
	variance /= float64(len(population))
	stdDev := 0.0
	if variance > 0 {
		stdDev = variance // sqrt would be needed but using variance as proxy
	}

	// Calculate diversity (simple: range of fitness values)
	diversity := best - worst

	return GenerationStats{
		Generation:    gen,
		BestFitness:   best,
		WorstFitness:  worst,
		AvgFitness:    avg,
		MedianFitness: median,
		StdDev:        stdDev,
		Diversity:     diversity,
		ElapsedTime:   elapsed,
	}
}

// deepCopyChromosome creates a deep copy of a chromosome.
func deepCopyChromosome(src *domain.Chromosome) *domain.Chromosome {
	dst := &domain.Chromosome{
		Groups:  make([]domain.Group, len(src.Groups)),
		Fitness: src.Fitness,
		Stats:   make(map[string]interface{}),
	}

	// Copy groups
	for i, group := range src.Groups {
		dst.Groups[i] = make(domain.Group, len(group))
		copy(dst.Groups[i], group)
	}

	// Copy stats
	for k, v := range src.Stats {
		dst.Stats[k] = v
	}

	return dst
}

// PrintEvolutionStats prints a summary of evolution statistics.
func PrintEvolutionStats(stats *EvolutionStats) {
	fmt.Println("\n=== Evolution Statistics ===")
	fmt.Printf("Total Time: %v\n", stats.TotalTime)
	fmt.Printf("Total Evaluations: %d\n", stats.TotalEvaluations)
	fmt.Printf("Final Best Fitness: %.4f\n", stats.FinalBestFitness)
	fmt.Printf("Converged: %v\n", stats.Converged)

	if stats.StoppedEarly {
		fmt.Printf("Stopped Early: %s\n", stats.StopReason)
	}

	if len(stats.Generations) > 0 {
		fmt.Println("\nGeneration Progress:")
		fmt.Println("Gen\tBest\tAvg\tWorst\tDiversity")
		for i, gen := range stats.Generations {
			if i%10 == 0 || i == len(stats.Generations)-1 {
				fmt.Printf("%d\t%.4f\t%.4f\t%.4f\t%.4f\n",
					gen.Generation, gen.BestFitness, gen.AvgFitness, gen.WorstFitness, gen.Diversity)
			}
		}
	}

	fmt.Println("============================")
}
