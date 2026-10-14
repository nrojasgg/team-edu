package genetic

import (
	"context"
	"testing"
	"time"

	"ms-core/internal/core/domain"
)

func TestNewEngine(t *testing.T) {
	config := DefaultEngineConfig()
	factory := NewCriterionFactory()
	factory.RegisterDefaults()
	fitnessConf := domain.Config{
		Weights: map[string]float64{"sex": 1.0},
	}
	ctx := domain.Context{}

	engine := NewEngine(config, factory, fitnessConf, ctx)

	if engine == nil {
		t.Fatal("expected engine to be created")
	}

	if engine.workers <= 0 {
		t.Errorf("expected positive worker count, got %d", engine.workers)
	}

	if engine.factory == nil {
		t.Error("expected factory to be set")
	}

	if engine.rng == nil {
		t.Error("expected rng to be initialized")
	}
}

func TestDefaultEngineConfig(t *testing.T) {
	config := DefaultEngineConfig()

	if config.PopulationSize <= 0 {
		t.Error("expected positive population size")
	}

	if config.Generations <= 0 {
		t.Error("expected positive generations")
	}

	if config.ElitismRate < 0 || config.ElitismRate > 1 {
		t.Error("expected elitism rate in [0, 1]")
	}

	if config.MutationRate < 0 || config.MutationRate > 1 {
		t.Error("expected mutation rate in [0, 1]")
	}

	if config.CrossoverRate < 0 || config.CrossoverRate > 1 {
		t.Error("expected crossover rate in [0, 1]")
	}

	if config.TournamentSize <= 0 {
		t.Error("expected positive tournament size")
	}

	if config.TargetGroupSize <= 0 {
		t.Error("expected positive target group size")
	}
}

func TestEngine_Run_BasicExecution(t *testing.T) {
	students := createTestStudents(12)

	config := DefaultEngineConfig()
	config.PopulationSize = 20
	config.Generations = 5
	config.WorkerCount = 2
	config.Seed = 42

	factory := NewCriterionFactory()
	factory.Register(NewCategoricalCriterion("sex", "sex"))

	fitnessConf := domain.Config{
		Weights: map[string]float64{"sex": 1.0},
	}

	ctx := domain.Context{}

	engine := NewEngine(config, factory, fitnessConf, ctx)

	goCtx := context.Background()
	best, stats, err := engine.Run(goCtx, students)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if best == nil {
		t.Fatal("expected best chromosome to be returned")
	}

	if stats == nil {
		t.Fatal("expected stats to be returned")
	}

	if len(stats.Generations) != config.Generations {
		t.Errorf("expected %d generation stats, got %d", config.Generations, len(stats.Generations))
	}

	if stats.FinalBestFitness <= 0 {
		t.Error("expected positive final fitness")
	}

	if stats.TotalEvaluations <= 0 {
		t.Error("expected positive evaluation count")
	}
}

func TestEngine_Run_EmptyStudents(t *testing.T) {
	config := DefaultEngineConfig()
	factory := NewCriterionFactory()
	fitnessConf := domain.Config{Weights: map[string]float64{"sex": 1.0}}
	ctx := domain.Context{}

	engine := NewEngine(config, factory, fitnessConf, ctx)

	_, _, err := engine.Run(context.Background(), []*domain.Student{})

	if err == nil {
		t.Error("expected error for empty students")
	}
}

func TestEngine_Run_ContextCancellation(t *testing.T) {
	students := createTestStudents(20)

	config := DefaultEngineConfig()
	config.Generations = 100 // Long run
	config.PopulationSize = 50

	factory := NewCriterionFactory()
	factory.Register(NewCategoricalCriterion("sex", "sex"))

	fitnessConf := domain.Config{
		Weights: map[string]float64{"sex": 1.0},
	}

	ctx := domain.Context{}
	engine := NewEngine(config, factory, fitnessConf, ctx)

	// Create cancellable context
	goCtx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, stats, err := engine.Run(goCtx, students)

	// Context cancellation may or may not occur depending on timing
	// Just verify that if cancelled, it's handled correctly
	if err != nil {
		if stats == nil {
			t.Fatal("expected stats even on cancellation")
		}
		if !stats.StoppedEarly {
			t.Error("expected StoppedEarly flag to be set on cancellation")
		}
	}
}

