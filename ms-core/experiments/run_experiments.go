package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"ms-core/internal/core"
	"ms-core/internal/core/domain"
	"ms-core/internal/core/genetic"
)

type ParetoPoint struct {
	F1 float64 `json:"f1"`
	F2 float64 `json:"f2"`
}

type ConvergencePoint struct {
	Generation int     `json:"generation"`
	HV         float64 `json:"hv"`
}

type RunResult struct {
	Run                  int                `json:"run"`
	Seed                 int64              `json:"seed"`
	ExecutionTimeMs      int64              `json:"execution_time_ms"`
	GenerationsCompleted int                `json:"generations_completed"`
	ParetoFront          []ParetoPoint      `json:"pareto_front"`
	Convergence          []ConvergencePoint `json:"convergence"`
}

type randomProvider struct {
	rng *rand.Rand
}

func (p *randomProvider) ProviderID() string   { return "random" }
func (p *randomProvider) ProviderName() string { return "Random Affinity Provider" }
func (p *randomProvider) SupportsContext(_ context.Context) bool {
	return true
}
func (p *randomProvider) CalculatePairwiseAffinity(_ context.Context, a, b string) (float64, error) {
	return 0.3 + p.rng.Float64()*0.7, nil
}
func (p *randomProvider) GetMetadata() core.ProviderMetadata {
	return core.ProviderMetadata{ID: "random", Name: "Random Affinity Provider", Category: "computed"}
}

var careerCodes = []string{"ISI", "IDI", "IIN", "IEL", "IIS", "IIC", "IEM", "IEN"}
var sexes = []string{"M", "F"}

func generateSyntheticStudents(n int, rng *rand.Rand) []core.TeamMember {
	students := make([]core.TeamMember, n)
	for i := 0; i < n; i++ {
		ocean := &domain.OCEANVector{
			Openness:          rng.Float64(),
			Conscientiousness: rng.Float64(),
			Extraversion:      rng.Float64(),
			Agreeableness:     rng.Float64(),
			Neuroticism:       rng.Float64(),
		}
		students[i] = core.TeamMember{
			ID:            fmt.Sprintf("student-%d", i),
			Psychometrics: ocean,
			Attributes: map[string]interface{}{
				"career_code": careerCodes[rng.Intn(len(careerCodes))],
				"sex":         sexes[rng.Intn(len(sexes))],
			},
		}
	}
	return students
}

