package genetic

import (
	"math"
	"testing"
)

func TestCheckDominance(t *testing.T) {
	tests := []struct {
		name string
		a    FitnessResult
		b    FitnessResult
		want bool
	}{
		{
			name: "A dominates B (better in both) - maximization: higher is better",
			a:    FitnessResult{PsyScore: 0.8, SocScore: 0.8, IsFeasible: true},
			b:    FitnessResult{PsyScore: 0.3, SocScore: 0.3, IsFeasible: true},
			want: true,
		},
		{
			name: "A dominates B (better in one, equal in other) - maximization",
			a:    FitnessResult{PsyScore: 0.8, SocScore: 0.5, IsFeasible: true},
			b:    FitnessResult{PsyScore: 0.5, SocScore: 0.5, IsFeasible: true},
			want: true,
		},
		{
			name: "A does not dominate B (worse in one) - maximization",
			a:    FitnessResult{PsyScore: 0.8, SocScore: 0.2, IsFeasible: true},
			b:    FitnessResult{PsyScore: 0.5, SocScore: 0.5, IsFeasible: true},
			want: false,
		},
		{
			name: "A does not dominate B (equal)",
			a:    FitnessResult{PsyScore: 0.5, SocScore: 0.5, IsFeasible: true},
			b:    FitnessResult{PsyScore: 0.5, SocScore: 0.5, IsFeasible: true},
			want: false,
		},
		{
			name: "A feasible, B infeasible (A dominates)",
			a:    FitnessResult{PsyScore: 0.1, SocScore: 0.1, IsFeasible: true},
			b:    FitnessResult{PsyScore: 0.9, SocScore: 0.9, IsFeasible: false},
			want: true,
		},
		{
			name: "A infeasible, B feasible (A does not dominate)",
			a:    FitnessResult{PsyScore: 0.9, SocScore: 0.9, IsFeasible: false},
			b:    FitnessResult{PsyScore: 0.1, SocScore: 0.1, IsFeasible: true},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckDominance(tt.a, tt.b); got != tt.want {
				t.Errorf("CheckDominance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCrowdedComparison(t *testing.T) {
	tests := []struct {
		name string
		a    FitnessResult
		b    FitnessResult
		want bool
	}{
		{
			name: "Better Rank wins",
			a:    FitnessResult{Rank: 1, Distance: 10},
			b:    FitnessResult{Rank: 2, Distance: 100},
			want: true,
		},
		{
			name: "Same Rank, larger Distance wins (diversity)",
			a:    FitnessResult{Rank: 1, Distance: 10},
			b:    FitnessResult{Rank: 1, Distance: 5},
			want: true,
		},
		{
			name: "Worse Rank loses",
			a:    FitnessResult{Rank: 2, Distance: 100},
			b:    FitnessResult{Rank: 1, Distance: 10},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CrowdedComparison(tt.a, tt.b); got != tt.want {
				t.Errorf("CrowdedComparison() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFastNonDominatedSort(t *testing.T) {
	// Setup population for MAXIMIZATION (higher is better)
	// Ind 1: (0.8, 0.8) - Front 1 (dominates Ind 2)
	// Ind 2: (0.2, 0.2) - Front 2 (dominated by 1)
	// Ind 3: (1.0, 0.5) - Front 1 (trade-off with 1)
	// Ind 4: (0.5, 1.0) - Front 1 (trade-off with 1)

	p1 := &Individual{Fitness: FitnessResult{PsyScore: 0.8, SocScore: 0.8, IsFeasible: true}, originalIndex: 0}
	p2 := &Individual{Fitness: FitnessResult{PsyScore: 0.2, SocScore: 0.2, IsFeasible: true}, originalIndex: 1}
	p3 := &Individual{Fitness: FitnessResult{PsyScore: 1.0, SocScore: 0.5, IsFeasible: true}, originalIndex: 2}
	p4 := &Individual{Fitness: FitnessResult{PsyScore: 0.5, SocScore: 1.0, IsFeasible: true}, originalIndex: 3}

	population := []*Individual{p1, p2, p3, p4}

	fronts := FastNonDominatedSort(population)

	if len(fronts) < 2 {
		t.Fatalf("Expected at least 2 fronts, got %d", len(fronts))
	}

	// Check Front 1 - should have 3 individuals (p1, p3, p4)
	if len(fronts[0]) != 3 {
		t.Errorf("Expected 3 individuals in Front 1, got %d", len(fronts[0]))
	}

	// Check Front 2 - should have p2 (dominated by p1)
	if len(fronts[1]) != 1 {
		t.Errorf("Expected 1 individual in Front 2, got %d", len(fronts[1]))
	}
	if fronts[1][0] != p2 {
		t.Errorf("Expected p2 in Front 2")
	}
}

func TestCalculateCrowdingDistance(t *testing.T) {
	// A front with 3 solutions in a line
	// A: (0,0)
	// B: (1,1)
	// C: (2,2)

	p1 := &Individual{Fitness: FitnessResult{PsyScore: 0, SocScore: 0}}
	p2 := &Individual{Fitness: FitnessResult{PsyScore: 1, SocScore: 1}}
	p3 := &Individual{Fitness: FitnessResult{PsyScore: 2, SocScore: 2}}

	front := []*Individual{p1, p2, p3}

	CalculateCrowdingDistance(front)

	// p1 and p3 are boundaries, should have Inf distance
	if !math.IsInf(p1.Fitness.Distance, 1) {
		t.Errorf("Expected p1 distance Inf, got %f", p1.Fitness.Distance)
	}
	if !math.IsInf(p3.Fitness.Distance, 1) {
		t.Errorf("Expected p3 distance Inf, got %f", p3.Fitness.Distance)
	}

	// p2 should have some finite distance
	// Range Psy: 2-0=2. p2 contrib: (2-0)/2 = 1? No, neighbors are p3(2) and p1(0). (2-0)/2 = 1.
	// Range Soc: 2-0=2. p2 contrib: (2-0)/2 = 1.
	// Total = 2.
	if math.IsInf(p2.Fitness.Distance, 0) || p2.Fitness.Distance <= 0 {
		t.Errorf("Expected p2 distance positive finite, got %f", p2.Fitness.Distance)
	}
}