func TestEngine_Run_TargetFitnessStopCondition(t *testing.T) {
	students := createTestStudents(8)

	config := DefaultEngineConfig()
	config.Generations = 100
	config.PopulationSize = 30
	config.TargetFitness = 0.7 // Achievable target
	config.Seed = 42

	factory := NewCriterionFactory()
	factory.Register(NewCategoricalCriterion("sex", "sex"))

	fitnessConf := domain.Config{
		Weights: map[string]float64{"sex": 1.0},
	}

	ctx := domain.Context{}
	engine := NewEngine(config, factory, fitnessConf, ctx)

	_, stats, err := engine.Run(context.Background(), students)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// May or may not reach target, but should not crash
	if stats.Converged && stats.FinalBestFitness < config.TargetFitness {
		t.Errorf("marked as converged but fitness %.4f < target %.4f",
			stats.FinalBestFitness, config.TargetFitness)
	}
}

func TestEngine_Run_StagnationStopCondition(t *testing.T) {
	students := createTestStudents(8)

	config := DefaultEngineConfig()
	config.Generations = 100
	config.PopulationSize = 10
	config.MaxStagnation = 5
	config.MutationRate = 0.01 // Very low mutation to encourage stagnation
	config.Seed = 42

	factory := NewCriterionFactory()
	factory.Register(NewCategoricalCriterion("sex", "sex"))

	fitnessConf := domain.Config{
		Weights: map[string]float64{"sex": 1.0},
	}

	ctx := domain.Context{}
	engine := NewEngine(config, factory, fitnessConf, ctx)

	_, stats, err := engine.Run(context.Background(), students)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should stop early due to stagnation or complete normally
	if stats.StoppedEarly && len(stats.Generations) >= config.Generations {
		t.Error("marked as stopped early but ran all generations")
	}
}

func TestEngine_EvaluatePopulation(t *testing.T) {
	students := createTestStudents(12)

	config := DefaultEngineConfig()
	config.WorkerCount = 4
	config.TargetGroupSize = 4

	factory := NewCriterionFactory()
	factory.Register(NewCategoricalCriterion("sex", "sex"))

	fitnessConf := domain.Config{
		Weights: map[string]float64{"sex": 1.0},
	}

	ctx := domain.Context{}
	engine := NewEngine(config, factory, fitnessConf, ctx)

	// Create test population
	population, _ := InitPopulation(students, config.TargetGroupSize, 20, 42)

	// Evaluate
	goCtx := context.Background()
	engine.evaluatePopulation(goCtx, population)

	// Check all chromosomes have fitness
	for i, chr := range population {
		if chr.Fitness == 0.0 {
			t.Errorf("chromosome %d has zero fitness", i)
		}
	}
}

func TestEngine_EvaluatePopulation_Parallel(t *testing.T) {
	students := createTestStudents(20)

	config := DefaultEngineConfig()
	config.WorkerCount = 4
	config.TargetGroupSize = 4

	factory := NewCriterionFactory()
	factory.Register(NewCategoricalCriterion("sex", "sex"))

	fitnessConf := domain.Config{
		Weights: map[string]float64{"sex": 1.0},
	}

	ctx := domain.Context{}
	engine := NewEngine(config, factory, fitnessConf, ctx)

	population, _ := InitPopulation(students, config.TargetGroupSize, 100, 42)

	start := time.Now()
	engine.evaluatePopulation(context.Background(), population)
	parallelTime := time.Since(start)

	// Just verify it completes and all have fitness
	for i, chr := range population {
		if chr.Fitness == 0.0 {
			t.Errorf("chromosome %d has zero fitness after parallel evaluation", i)
		}
	}

	t.Logf("Parallel evaluation of %d chromosomes took %v", len(population), parallelTime)
}