func reportDataHeterogeneity(candidates []core.TeamMember, affinityMatrix map[string]map[string]float64) {
	fmt.Printf("\n=== DATA HETEROGENEITY REPORT ===\n\n")

	// OCEAN Statistics
	traits := []string{"Openness", "Conscientiousness", "Extraversion", "Agreeableness", "Neuroticism"}
	oceanData := make(map[string][]float64)
	for _, t := range traits {
		oceanData[t] = make([]float64, len(candidates))
	}

	for i, c := range candidates {
		if c.Psychometrics != nil {
			oceanData["Openness"][i] = c.Psychometrics.Openness
			oceanData["Conscientiousness"][i] = c.Psychometrics.Conscientiousness
			oceanData["Extraversion"][i] = c.Psychometrics.Extraversion
			oceanData["Agreeableness"][i] = c.Psychometrics.Agreeableness
			oceanData["Neuroticism"][i] = c.Psychometrics.Neuroticism
		}
	}

	fmt.Println("OCEAN Trait Statistics:")
	fmt.Printf("%-20s %-10s %-10s %-10s %-10s\n", "Trait", "Mean", "Std", "Min", "Max")
	fmt.Println(strings.Repeat("-", 60))
	for _, t := range traits {
		data := oceanData[t]
		mean := computeMean(data)
		std := computeStd(data, mean)
		min := computeMin(data)
		max := computeMax(data)
		fmt.Printf("%-20s %-10.4f %-10.4f %-10.4f %-10.4f\n", t, mean, std, min, max)
	}

	// Demographic Statistics
	careerCount := make(map[string]int)
	sexCount := make(map[string]int)
	for _, c := range candidates {
		if c.Attributes == nil {
			continue
		}
		if career, exists := c.Attributes["career_code"]; exists {
			careerCount[fmt.Sprintf("%v", career)]++
		}
		if sex, exists := c.Attributes["sex"]; exists {
			sexCount[fmt.Sprintf("%v", sex)]++
		}
	}

	fmt.Printf("\nDemographic Distribution (%d students):\n", len(candidates))
	fmt.Println("  Career Codes:")
	for career, count := range careerCount {
		fmt.Printf("    %s: %d (%.1f%%)\n", career, count, 100*float64(count)/float64(len(candidates)))
	}
	fmt.Println("  Sex:")
	for sex, count := range sexCount {
		fmt.Printf("    %s: %d (%.1f%%)\n", sex, count, 100*float64(count)/float64(len(candidates)))
	}

	// Affinity Matrix Statistics
	affinityValues := make([]float64, 0)
	lowCount := 0
	highCount := 0
	conflictCount := 0

	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			score := affinityMatrix[candidates[i].ID][candidates[j].ID]
			affinityValues = append(affinityValues, score)
			if score <= 0.3 {
				lowCount++
				if score <= 0.1 {
					conflictCount++ // severe conflict
				}
			} else if score >= 0.7 {
				highCount++
			}
		}
	}

	sort.Float64s(affinityValues)
	fmt.Printf("\nAffinity Matrix Statistics (%d pairs):\n", len(affinityValues))
	fmt.Printf("  Mean: %.4f\n", computeMean(affinityValues))
	fmt.Printf("  Std:  %.4f\n", computeStd(affinityValues, computeMean(affinityValues)))
	fmt.Printf("  Min:  %.4f\n", affinityValues[0])
	fmt.Printf("  Max:  %.4f\n", affinityValues[len(affinityValues)-1])
	fmt.Printf("  Median: %.4f\n", affinityValues[len(affinityValues)/2])

	fmt.Printf("\nAffinity Distribution:\n")
	fmt.Printf("  Low (<=0.3):    %d pairs (%.1f%%)\n", lowCount, 100*float64(lowCount)/float64(len(affinityValues)))
	fmt.Printf("  High (>=0.7):   %d pairs (%.1f%%)\n", highCount, 100*float64(highCount)/float64(len(affinityValues)))
	fmt.Printf("  Conflicts (<=0.1): %d pairs (%.1f%%)\n", conflictCount, 100*float64(conflictCount)/float64(len(affinityValues)))
	fmt.Println()
}

