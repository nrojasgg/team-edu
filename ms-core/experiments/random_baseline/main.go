package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"time"

	"ms-core/internal/core"
	"ms-core/internal/core/domain"
)

type ParetoPoint struct {
	F1 float64 `json:"f1"`
	F2 float64 `json:"f2"`
}

type RandomSearchResult struct {
	Run             int           `json:"run"`
	Seed            int64         `json:"seed"`
	ExecutionTimeMs int64         `json:"execution_time_ms"`
	HV              float64       `json:"hv"`
	SolutionsCount  int           `json:"solutions_count"`
	ParetoFront     []ParetoPoint `json:"pareto_front"`
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

func hypervolume2D(points []ParetoPoint, refX, refY float64) float64 {
	if len(points) == 0 {
		return 0.0
	}
	sortedPoints := make([]ParetoPoint, len(points))
	copy(sortedPoints, points)
	sort.Slice(sortedPoints, func(i, j int) bool {
		return sortedPoints[i].F1 < sortedPoints[j].F1
	})

	hv := 0.0
	prevX := 0.0
	prevY := refY
	for _, p := range sortedPoints {
		if p.F1 >= refX || p.F2 >= refY {
			continue
		}
		width := p.F1 - prevX
		hv += width * prevY
		if p.F2 < prevY {
			prevY = p.F2
		}
		prevX = p.F1
	}
	hv += (refX - prevX) * prevY
	return hv
}

func filterParetoFront(points []ParetoPoint) []ParetoPoint {
	if len(points) == 0 {
		return points
	}
	pareto := make([]ParetoPoint, 0)
	for i, p := range points {
		isDominated := false
		for j := range points {
			if i != j && dominates(points[j], p) {
				isDominated = true
				break
			}
		}
		if !isDominated {
			pareto = append(pareto, p)
		}
	}
	return pareto
}

func dominates(a, b ParetoPoint) bool {
	return (a.F1 <= b.F1 && a.F2 <= b.F2) && (a.F1 < b.F1 || a.F2 < b.F2)
}

func main() {
	const (
		numStudents    = 50
		teamSize       = 5
		populationSize = 100
		generations    = 50
		totalEvals     = populationSize * generations // 5000 evaluations
	)

	fmt.Printf("TeamEdu Random Search Baseline\n")
	fmt.Printf("================================\n")
	fmt.Printf("Students: %d, TeamSize: %d\n", numStudents, teamSize)
	fmt.Printf("Total Random Evaluations: %d\n\n", totalEvals)

	rng := rand.New(rand.NewSource(42))

	config := core.DefaultTeamFormationConfig()
	config.GeneticParams = &core.GeneticAlgorithmParams{
		PopulationSize: populationSize,
		Generations:    generations,
		MutationRate:   0.1,
		ElitismRate:    0.1,
	}

	engine := core.NewTeamFormationEngine(config)
	provider := &randomProvider{rng: rng}
	if err := engine.RegisterProvider(provider); err != nil {
		fmt.Printf("Error registering provider: %v\n", err)
		os.Exit(1)
	}

	candidates := generateSyntheticStudents(numStudents, rng)
	ctx := context.Background()
	affinityMatrix := engine.PrecomputeAffinityMatrix(ctx, candidates)

	// Run Random Search 5 times
	numRuns := 5
	hvValues := make([]float64, numRuns)
	solutionsCounts := make([]int, numRuns)
	executionTimes := make([]int, numRuns)

	for run := 0; run < numRuns; run++ {
		seed := rand.Int63()
		runRng := rand.New(rand.NewSource(seed))

		allSolutions := make([]ParetoPoint, 0, totalEvals)
		startTime := time.Now()

		for eval := 0; eval < totalEvals; eval++ {
			chromosome := runRng.Perm(len(candidates))
			fitness := engine.EvaluateChromosome(chromosome, candidates, affinityMatrix, teamSize)
			if fitness.IsFeasible {
				allSolutions = append(allSolutions, ParetoPoint{
					F1: fitness.PsyScore,
					F2: fitness.SocScore,
				})
			}
		}

		executionTime := time.Since(startTime)
		paretoFront := filterParetoFront(allSolutions)

		hv := hypervolume2D(paretoFront, 1.0, 1.0)
		hvValues[run] = hv
		solutionsCounts[run] = len(paretoFront)
		executionTimes[run] = int(executionTime.Milliseconds())

		fmt.Printf("Run %d: HV=%.6f, Solutions=%d, Time=%dms\n",
			run+1, hv, len(paretoFront), executionTime.Milliseconds())
	}

	// Compute statistics
	hvMean := mean(hvValues)
	hvStd := std(hvValues, hvMean)
	solutionsMean := meanFloat(intSliceToFloat(solutionsCounts))
	timeMean := meanFloat(intSliceToFloat(executionTimes))

	fmt.Printf("\nRandom Search Summary (%d runs):\n", numRuns)
	fmt.Printf("  HV Mean: %.6f\n", hvMean)
	fmt.Printf("  HV Std:  %.6f\n", hvStd)
	fmt.Printf("  Solutions Mean: %.1f\n", solutionsMean)
	fmt.Printf("  Time Mean: %.1fms\n", timeMean)

	// Save to JSON
	result := map[string]interface{}{
		"algorithm":          "Random Search",
		"num_runs":           numRuns,
		"total_evaluations":  totalEvals,
		"hv_values":          hvValues,
		"hv_mean":            hvMean,
		"hv_std":             hvStd,
		"solutions_counts":   solutionsCounts,
		"execution_times_ms": executionTimes,
	}

	outputFile := filepath.Join("..", "..", "experiments", "random_search_baseline.json")
	data, _ := json.MarshalIndent(result, "", "  ")
	os.WriteFile(outputFile, data, 0644)
	fmt.Printf("\nResults saved to: %s\n", outputFile)
}

func mean(arr []float64) float64 {
	if len(arr) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range arr {
		sum += v
	}
	return sum / float64(len(arr))
}

func std(arr []float64, mean float64) float64 {
	if len(arr) == 0 {
		return 0
	}
	sumSq := 0.0
	for _, v := range arr {
		sumSq += (v - mean) * (v - mean)
	}
	return math.Sqrt(sumSq / float64(len(arr)))
}

func intSliceToFloat(arr []int) []float64 {
	result := make([]float64, len(arr))
	for i, v := range arr {
		result[i] = float64(v)
	}
	return result
}

func meanFloat(arr []float64) float64 {
	if len(arr) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range arr {
		sum += v
	}
	return sum / float64(len(arr))
}