func TestEngine_TournamentSelection(t *testing.T) {
	config := DefaultEngineConfig()
	config.TournamentSize = 3
	factory := NewCriterionFactory()
	fitnessConf := domain.Config{Weights: map[string]float64{"sex": 1.0}}
	ctx := domain.Context{}

	engine := NewEngine(config, factory, fitnessConf, ctx)

	// Create population with known fitness values
	population := make([]*domain.Chromosome, 10)
	for i := 0; i < 10; i++ {
		population[i] = &domain.Chromosome{
			Fitness: float64(i) / 10.0, // Fitness from 0.0 to 0.9
		}
	}

	// Run tournament multiple times
	selected := make(map[float64]int)
	for i := 0; i < 100; i++ {
		winner := engine.tournamentSelection(population)
		selected[winner.Fitness]++
	}

	// Higher fitness should be selected more often (not strict test, probabilistic)
	// Just verify it doesn't crash and selects something
	if len(selected) == 0 {
		t.Error("tournament selection selected nothing")
	}
}

func TestEngine_OrderCrossover(t *testing.T) {
	students := createTestStudents(12)

	config := DefaultEngineConfig()
	factory := NewCriterionFactory()
	fitnessConf := domain.Config{Weights: map[string]float64{"sex": 1.0}}
	ctx := domain.Context{}

	engine := NewEngine(config, factory, fitnessConf, ctx)

	// Create two parent chromosomes
	parent1 := &domain.Chromosome{
		Groups: []domain.Group{
			{students[0], students[1], students[2]},
			{students[3], students[4], students[5]},
			{students[6], students[7], students[8]},
			{students[9], students[10], students[11]},
		},
	}

	parent2 := &domain.Chromosome{
		Groups: []domain.Group{
			{students[11], students[10], students[9]},
			{students[8], students[7], students[6]},
			{students[5], students[4], students[3]},
			{students[2], students[1], students[0]},
		},
	}

	offspring := engine.orderCrossover(parent1, parent2, students)

	if offspring == nil {
		t.Fatal("expected offspring to be created")
	}

	if len(offspring.Groups) != len(parent1.Groups) {
		t.Errorf("expected %d groups, got %d", len(parent1.Groups), len(offspring.Groups))
	}

	// Verify all students present exactly once
	seen := make(map[string]bool)
	for _, group := range offspring.Groups {
		for _, student := range group {
			if seen[student.ID] {
				t.Errorf("student %s appears multiple times", student.ID)
			}
			seen[student.ID] = true
		}
	}

	if len(seen) != len(students) {
		t.Errorf("expected %d students, got %d", len(students), len(seen))
	}
}

func TestEngine_SwapMutation_InterGroup(t *testing.T) {
	students := createTestStudents(12)

	config := DefaultEngineConfig()
	config.Seed = 42
	factory := NewCriterionFactory()
	fitnessConf := domain.Config{Weights: map[string]float64{"sex": 1.0}}
	ctx := domain.Context{}

	engine := NewEngine(config, factory, fitnessConf, ctx)

	chromosome := &domain.Chromosome{
		Groups: []domain.Group{
			{students[0], students[1], students[2]},
			{students[3], students[4], students[5]},
			{students[6], students[7], students[8]},
		},
	}

	originalIDs := make([][]string, len(chromosome.Groups))
	for i, group := range chromosome.Groups {
		originalIDs[i] = make([]string, len(group))
		for j, student := range group {
			originalIDs[i][j] = student.ID
		}
	}

	// Apply mutation multiple times to test both types
	mutationOccurred := false
	for i := 0; i < 50; i++ {
		testChromosome := deepCopyChromosome(chromosome)
		engine.swapMutation(testChromosome)

		// Check if anything changed
		changed := false
		for gi, group := range testChromosome.Groups {
			for si, student := range group {
				if student.ID != originalIDs[gi][si] {
					changed = true
					break
				}
			}
		}

		if changed {
			mutationOccurred = true
		}

		// Verify no duplicates
		seen := make(map[string]bool)
		for _, group := range testChromosome.Groups {
			for _, student := range group {
				if seen[student.ID] {
					t.Errorf("mutation created duplicate: %s", student.ID)
				}
				seen[student.ID] = true
			}
		}

		// Verify group sizes preserved
		for gi, group := range testChromosome.Groups {
			if len(group) != len(chromosome.Groups[gi]) {
				t.Errorf("mutation changed group size: %d -> %d", len(chromosome.Groups[gi]), len(group))
			}
		}
	}

	if !mutationOccurred {
		t.Log("Warning: mutation didn't occur in 50 attempts (probabilistic, may happen)")
	}
}

