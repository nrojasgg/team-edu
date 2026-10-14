package genetic

import (
	"fmt"
	"math/rand"
	"testing"

	"ms-core/internal/core/domain"
)

func TestCalculateGroupDistribution_PerfectDivision(t *testing.T) {
	tests := []struct {
		name              string
		numStudents       int
		targetGroupSize   int
		expectedSizes     []int
		expectedNumGroups int
	}{
		{
			name:              "12 students, size 4",
			numStudents:       12,
			targetGroupSize:   4,
			expectedSizes:     []int{4, 4, 4},
			expectedNumGroups: 3,
		},
		{
			name:              "15 students, size 5",
			numStudents:       15,
			targetGroupSize:   5,
			expectedSizes:     []int{5, 5, 5},
			expectedNumGroups: 3,
		},
		{
			name:              "8 students, size 4",
			numStudents:       8,
			targetGroupSize:   4,
			expectedSizes:     []int{4, 4},
			expectedNumGroups: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distribution, err := CalculateGroupDistribution(tt.numStudents, tt.targetGroupSize)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(distribution) != tt.expectedNumGroups {
				t.Errorf("expected %d groups, got %d", tt.expectedNumGroups, len(distribution))
			}

			// Verify total
			total := 0
			for _, size := range distribution {
				total += size
			}
			if total != tt.numStudents {
				t.Errorf("expected total %d, got %d", tt.numStudents, total)
			}

			// Check distribution matches
			for i, expectedSize := range tt.expectedSizes {
				if i >= len(distribution) {
					t.Errorf("missing expected group at index %d", i)
					continue
				}
				if distribution[i] != expectedSize {
					t.Errorf("group %d: expected size %d, got %d", i, expectedSize, distribution[i])
				}
			}
		})
	}
}

func TestCalculateGroupDistribution_WithRemainder(t *testing.T) {
	tests := []struct {
		name            string
		numStudents     int
		targetGroupSize int
		validateFunc    func([]int) bool
	}{
		{
			name:            "26 students, size 4",
			numStudents:     26,
			targetGroupSize: 4,
			validateFunc: func(dist []int) bool {
				// Should be [4,4,4,4,5,5] or similar balanced distribution
				total := 0
				for _, s := range dist {
					total += s
				}
				if total != 26 {
					return false
				}
				// Check variance is minimal
				min, max := dist[0], dist[0]
				for _, s := range dist {
					if s < min {
						min = s
					}
					if s > max {
						max = s
					}
				}
				return max-min <= 1 // Balanced if difference is at most 1
			},
		},
		{
			name:            "10 students, size 3",
			numStudents:     10,
			targetGroupSize: 3,
			validateFunc: func(dist []int) bool {
				// Should be [3,3,4] or similar
				total := 0
				for _, s := range dist {
					total += s
				}
				if total != 10 {
					return false
				}
				min, max := dist[0], dist[0]
				for _, s := range dist {
					if s < min {
						min = s
					}
					if s > max {
						max = s
					}
				}
				return max-min <= 1
			},
		},
		{
			name:            "7 students, size 3",
			numStudents:     7,
			targetGroupSize: 3,
			validateFunc: func(dist []int) bool {
				// Should avoid singletons: prefer [3,4] over [3,3,1]
				total := 0
				for _, s := range dist {
					total += s
					if s < 2 {
						return false // No singletons
					}
				}
				return total == 7
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distribution, err := CalculateGroupDistribution(tt.numStudents, tt.targetGroupSize)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !tt.validateFunc(distribution) {
				t.Errorf("distribution %v failed validation", distribution)
			}

			// Always verify total
			total := 0
			for _, size := range distribution {
				total += size
			}
			if total != tt.numStudents {
				t.Errorf("expected total %d, got %d (distribution: %v)", tt.numStudents, total, distribution)
			}
		})
	}
}