func computeMean(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

func computeStd(data []float64, mean float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sumSq := 0.0
	for _, v := range data {
		sumSq += (v - mean) * (v - mean)
	}
	return math.Sqrt(sumSq / float64(len(data)))
}

func computeMin(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	min := data[0]
	for _, v := range data {
		if v < min {
			min = v
		}
	}
	return min
}

func computeMax(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	max := data[0]
	for _, v := range data {
		if v > max {
			max = v
		}
	}
	return max
}

func main() {
	const (
		numRuns        = 30
		numStudents    = 50
		teamSize       = 5
		populationSize = 100
		generations    = 50
		mutationRate   = 0.1
		elitismRate    = 0.1
	)

	runsDir := filepath.Join("experiments", "runs")
	if err := os.MkdirAll(runsDir, 0755); err != nil {
		fmt.Printf("Error creating runs directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("TeamEdu NSGA-II Experiment Runner\n")
	fmt.Printf("================================\n")
	fmt.Printf("Runs: %d, Students: %d, TeamSize: %d\n", numRuns, numStudents, teamSize)
	fmt.Printf("Population: %d, Generations: %d, MutationRate: %.2f, ElitismRate: %.2f\n\n", populationSize, generations, mutationRate, elitismRate)

	for run := 1; run <= numRuns; run++ {
		seed := rand.Int63() // Truly random seed per run
		rng := rand.New(rand.NewSource(seed))

		config := core.DefaultTeamFormationConfig()
		config.GeneticParams = &core.GeneticAlgorithmParams{
			PopulationSize: populationSize,
			Generations:    generations,
			MutationRate:   mutationRate,
			ElitismRate:    elitismRate,
		}

		engine := core.NewTeamFormationEngine(config)
		provider := &randomProvider{rng: rng}
		if err := engine.RegisterProvider(provider); err != nil {
			fmt.Printf("Run %d: Error registering provider: %v\n", run, err)
			continue
		}

		candidates := generateSyntheticStudents(numStudents, rng)

		ctx := context.Background()
		affinityMatrix := engine.PrecomputeAffinityMatrix(ctx, candidates)

		population := make([]*genetic.Individual, populationSize)
		for i := range population {
			chromosome := rng.Perm(len(candidates))
			population[i] = &genetic.Individual{
				Chromosome: chromosome,
				Fitness:    engine.EvaluateChromosome(chromosome, candidates, affinityMatrix, teamSize),
			}
		}

		// Validation 1: Log first 5 genes of first individual for runs 1 and 2
		if run == 1 || run == 2 {
			first5 := population[0].Chromosome[:5]
			fmt.Printf("  [SEED VALIDATION] Run %d: First 5 genes of individual[0]: %v\n", run, first5)
		}

		// Validation 2: Report data heterogeneity once (run 1)
		if run == 1 {
			reportDataHeterogeneity(candidates, affinityMatrix)
		}

		evalFn := func(c []int) genetic.FitnessResult {
			return engine.EvaluateChromosome(c, candidates, affinityMatrix, teamSize)
		}

		convergence := make([]ConvergencePoint, 0, generations)
		onConvergence := func(gen int, hv float64) {
			convergence = append(convergence, ConvergencePoint{Generation: gen, HV: hv})
		}

		startTime := time.Now()
		finalFronts := engine.Nsga2Evolve(population, config.GeneticParams, rng, evalFn, onConvergence)
		executionTime := time.Since(startTime)

		generationsCompleted := generations
		if len(finalFronts) == 0 || len(finalFronts[0]) == 0 {
			fmt.Printf("Run %d (seed %d): WARNING - Empty Pareto front\n", run, seed)
			generationsCompleted = 0
		}

		paretoFront := make([]ParetoPoint, 0)
		if len(finalFronts) > 0 && len(finalFronts[0]) > 0 {
			for _, ind := range finalFronts[0] {
				paretoFront = append(paretoFront, ParetoPoint{
					F1: math.Round(ind.Fitness.PsyScore*10000) / 10000,
					F2: math.Round(ind.Fitness.SocScore*10000) / 10000,
				})
			}
		}

		result := RunResult{
			Run:                  run,
			Seed:                 seed,
			ExecutionTimeMs:      executionTime.Milliseconds(),
			GenerationsCompleted: generationsCompleted,
			ParetoFront:          paretoFront,
			Convergence:          convergence,
		}

		filename := filepath.Join(runsDir, fmt.Sprintf("run_%d.json", run))
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Printf("Run %d: Error marshaling JSON: %v\n", run, err)
			continue
		}

		if err := os.WriteFile(filename, data, 0644); err != nil {
			fmt.Printf("Run %d: Error writing file: %v\n", run, err)
			continue
		}

		fmt.Printf("Run %d (seed %d): %d solutions in Pareto front, time: %dms\n",
			run, seed, len(paretoFront), executionTime.Milliseconds())
	}

	fmt.Printf("\nAll experiments completed. Results saved to %s/\n", runsDir)

	fmt.Printf("\nRunning Python analysis...\n")
	pythonScript := filepath.Join("..", "experiments", "analyze_results.py")
	absScript, _ := filepath.Abs(pythonScript)
	cmd := exec.Command("py", absScript)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Python analysis warning: %v\n%s\n", err, string(output))
	} else {
		fmt.Printf("%s", output)
	}
}
