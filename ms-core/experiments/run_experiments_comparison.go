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

type ComparisonParetoPoint struct {
	F1 float64 `json:"f1"`
	F2 float64 `json:"f2"`
}

type ComparisonConvergencePoint struct {
	Generation int     `json:"generation"`
	HV         float64 `json:"hv"`
}

type ComparisonRunResult struct {
	Run                  int                          `json:"run"`
	Seed                 int64                        `json:"seed"`
	ExecutionTimeMs      int64                        `json:"execution_time_ms"`
	GenerationsCompleted int                          `json:"generations_completed"`
	ParetoFront          []ComparisonParetoPoint      `json:"pareto_front"`
	Convergence          []ComparisonConvergencePoint `json:"convergence"`
}

type ComparisonAlgorithmResult struct {
	Name           string                `json:"name"`
	Runs           []ComparisonRunResult `json:"runs"`
	HvValues       []float64             `json:"hv_values"`
	SpacingValues  []float64             `json:"spacing_values"`
	AvgHv          float64               `json:"avg_hv"`
	AvgSpacing     float64               `json:"avg_spacing"`
	AvgExecutionMs int64                 `json:"avg_execution_ms"`
}

type comparisonRandomProvider struct {
	rng *rand.Rand
}

func (p *comparisonRandomProvider) ProviderID() string { return "comparison_random" }
func (p *comparisonRandomProvider) ProviderName() string {
	return "Comparison Random Affinity Provider"
}
func (p *comparisonRandomProvider) SupportsContext(_ context.Context) bool {
	return true
}
func (p *comparisonRandomProvider) CalculatePairwiseAffinity(_ context.Context, a, b string) (float64, error) {
	return 0.3 + p.rng.Float64()*0.7, nil
}
func (p *comparisonRandomProvider) GetMetadata() core.ProviderMetadata {
	return core.ProviderMetadata{ID: "comparison_random", Name: "Comparison Random Affinity Provider", Category: "computed"}
}

var comparisonCareerCodes = []string{"ISI", "IDI", "IIN", "IEL", "IIS", "IIC", "IEM", "IEN"}
var comparisonSexes = []string{"M", "F"}

func comparisonGenerateSyntheticStudents(n int, rng *rand.Rand) []core.TeamMember {
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
				"career_code": comparisonCareerCodes[rng.Intn(len(comparisonCareerCodes))],
				"sex":         comparisonSexes[rng.Intn(len(comparisonSexes))],
			},
		}
	}
	return students
}

func comparisonReportDataHeterogeneity(candidates []core.TeamMember, affinityMatrix map[string]map[string]float64) {
	fmt.Printf("\n=== DATA HETEROGENEITY REPORT ===\n\n")

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
		mean := comparisonComputeMean(data)
		std := comparisonComputeStd(data, mean)
		min := comparisonComputeMin(data)
		max := comparisonComputeMax(data)
		fmt.Printf("%-20s %-10.4f %-10.4f %-10.4f %-10.4f\n", t, mean, std, min, max)
	}

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
					conflictCount++
				}
			} else if score >= 0.7 {
				highCount++
			}
		}
	}

	sort.Float64s(affinityValues)
	fmt.Printf("\nAffinity Matrix Statistics (%d pairs):\n", len(affinityValues))
	fmt.Printf("  Mean: %.4f\n", comparisonComputeMean(affinityValues))
	fmt.Printf("  Std:  %.4f\n", comparisonComputeStd(affinityValues, comparisonComputeMean(affinityValues)))
	fmt.Printf("  Min:  %.4f\n", affinityValues[0])
	fmt.Printf("  Max:  %.4f\n", affinityValues[len(affinityValues)-1])
	fmt.Printf("  Median: %.4f\n", affinityValues[len(affinityValues)/2])

	fmt.Printf("\nAffinity Distribution:\n")
	fmt.Printf("  Low (<=0.3):    %d pairs (%.1f%%)\n", lowCount, 100*float64(lowCount)/float64(len(affinityValues)))
	fmt.Printf("  High (>=0.7):   %d pairs (%.1f%%)\n", highCount, 100*float64(highCount)/float64(len(affinityValues)))
	fmt.Printf("  Conflicts (<=0.1): %d pairs (%.1f%%)\n", conflictCount, 100*float64(conflictCount)/float64(len(affinityValues)))
	fmt.Println()
}