func TestEngine_SwapMutation_SingleGroup(t *testing.T) {
	students := createTestStudents(4)

	config := DefaultEngineConfig()
	config.Seed = 42
	factory := NewCriterionFactory()
	fitnessConf := domain.Config{Weights: map[string]float64{"sex": 1.0}}
	ctx := domain.Context{}

	engine := NewEngine(config, factory, fitnessConf, ctx)

	chromosome := &domain.Chromosome{
		Groups: []domain.Group{
			{students[0], students[1], students[2], students[3]},
		},
	}

	// Mutation should not crash with single group
	engine.swapMutation(chromosome)

	// Verify all students still present
	if len(chromosome.Groups[0]) != 4 {
		t.Error("mutation changed group size")
	}
}

func TestEngine_CalculateGenerationStats(t *testing.T) {
	config := DefaultEngineConfig()
	factory := NewCriterionFactory()
	fitnessConf := domain.Config{Weights: map[string]float64{"sex": 1.0}}
	ctx := domain.Context{}

	engine := NewEngine(config, factory, fitnessConf, ctx)

	population := []*domain.Chromosome{
		{Fitness: 0.9},
		{Fitness: 0.8},
		{Fitness: 0.7},
		{Fitness: 0.6},
		{Fitness: 0.5},
	}

	stats := engine.calculateGenerationStats(5, population, 100*time.Millisecond)

	if stats.Generation != 5 {
		t.Errorf("expected generation 5, got %d", stats.Generation)
	}

	if stats.BestFitness != 0.9 {
		t.Errorf("expected best fitness 0.9, got %.2f", stats.BestFitness)
	}

	if stats.WorstFitness != 0.5 {
		t.Errorf("expected worst fitness 0.5, got %.2f", stats.WorstFitness)
	}

	expectedAvg := 0.7
	if stats.AvgFitness < expectedAvg-0.01 || stats.AvgFitness > expectedAvg+0.01 {
		t.Errorf("expected avg fitness ~%.2f, got %.2f", expectedAvg, stats.AvgFitness)
	}

	if stats.Diversity <= 0 {
		t.Error("expected positive diversity")
	}
}

func TestDeepCopyChromosome(t *testing.T) {
	students := createTestStudents(8)

	original := &domain.Chromosome{
		Groups: []domain.Group{
			{students[0], students[1]},
			{students[2], students[3]},
		},
		Fitness: 0.75,
		Stats: map[string]interface{}{
			"generation": 5,
			"rank":       1,
		},
	}

	copy := deepCopyChromosome(original)

	// Verify values copied
	if copy.Fitness != original.Fitness {
		t.Error("fitness not copied")
	}

	if len(copy.Groups) != len(original.Groups) {
		t.Error("groups not copied")
	}

	if len(copy.Stats) != len(original.Stats) {
		t.Error("stats not copied")
	}

	// Verify deep copy (modifying copy doesn't affect original)
	copy.Fitness = 0.5
	if original.Fitness == 0.5 {
		t.Error("modifying copy affected original fitness")
	}

	copy.Groups[0][0] = students[7]
	if original.Groups[0][0].ID == students[7].ID {
		t.Error("modifying copy affected original groups")
	}
}