func TestCalculateGroupDistribution_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		numStudents   int
		targetSize    int
		expectError   bool
		errorContains string
	}{
		{
			name:          "zero students",
			numStudents:   0,
			targetSize:    4,
			expectError:   true,
			errorContains: "must be positive",
		},
		{
			name:          "negative students",
			numStudents:   -5,
			targetSize:    4,
			expectError:   true,
			errorContains: "must be positive",
		},
		{
			name:          "zero target size",
			numStudents:   10,
			targetSize:    0,
			expectError:   true,
			errorContains: "must be positive",
		},
		{
			name:          "negative target size",
			numStudents:   10,
			targetSize:    -3,
			expectError:   true,
			errorContains: "must be positive",
		},
		{
			name:        "target larger than students",
			numStudents: 3,
			targetSize:  5,
			expectError: false, // Should return [3]
		},
		{
			name:        "single student",
			numStudents: 1,
			targetSize:  4,
			expectError: false, // Should return [1]
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distribution, err := CalculateGroupDistribution(tt.numStudents, tt.targetSize)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errorContains)
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				// Verify total for valid cases
				if distribution != nil {
					total := 0
					for _, size := range distribution {
						total += size
					}
					if total != tt.numStudents {
						t.Errorf("expected total %d, got %d", tt.numStudents, total)
					}
				}
			}
		})
	}
}

