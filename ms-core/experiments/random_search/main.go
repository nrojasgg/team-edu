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
	Evaluations     int           `json:"evaluations"`
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

func generateSyntheticStudents(n int, rng *rand.Rand) []core.TeamMember {
	students := make([]core.TeamMember, n)
	for i := 0; i < n; i++ {
		students[i] = core.TeamMember{
			ID: fmt.Sprintf("student-%d", i),
			Psychometrics: &domain.OCEANVector{
				Openness:          rng.Float64(),
				Conscientiousness: rng.Float64(),
				Extraversion:      rng.Float64(),
				Agreeableness:     rng.Float64(),
				Neuroticism:       rng.Float64(),
			},
		}
	}
	return students
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

func dominates(a, b ParetoPoint) bool {
	return (a.F1 >= b.F1 && a.F2 >= b.F2) && (a.F1 > b.F1 || a.F2 > b.F2)
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

func main() {
	const (
		numStudents    = 50
		teamSize       = 5
		populationSize = 100
		generations    = 50
	)

	// Same number of fitness evaluations as NSGA-II
	totalEvaluations := populationSize * generations
	fmt.Printf("TeamEdu Random Search vs NSGA-II Comparison\n")
	fmt.Printf("============================================\n")
	fmt.Printf("Students: %d, TeamSize: %d\n", numStudents, teamSize)
	fmt.Printf("Population: %d, Generations: %d\n", populationSize, generations)
	fmt.Printf("Total Fitness Evaluations: %d\n\n", totalEvaluations)

	// Run Random Search 5 times for statistical significance
	numRuns := 5
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

	fmt.Printf("Running Random Search (%d runs)...\n", numRuns)
	randomHVValues := make([]float64, numRuns)
	randomSpacingValues := make([]float64, numRuns)

	for run := 1; run <= numRuns; run++ {
		seed := int64(run)
		runRng := rand.New(rand.NewSource(seed))

		allSolutions := make([]ParetoPoint, 0, totalEvaluations)
		startTime := time.Now()

		for eval := 0; eval < totalEvaluations; eval++ {
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
		randomHVValues[run-1] = hv

		// Compute spacing
		spacing := 0.0
		if len(paretoFront) > 1 {
			distances := make([]float64, len(paretoFront))
			for i := range paretoFront {
				minDist := math.MaxFloat64
				for j := range paretoFront {
					if i != j {
						d := math.Sqrt(math.Pow(paretoFront[i].F1-paretoFront[j].F1, 2) +
							math.Pow(paretoFront[i].F2-paretoFront[j].F2, 2))
						if d < minDist {
							minDist = d
						}
					}
				}
				distances[i] = minDist
			}
			dMean := computeMean(distances)
			spacing = computeStd(distances, dMean)
		}
		randomSpacingValues[run-1] = spacing

		fmt.Printf("  Random Search Run %d: HV=%.6f, Spacing=%.6f, Solutions=%d, Time=%dms\n",
			run, hv, spacing, len(paretoFront), executionTime.Milliseconds())
	}

	randomHVMean := computeMean(randomHVValues)
	randomHVStd := computeStd(randomHVValues, randomHVMean)

	fmt.Printf("\nRandom Search Summary:\n")
	fmt.Printf("  HV Mean: %.6f\n", randomHVMean)
	fmt.Printf("  HV Std:  %.6f\n", randomHVStd)

	// Compare with NSGA-II results (from previous run)
	fmt.Printf("\n=== COMPARISON WITH NSGA-II ===\n")
	fmt.Printf("Note: NSGA-II results from experiments/runs/\n")

	// Load NSGA-II results
	nsgaHVValues := make([]float64, 0)
	runsDir := filepath.Join("experiments", "runs")
	for i := 1; i <= 30; i++ {
		filename := filepath.Join(runsDir, fmt.Sprintf("run_%d.json", i))
		data, err := os.ReadFile(filename)
		if err != nil {
			continue
		}
		var result struct {
			ParetoFront []ParetoPoint `json:"pareto_front"`
		}
		if err := json.Unmarshal(data, &result); err != nil {
			continue
		}
		hv := hypervolume2D(result.ParetoFront, 1.0, 1.0)
		nsgaHVValues = append(nsgaHVValues, hv)
	}

	if len(nsgaHVValues) > 0 {
		nsgaHVMean := computeMean(nsgaHVValues)
		nsgaHVStd := computeStd(nsgaHVValues, nsgaHVMean)

		fmt.Printf("\nNSGA-II (30 runs):\n")
		fmt.Printf("  HV Mean: %.6f\n", nsgaHVMean)
		fmt.Printf("  HV Std:  %.6f\n", nsgaHVStd)

		fmt.Printf("\nRandom Search (5 runs):\n")
		fmt.Printf("  HV Mean: %.6f\n", randomHVMean)
		fmt.Printf("  HV Std:  %.6f\n", randomHVStd)

		improvement := ((nsgaHVMean - randomHVMean) / randomHVMean) * 100
		fmt.Printf("\n*** NSGA-II Improvement over Random Search: %.1f%% ***\n", improvement)

		if improvement > 50 {
			fmt.Printf("\n=> CONCLUSION: NSGA-II is significantly superior (>50%% improvement)\n")
			fmt.Printf("   The problem is NON-TRIVIAL and the algorithm is effective.\n")
		} else if improvement > 20 {
			fmt.Printf("\n=> CONCLUSION: NSGA-II shows moderate improvement (20-50%%)\n")
			fmt.Printf("   The problem has some structure that NSGA-II exploits.\n")
		} else {
			fmt.Printf("\n=> WARNING: NSGA-II shows marginal improvement (<20%%)\n")
			fmt.Printf("   The problem may be too easy or the search space is degenerate.\n")
		}
	} else {
		fmt.Printf("\nNo NSGA-II results found. Run experiments first.\n")
	}
}