func TestEngine_Elitism(t *testing.T) {
	students := createTestStudents(12)

	config := DefaultEngineConfig()
	config.PopulationSize = 10
	config.ElitismRate = 0.2 // Top 20% = 2 individuals
	config.Seed = 42

	factory := NewCriterionFactory()
	factory.Register(NewCategoricalCriterion("sex", "sex"))

	fitnessConf := domain.Config{
		Weights: map[string]float64{"sex": 1.0},
	}

	ctx := domain.Context{}
	engine := NewEngine(config, factory, fitnessConf, ctx)

	// Create population with known fitness sorted descending
	population := make([]*domain.Chromosome, 10)
	for i := 0; i < 10; i++ {
		chr, _ := InitPopulation(students, 4, 1, int64(i))
		population[i] = chr[0]
		// Fitness in descending order: 0.9, 0.8, ..., 0.0
		population[i].Fitness = float64(9-i) / 10.0
	}

	// Top fitness is first (already sorted)
	topFitness := population[0].Fitness

	// Evolve generation
	nextGen := engine.evolveGeneration(population, students)

	// Check that elitist individuals are present at start
	// ElitismRate of 0.2 means at least 2 individuals (20% of 10)
	eliteCount := int(float64(config.PopulationSize) * config.ElitismRate)
	if eliteCount < 1 {
		eliteCount = 1
	}

	// Verify at least the top individual is preserved
	if len(nextGen) < 1 {
		t.Fatal("next generation is empty")
	}

	if nextGen[0].Fitness != topFitness {
		t.Errorf("elitism failed: first individual has fitness %.4f, expected %.4f",
			nextGen[0].Fitness, topFitness)
	}
}

func TestPrintEvolutionStats(t *testing.T) {
	stats := &EvolutionStats{
		Generations: []GenerationStats{
			{Generation: 0, BestFitness: 0.5, AvgFitness: 0.3, WorstFitness: 0.1, Diversity: 0.4},
			{Generation: 10, BestFitness: 0.7, AvgFitness: 0.5, WorstFitness: 0.3, Diversity: 0.4},
			{Generation: 20, BestFitness: 0.85, AvgFitness: 0.7, WorstFitness: 0.5, Diversity: 0.35},
		},
		TotalTime:        500 * time.Millisecond,
		FinalBestFitness: 0.85,
		Converged:        true,
		StoppedEarly:     false,
		TotalEvaluations: 3000,
	}

	// Just verify it doesn't crash
	PrintEvolutionStats(stats)
}

func BenchmarkEngine_Run(b *testing.B) {
	students := createTestStudents(20)

	config := DefaultEngineConfig()
	config.PopulationSize = 30
	config.Generations = 20
	config.WorkerCount = 4

	factory := NewCriterionFactory()
	factory.Register(NewCategoricalCriterion("sex", "sex"))

	fitnessConf := domain.Config{
		Weights: map[string]float64{"sex": 1.0},
	}

	ctx := domain.Context{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine := NewEngine(config, factory, fitnessConf, ctx)
		_, _, _ = engine.Run(context.Background(), students)
	}
}

func BenchmarkEngine_EvaluatePopulation(b *testing.B) {
	students := createTestStudents(20)

	config := DefaultEngineConfig()
	config.WorkerCount = 4

	factory := NewCriterionFactory()
	factory.Register(NewCategoricalCriterion("sex", "sex"))

	fitnessConf := domain.Config{
		Weights: map[string]float64{"sex": 1.0},
	}

	ctx := domain.Context{}
	engine := NewEngine(config, factory, fitnessConf, ctx)

	population, _ := InitPopulation(students, 4, 100, 42)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.evaluatePopulation(context.Background(), population)
	}
}