func TestInitPopulation_BasicFunctionality(t *testing.T) {
	students := createTestStudents(12)
	targetSize := 4
	populationSize := 10
	seed := int64(42)

	population, err := InitPopulation(students, targetSize, populationSize, seed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(population) != populationSize {
		t.Errorf("expected population size %d, got %d", populationSize, len(population))
	}

	// Verify each chromosome
	for i, chromosome := range population {
		// Should have 3 groups of 4
		if len(chromosome.Groups) != 3 {
			t.Errorf("chromosome %d: expected 3 groups, got %d", i, len(chromosome.Groups))
		}

		// Verify no duplicates and all students present
		if err := ValidateChromosome(chromosome, students); err != nil {
			t.Errorf("chromosome %d failed validation: %v", i, err)
		}

		// Check group sizes
		for j, group := range chromosome.Groups {
			if len(group) != 4 {
				t.Errorf("chromosome %d, group %d: expected size 4, got %d", i, j, len(group))
			}
		}
	}
}

func TestInitPopulation_Deterministic(t *testing.T) {
	students := createTestStudents(20)
	targetSize := 4
	populationSize := 5
	seed := int64(123)

	// Generate population twice with same seed
	pop1, err1 := InitPopulation(students, targetSize, populationSize, seed)
	if err1 != nil {
		t.Fatalf("first generation error: %v", err1)
	}

	pop2, err2 := InitPopulation(students, targetSize, populationSize, seed)
	if err2 != nil {
		t.Fatalf("second generation error: %v", err2)
	}

	// Verify they're identical
	if len(pop1) != len(pop2) {
		t.Fatalf("population sizes differ: %d vs %d", len(pop1), len(pop2))
	}

	for i := 0; i < len(pop1); i++ {
		chr1 := pop1[i]
		chr2 := pop2[i]

		if len(chr1.Groups) != len(chr2.Groups) {
			t.Errorf("chromosome %d: different number of groups", i)
			continue
		}

		// Check each group
		for j := 0; j < len(chr1.Groups); j++ {
			if len(chr1.Groups[j]) != len(chr2.Groups[j]) {
				t.Errorf("chromosome %d, group %d: different sizes", i, j)
				continue
			}

			// Check student order
			for k := 0; k < len(chr1.Groups[j]); k++ {
				if chr1.Groups[j][k].ID != chr2.Groups[j][k].ID {
					t.Errorf("chromosome %d, group %d, position %d: different students (%s vs %s)",
						i, j, k, chr1.Groups[j][k].ID, chr2.Groups[j][k].ID)
				}
			}
		}
	}
}

func TestInitPopulation_DifferentSeeds(t *testing.T) {
	students := createTestStudents(16)
	targetSize := 4
	populationSize := 5

	pop1, _ := InitPopulation(students, targetSize, populationSize, 100)
	pop2, _ := InitPopulation(students, targetSize, populationSize, 200)

	// They should be different (at least some chromosomes)
	allIdentical := true
	for i := 0; i < len(pop1); i++ {
		if !chromosomesEqual(pop1[i], pop2[i]) {
			allIdentical = false
			break
		}
	}

	if allIdentical {
		t.Error("populations with different seeds should produce different results")
	}
}

func TestInitPopulation_InvalidInputs(t *testing.T) {
	students := createTestStudents(10)

	tests := []struct {
		name           string
		students       []*domain.Student
		targetSize     int
		populationSize int
		seed           int64
		expectError    bool
	}{
		{
			name:           "empty students",
			students:       []*domain.Student{},
			targetSize:     4,
			populationSize: 5,
			seed:           1,
			expectError:    true,
		},
		{
			name:           "nil students",
			students:       nil,
			targetSize:     4,
			populationSize: 5,
			seed:           1,
			expectError:    true,
		},
		{
			name:           "zero target size",
			students:       students,
			targetSize:     0,
			populationSize: 5,
			seed:           1,
			expectError:    true,
		},
		{
			name:           "negative target size",
			students:       students,
			targetSize:     -1,
			populationSize: 5,
			seed:           1,
			expectError:    true,
		},
		{
			name:           "zero population size",
			students:       students,
			targetSize:     4,
			populationSize: 0,
			seed:           1,
			expectError:    true,
		},
		{
			name:           "negative population size",
			students:       students,
			targetSize:     4,
			populationSize: -5,
			seed:           1,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := InitPopulation(tt.students, tt.targetSize, tt.populationSize, tt.seed)
			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestInitPopulationWithDistribution(t *testing.T) {
	students := createTestStudents(26)
	distribution := []int{4, 4, 4, 4, 5, 5}
	populationSize := 10
	seed := int64(42)

	population, err := InitPopulationWithDistribution(students, distribution, populationSize, seed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(population) != populationSize {
		t.Errorf("expected population size %d, got %d", populationSize, len(population))
	}

	for i, chromosome := range population {
		// Verify distribution
		if len(chromosome.Groups) != len(distribution) {
			t.Errorf("chromosome %d: expected %d groups, got %d", i, len(distribution), len(chromosome.Groups))
		}

		// Verify each group size matches distribution
		for j, group := range chromosome.Groups {
			if len(group) != distribution[j] {
				t.Errorf("chromosome %d, group %d: expected size %d, got %d",
					i, j, distribution[j], len(group))
			}
		}

		// Validate no duplicates
		if err := ValidateChromosome(chromosome, students); err != nil {
			t.Errorf("chromosome %d failed validation: %v", i, err)
		}
	}
}

func TestInitPopulationWithDistribution_InvalidDistribution(t *testing.T) {
	students := createTestStudents(20)

	tests := []struct {
		name         string
		distribution []int
		expectError  bool
	}{
		{
			name:         "distribution sum mismatch",
			distribution: []int{4, 4, 4}, // sums to 12, not 20
			expectError:  true,
		},
		{
			name:         "empty distribution",
			distribution: []int{},
			expectError:  true,
		},
		{
			name:         "negative size",
			distribution: []int{4, -2, 18},
			expectError:  true,
		},
		{
			name:         "zero size",
			distribution: []int{10, 0, 10},
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := InitPopulationWithDistribution(students, tt.distribution, 5, 42)
			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateChromosome(t *testing.T) {
	students := createTestStudents(12)

	tests := []struct {
		name        string
		setupFunc   func() *domain.Chromosome
		expectError bool
	}{
		{
			name: "valid chromosome",
			setupFunc: func() *domain.Chromosome {
				return &domain.Chromosome{
					Groups: []domain.Group{
						{students[0], students[1], students[2], students[3]},
						{students[4], students[5], students[6], students[7]},
						{students[8], students[9], students[10], students[11]},
					},
				}
			},
			expectError: false,
		},
		{
			name: "nil chromosome",
			setupFunc: func() *domain.Chromosome {
				return nil
			},
			expectError: true,
		},
		{
			name: "empty groups",
			setupFunc: func() *domain.Chromosome {
				return &domain.Chromosome{Groups: []domain.Group{}}
			},
			expectError: true,
		},
		{
			name: "duplicate student",
			setupFunc: func() *domain.Chromosome {
				return &domain.Chromosome{
					Groups: []domain.Group{
						{students[0], students[1], students[2]},
						{students[2], students[3], students[4]}, // students[2] duplicated
					},
				}
			},
			expectError: true,
		},
		{
			name: "missing student",
			setupFunc: func() *domain.Chromosome {
				return &domain.Chromosome{
					Groups: []domain.Group{
						{students[0], students[1], students[2]},
						{students[3], students[4], students[5]},
						// Missing students[6] through students[11]
					},
				}
			},
			expectError: true,
		},
		{
			name: "nil student in group",
			setupFunc: func() *domain.Chromosome {
				return &domain.Chromosome{
					Groups: []domain.Group{
						{students[0], nil, students[2]},
					},
				}
			},
			expectError: true,
		},
		{
			name: "empty group",
			setupFunc: func() *domain.Chromosome {
				return &domain.Chromosome{
					Groups: []domain.Group{
						{students[0], students[1]},
						{}, // Empty group
					},
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chromosome := tt.setupFunc()
			err := ValidateChromosome(chromosome, students)
			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGetDistributionStats(t *testing.T) {
	tests := []struct {
		name         string
		distribution []int
		checkFunc    func(map[string]interface{}) bool
	}{
		{
			name:         "balanced distribution",
			distribution: []int{4, 4, 4},
			checkFunc: func(stats map[string]interface{}) bool {
				return stats["num_groups"] == 3 &&
					stats["min_size"] == 4 &&
					stats["max_size"] == 4 &&
					stats["total_slots"] == 12 &&
					stats["is_balanced"] == true
			},
		},
		{
			name:         "slightly unbalanced",
			distribution: []int{4, 4, 5, 5},
			checkFunc: func(stats map[string]interface{}) bool {
				return stats["num_groups"] == 4 &&
					stats["min_size"] == 4 &&
					stats["max_size"] == 5 &&
					stats["total_slots"] == 18 &&
					stats["is_balanced"] == true // max-min = 1
			},
		},
		{
			name:         "empty distribution",
			distribution: []int{},
			checkFunc: func(stats map[string]interface{}) bool {
				return stats["num_groups"] == 0 &&
					stats["total_slots"] == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := GetDistributionStats(tt.distribution)
			if !tt.checkFunc(stats) {
				t.Errorf("stats validation failed: %v", stats)
			}
		})
	}
}

func TestShuffleStudents(t *testing.T) {
	students := createTestStudents(10)
	originalOrder := make([]string, len(students))
	for i, s := range students {
		originalOrder[i] = s.ID
	}

	// Create two copies
	copy1 := make([]*domain.Student, len(students))
	copy2 := make([]*domain.Student, len(students))
	copy(copy1, students)
	copy(copy2, students)

	// Shuffle with different seeds
	rng1 := createRNG(100)
	rng2 := createRNG(200)

	shuffleStudents(copy1, rng1)
	shuffleStudents(copy2, rng2)

	// Check that order changed for at least one copy
	changed1 := false
	changed2 := false

	for i := range students {
		if copy1[i].ID != originalOrder[i] {
			changed1 = true
		}
		if copy2[i].ID != originalOrder[i] {
			changed2 = true
		}
	}

	if !changed1 && !changed2 {
		t.Error("shuffle didn't change order")
	}

	// Verify all students still present in copy1
	ids1 := make(map[string]bool)
	for _, s := range copy1 {
		ids1[s.ID] = true
	}
	if len(ids1) != len(students) {
		t.Error("shuffle lost students")
	}
}

// Helper functions

func createTestStudents(n int) []*domain.Student {
	students := make([]*domain.Student, n)
	for i := 0; i < n; i++ {
		students[i] = &domain.Student{
			ID:               fmt.Sprintf("S%03d", i+1),
			NumericAttrs:     make(map[string]float64),
			CategoricalAttrs: make(map[string]string),
		}
	}
	return students
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func chromosomesEqual(c1, c2 *domain.Chromosome) bool {
	if len(c1.Groups) != len(c2.Groups) {
		return false
	}

	for i := 0; i < len(c1.Groups); i++ {
		if len(c1.Groups[i]) != len(c2.Groups[i]) {
			return false
		}
		for j := 0; j < len(c1.Groups[i]); j++ {
			if c1.Groups[i][j].ID != c2.Groups[i][j].ID {
				return false
			}
		}
	}

	return true
}

func createRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

func BenchmarkCalculateGroupDistribution(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = CalculateGroupDistribution(100, 5)
	}
}

func BenchmarkInitPopulation(b *testing.B) {
	students := createTestStudents(100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = InitPopulation(students, 5, 20, int64(i))
	}
}

func BenchmarkValidateChromosome(b *testing.B) {
	students := createTestStudents(100)
	population, _ := InitPopulation(students, 5, 1, 42)
	chromosome := population[0]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateChromosome(chromosome, students)
	}
}