func comparisonComputeMean(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

func comparisonComputeStd(data []float64, mean float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sumSq := 0.0
	for _, v := range data {
		sumSq += (v - mean) * (v - mean)
	}
	return math.Sqrt(sumSq / float64(len(data)))
}

func comparisonComputeMin(data []float64) float64 {
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

func comparisonComputeMax(data []float64) float64 {
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

func comparisonComputeSpacing(front []ComparisonParetoPoint) float64 {
	if len(front) < 2 {
		return 0
	}
	distances := make([]float64, 0)
	for i := 0; i < len(front); i++ {
		minDist := math.MaxFloat64
		for j := 0; j < len(front); j++ {
			if i == j {
				continue
			}
			dx := front[i].F1 - front[j].F1
			dy := front[i].F2 - front[j].F2
			d := math.Sqrt(dx*dx + dy*dy)
			if d < minDist {
				minDist = d
			}
		}
		distances = append(distances, minDist)
	}
	return comparisonComputeMean(distances)
}

func runNsga2(engine *core.TeamFormationEngine, population []*genetic.Individual, params *core.GeneticAlgorithmParams, rng *rand.Rand, evalFn func([]int) genetic.FitnessResult) (ComparisonAlgorithmResult, []ComparisonParetoPoint) {
	convergence := make([]ComparisonConvergencePoint, 0, params.Generations)
	onConvergence := func(gen int, hv float64) {
		convergence = append(convergence, ComparisonConvergencePoint{Generation: gen, HV: hv})
	}

	startTime := time.Now()
	finalFronts := engine.Nsga2Evolve(population, params, rng, evalFn, onConvergence)
	_ = time.Since(startTime)

	paretoFront := make([]ComparisonParetoPoint, 0)
	if len(finalFronts) > 0 && len(finalFronts[0]) > 0 {
		for _, ind := range finalFronts[0] {
			paretoFront = append(paretoFront, ComparisonParetoPoint{
				F1: math.Round(ind.Fitness.PsyScore*10000) / 10000,
				F2: math.Round(ind.Fitness.SocScore*10000) / 10000,
			})
		}
	}

	hv := 0.0
	if len(finalFronts) > 0 && len(finalFronts[0]) > 0 {
		hv = genetic.Hypervolume2D(finalFronts[0], 1.0, 1.0)
	}

	spacing := comparisonComputeSpacing(paretoFront)

	result := ComparisonAlgorithmResult{
		Name:          "NSGA-II",
		HvValues:      []float64{hv},
		SpacingValues: []float64{spacing},
	}

	return result, paretoFront
}

func runNsga2NoCrowding(engine *core.TeamFormationEngine, population []*genetic.Individual, params *core.GeneticAlgorithmParams, rng *rand.Rand, evalFn func([]int) genetic.FitnessResult) (ComparisonAlgorithmResult, []ComparisonParetoPoint) {
	convergence := make([]ComparisonConvergencePoint, 0, params.Generations)
	onConvergence := func(gen int, hv float64) {
		convergence = append(convergence, ComparisonConvergencePoint{Generation: gen, HV: hv})
	}

	startTime := time.Now()
	finalFronts := engine.Nsga2EvolveNoCrowding(population, params, rng, evalFn, onConvergence)
	_ = time.Since(startTime)

	paretoFront := make([]ComparisonParetoPoint, 0)
	if len(finalFronts) > 0 && len(finalFronts[0]) > 0 {
		for _, ind := range finalFronts[0] {
			paretoFront = append(paretoFront, ComparisonParetoPoint{
				F1: math.Round(ind.Fitness.PsyScore*10000) / 10000,
				F2: math.Round(ind.Fitness.SocScore*10000) / 10000,
			})
		}
	}

	hv := 0.0
	if len(finalFronts) > 0 && len(finalFronts[0]) > 0 {
		hv = genetic.Hypervolume2D(finalFronts[0], 1.0, 1.0)
	}

	spacing := comparisonComputeSpacing(paretoFront)

	result := ComparisonAlgorithmResult{
		Name:          "NSGA-II NoCrowding",
		HvValues:      []float64{hv},
		SpacingValues: []float64{spacing},
	}

	return result, paretoFront
}

func runWeightedSumGA(engine *core.TeamFormationEngine, population []*genetic.Individual, params *core.GeneticAlgorithmParams, rng *rand.Rand, evalFn func([]int) genetic.FitnessResult) (ComparisonAlgorithmResult, []ComparisonParetoPoint) {
	convergence := make([]ComparisonConvergencePoint, 0, params.Generations)
	onConvergence := func(gen int, hv float64) {
		convergence = append(convergence, ComparisonConvergencePoint{Generation: gen, HV: hv})
	}

	startTime := time.Now()
	finalPopulation := engine.WeightedSumEvolve(population, params, rng, evalFn, onConvergence)
	_ = time.Since(startTime)

	paretoFront := make([]ComparisonParetoPoint, 0)
	for _, ind := range finalPopulation {
		paretoFront = append(paretoFront, ComparisonParetoPoint{
			F1: math.Round(ind.Fitness.PsyScore*10000) / 10000,
			F2: math.Round(ind.Fitness.SocScore*10000) / 10000,
		})
	}

	hv := 0.0
	if len(finalPopulation) > 0 {
		fronts := genetic.FastNonDominatedSort(finalPopulation)
		if len(fronts) > 0 && len(fronts[0]) > 0 {
			hv = genetic.Hypervolume2D(fronts[0], 1.0, 1.0)
		}
	}

	spacing := comparisonComputeSpacing(paretoFront)

	result := ComparisonAlgorithmResult{
		Name:          "Weighted Sum GA",
		HvValues:      []float64{hv},
		SpacingValues: []float64{spacing},
	}

	return result, paretoFront
}

func runRandomSearch(engine *core.TeamFormationEngine, numCandidates int, teamSize int, numEvaluations int, rng *rand.Rand, evalFn func([]int) genetic.FitnessResult) (ComparisonAlgorithmResult, []ComparisonParetoPoint) {
	allIndividuals := make([]*genetic.Individual, 0, numEvaluations)

	for i := 0; i < numEvaluations; i++ {
		chromosome := rng.Perm(numCandidates)
		fitness := evalFn(chromosome)
		allIndividuals = append(allIndividuals, &genetic.Individual{
			Chromosome: chromosome,
			Fitness:    fitness,
		})
	}

	paretoFront := make([]ComparisonParetoPoint, 0)
	if len(allIndividuals) > 0 {
		fronts := genetic.FastNonDominatedSort(allIndividuals)
		if len(fronts) > 0 && len(fronts[0]) > 0 {
			for _, ind := range fronts[0] {
				paretoFront = append(paretoFront, ComparisonParetoPoint{
					F1: math.Round(ind.Fitness.PsyScore*10000) / 10000,
					F2: math.Round(ind.Fitness.SocScore*10000) / 10000,
				})
			}
			hv := genetic.Hypervolume2D(fronts[0], 1.0, 1.0)
			spacing := comparisonComputeSpacing(paretoFront)

			result := ComparisonAlgorithmResult{
				Name:          "Random Search",
				HvValues:      []float64{hv},
				SpacingValues: []float64{spacing},
			}
			return result, paretoFront
		}
	}

	result := ComparisonAlgorithmResult{
		Name:          "Random Search",
		HvValues:      []float64{0.0},
		SpacingValues: []float64{0.0},
	}
	return result, paretoFront
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

	resultsDir := filepath.Join("experiments", "comparison")
	if err := os.MkdirAll(resultsDir, 0755); err != nil {
		fmt.Printf("Error creating results directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("TeamEdu Algorithm Comparison Experiment Runner\n")
	fmt.Printf("===============================================\n")
	fmt.Printf("Runs: %d, Students: %d, TeamSize: %d\n", numRuns, numStudents, teamSize)
	fmt.Printf("Population: %d, Generations: %d, MutationRate: %.2f, ElitismRate: %.2f\n\n", populationSize, generations, mutationRate, elitismRate)

	nsga2Results := ComparisonAlgorithmResult{Name: "NSGA-II", HvValues: make([]float64, 0), SpacingValues: make([]float64, 0)}
	nsga2NoCrowdingResults := ComparisonAlgorithmResult{Name: "NSGA-II NoCrowding", HvValues: make([]float64, 0), SpacingValues: make([]float64, 0)}
	weightedSumResults := ComparisonAlgorithmResult{Name: "Weighted Sum GA", HvValues: make([]float64, 0), SpacingValues: make([]float64, 0)}
	randomResults := ComparisonAlgorithmResult{Name: "Random Search", HvValues: make([]float64, 0), SpacingValues: make([]float64, 0)}

	for run := 1; run <= numRuns; run++ {
		seed := rand.Int63()
		rng := rand.New(rand.NewSource(seed))

		config := core.DefaultTeamFormationConfig()
		config.GeneticParams = &core.GeneticAlgorithmParams{
			PopulationSize: populationSize,
			Generations:    generations,
			MutationRate:   mutationRate,
			ElitismRate:    elitismRate,
		}

		engine := core.NewTeamFormationEngine(config)
		provider := &comparisonRandomProvider{rng: rng}
		if err := engine.RegisterProvider(provider); err != nil {
			fmt.Printf("Run %d: Error registering provider: %v\n", run, err)
			continue
		}

		candidates := comparisonGenerateSyntheticStudents(numStudents, rng)

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

		if run == 1 {
			comparisonReportDataHeterogeneity(candidates, affinityMatrix)
		}

		evalFn := func(c []int) genetic.FitnessResult {
			return engine.EvaluateChromosome(c, candidates, affinityMatrix, teamSize)
		}

		fmt.Printf("\n--- Run %d/%d (seed %d) ---\n", run, numRuns, seed)

		fmt.Printf("  Running NSGA-II...\n")
		nsga2Result, nsga2Front := runNsga2(engine, population, config.GeneticParams, rng, evalFn)
		nsga2Hv := nsga2Result.HvValues[0]
		nsga2Spacing := nsga2Result.SpacingValues[0]
		nsga2Results.HvValues = append(nsga2Results.HvValues, nsga2Hv)
		nsga2Results.SpacingValues = append(nsga2Results.SpacingValues, nsga2Spacing)

		newPopulation := make([]*genetic.Individual, populationSize)
		for i := range newPopulation {
			chromosome := rng.Perm(len(candidates))
			newPopulation[i] = &genetic.Individual{
				Chromosome: chromosome,
				Fitness:    engine.EvaluateChromosome(chromosome, candidates, affinityMatrix, teamSize),
			}
		}

		fmt.Printf("  Running NSGA-II NoCrowding...\n")
		nsga2ncResult, nsga2ncFront := runNsga2NoCrowding(engine, newPopulation, config.GeneticParams, rng, evalFn)
		nsga2ncHv := nsga2ncResult.HvValues[0]
		nsga2ncSpacing := nsga2ncResult.SpacingValues[0]
		nsga2NoCrowdingResults.HvValues = append(nsga2NoCrowdingResults.HvValues, nsga2ncHv)
		nsga2NoCrowdingResults.SpacingValues = append(nsga2NoCrowdingResults.SpacingValues, nsga2ncSpacing)

		newPopulation = make([]*genetic.Individual, populationSize)
		for i := range newPopulation {
			chromosome := rng.Perm(len(candidates))
			newPopulation[i] = &genetic.Individual{
				Chromosome: chromosome,
				Fitness:    engine.EvaluateChromosome(chromosome, candidates, affinityMatrix, teamSize),
			}
		}

		fmt.Printf("  Running Weighted Sum GA...\n")
		wsResult, wsFront := runWeightedSumGA(engine, newPopulation, config.GeneticParams, rng, evalFn)
		wsHv := wsResult.HvValues[0]
		wsSpacing := wsResult.SpacingValues[0]
		weightedSumResults.HvValues = append(weightedSumResults.HvValues, wsHv)
		weightedSumResults.SpacingValues = append(weightedSumResults.SpacingValues, wsSpacing)

		fmt.Printf("  Running Random Search (5000 evaluations)...\n")
		randomResult, randomFront := runRandomSearch(engine, len(candidates), teamSize, populationSize*generations, rng, evalFn)
		randomHv := randomResult.HvValues[0]
		randomSpacing := randomResult.SpacingValues[0]
		randomResults.HvValues = append(randomResults.HvValues, randomHv)
		randomResults.SpacingValues = append(randomResults.SpacingValues, randomSpacing)

		fmt.Printf("  NSGA-II: HV=%.4f, Spacing=%.4f, FrontSize=%d\n", nsga2Hv, nsga2Spacing, len(nsga2Front))
		fmt.Printf("  NSGA-II NoCrowding: HV=%.4f, Spacing=%.4f, FrontSize=%d\n", nsga2ncHv, nsga2ncSpacing, len(nsga2ncFront))
		fmt.Printf("  Weighted Sum GA: HV=%.4f, Spacing=%.4f, FrontSize=%d\n", wsHv, wsSpacing, len(wsFront))
		fmt.Printf("  Random Search: HV=%.4f, Spacing=%.4f, FrontSize=%d\n", randomHv, randomSpacing, len(randomFront))
	}

	nsga2Results.AvgHv = comparisonComputeMean(nsga2Results.HvValues)
	nsga2Results.AvgSpacing = comparisonComputeMean(nsga2Results.SpacingValues)
	nsga2NoCrowdingResults.AvgHv = comparisonComputeMean(nsga2NoCrowdingResults.HvValues)
	nsga2NoCrowdingResults.AvgSpacing = comparisonComputeMean(nsga2NoCrowdingResults.SpacingValues)
	weightedSumResults.AvgHv = comparisonComputeMean(weightedSumResults.HvValues)
	weightedSumResults.AvgSpacing = comparisonComputeMean(weightedSumResults.SpacingValues)
	randomResults.AvgHv = comparisonComputeMean(randomResults.HvValues)
	randomResults.AvgSpacing = comparisonComputeMean(randomResults.SpacingValues)

	allResults := []ComparisonAlgorithmResult{nsga2Results, nsga2NoCrowdingResults, weightedSumResults, randomResults}

	fmt.Printf("\n=== SUMMARY ===\n")
	fmt.Printf("%-20s %-12s %-12s\n", "Algorithm", "Avg HV", "Avg Spacing")
	fmt.Println(strings.Repeat("-", 50))
	for _, r := range allResults {
		fmt.Printf("%-20s %-12.4f %-12.4f\n", r.Name, r.AvgHv, r.AvgSpacing)
	}

	comparisonFile := filepath.Join(resultsDir, "algorithm_comparison.json")
	data, err := json.MarshalIndent(allResults, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(comparisonFile, data, 0644); err != nil {
		fmt.Printf("Error writing comparison file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nAll experiments completed. Results saved to %s\n", comparisonFile)

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
