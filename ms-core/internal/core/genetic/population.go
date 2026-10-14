package genetic

import (
	"fmt"
	"math/rand"

	"ms-core/internal/core/domain"
)

// buildEvenDistribution returns a slice of numGroups elements all equal to groupSize.
func buildEvenDistribution(numGroups, groupSize int) []int {
	d := make([]int, numGroups)
	for i := range d {
		d[i] = groupSize
	}
	return d
}

// buildDistributionWithRemainder builds (numGroups-remainder) groups of size K
// and remainder groups of size K+1, avoiding singletons.
func buildDistributionWithRemainder(numGroups, groupSize, remainder int) []int {
	numSmaller := numGroups - remainder
	numLarger := remainder
	d := make([]int, 0, numGroups)
	for i := 0; i < numSmaller; i++ {
		d = append(d, groupSize)
	}
	for i := 0; i < numLarger; i++ {
		d = append(d, groupSize+1)
	}
	return d
}

// CalculateGroupDistribution computes the optimal distribution of students into groups.
// Given N students and desired group size K, it returns a slice of group sizes that:
// - Minimizes size variance
// - Ensures all students are assigned
// - Prefers slightly larger groups over leaving singletons
//
// Example: N=26, K=4 → [4, 4, 4, 4, 5, 5] (26 students total)
// Example: N=10, K=3 → [3, 3, 4] (10 students total)
// Example: N=12, K=4 → [4, 4, 4] (12 students total)
func CalculateGroupDistribution(numStudents, targetGroupSize int) ([]int, error) {
	if numStudents <= 0 {
		return nil, fmt.Errorf("number of students must be positive, got %d", numStudents)
	}

	if targetGroupSize <= 0 {
		return nil, fmt.Errorf("target group size must be positive, got %d", targetGroupSize)
	}

	if targetGroupSize > numStudents {
		// If target size is larger than total students, return one group with all students
		return []int{numStudents}, nil
	}

	// Calculate base number of groups and remainder
	baseNumGroups := numStudents / targetGroupSize
	remainder := numStudents % targetGroupSize

	var distribution []int
	switch {
	case remainder == 0:
		distribution = buildEvenDistribution(baseNumGroups, targetGroupSize)
	case remainder == 1 && baseNumGroups > 1:
		// Special case: avoid singleton — merge extra student into last group
		distribution = buildEvenDistribution(baseNumGroups-1, targetGroupSize)
		distribution = append(distribution, targetGroupSize+1)
	default:
		distribution = buildDistributionWithRemainder(baseNumGroups, targetGroupSize, remainder)
	}

	// Validate distribution
	total := 0
	for _, size := range distribution {
		total += size
	}

	if total != numStudents {
		return nil, fmt.Errorf("distribution error: expected %d students, got %d (distribution: %v)",
			numStudents, total, distribution)
	}

	return distribution, nil
}

// buildChromosomeFromStudents distributes the provided (already-shuffled) students
// into groups according to distribution and returns the resulting Chromosome.
func buildChromosomeFromStudents(students []*domain.Student, distribution []int) (*domain.Chromosome, error) {
	groups := make([]domain.Group, 0, len(distribution))
	idx := 0

	for _, groupSize := range distribution {
		group := make(domain.Group, 0, groupSize)
		for j := 0; j < groupSize; j++ {
			if idx >= len(students) {
				return nil, fmt.Errorf("distribution error: ran out of students at index %d", idx)
			}
			group = append(group, students[idx])
			idx++
		}
		groups = append(groups, group)
	}

	chromosome := &domain.Chromosome{
		Groups:  groups,
		Fitness: 0.0,
		Stats:   make(map[string]interface{}),
	}
	chromosome.Stats["distribution"] = distribution
	chromosome.Stats["num_groups"] = len(groups)
	return chromosome, nil
}

// InitPopulation creates an initial population of chromosomes with deterministic seeding.
// It shuffles students and distributes them into balanced groups according to the distribution.
// Each student appears exactly once per chromosome.
//
// Parameters:
// - students: slice of student pointers to be distributed
// - targetGroupSize: desired group size (actual sizes may vary slightly for balance)
// - populationSize: number of chromosomes to generate
// - seed: random seed for reproducibility (use different seeds for different runs)
//
// Returns:
// - slice of chromosomes, each containing balanced groups
// - error if inputs are invalid
func InitPopulation(students []*domain.Student, targetGroupSize, populationSize int, seed int64) ([]*domain.Chromosome, error) {
	if len(students) == 0 {
		return nil, fmt.Errorf("cannot initialize population with zero students")
	}

	if targetGroupSize <= 0 {
		return nil, fmt.Errorf("target group size must be positive, got %d", targetGroupSize)
	}

	if populationSize <= 0 {
		return nil, fmt.Errorf("population size must be positive, got %d", populationSize)
	}

	// Calculate optimal group distribution
	distribution, err := CalculateGroupDistribution(len(students), targetGroupSize)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate group distribution: %w", err)
	}

	// Create local random generator with seed (not global)
	rng := rand.New(rand.NewSource(seed))

	population := make([]*domain.Chromosome, 0, populationSize)

	for i := 0; i < populationSize; i++ {
		studentsCopy := make([]*domain.Student, len(students))
		copy(studentsCopy, students)

		rng.Seed(seed + int64(i))
		shuffleStudents(studentsCopy, rng)

		chromosome, buildErr := buildChromosomeFromStudents(studentsCopy, distribution)
		if buildErr != nil {
			return nil, buildErr
		}
		chromosome.Stats["seed_offset"] = i

		// Verify all students were assigned
		if totalAssigned := countAssigned(chromosome); totalAssigned != len(students) {
			return nil, fmt.Errorf("distribution error: assigned %d students, expected %d", totalAssigned, len(students))
		}

		population = append(population, chromosome)
	}

	return population, nil
}

// countAssigned returns the total number of students in all groups of a chromosome.
func countAssigned(chromosome *domain.Chromosome) int {
	total := 0
	for _, g := range chromosome.Groups {
		total += len(g)
	}
	return total
}

// InitPopulationWithDistribution creates a population using a specific group size distribution.
// This allows manual control over group sizes (e.g., [4, 4, 5, 5] for mixed sizes).
//
// Parameters:
// - students: slice of student pointers
// - distribution: explicit group sizes (must sum to len(students))
// - populationSize: number of chromosomes to generate
// - seed: random seed for reproducibility
func InitPopulationWithDistribution(students []*domain.Student, distribution []int, populationSize int, seed int64) ([]*domain.Chromosome, error) {
	if len(students) == 0 {
		return nil, fmt.Errorf("cannot initialize population with zero students")
	}

	if len(distribution) == 0 {
		return nil, fmt.Errorf("distribution cannot be empty")
	}

	if populationSize <= 0 {
		return nil, fmt.Errorf("population size must be positive, got %d", populationSize)
	}

	// Validate distribution sums to total students
	total := 0
	for _, size := range distribution {
		if size <= 0 {
			return nil, fmt.Errorf("all group sizes must be positive, got %d in distribution", size)
		}
		total += size
	}

	if total != len(students) {
		return nil, fmt.Errorf("distribution sum (%d) does not match number of students (%d)", total, len(students))
	}

	// Create local random generator
	rng := rand.New(rand.NewSource(seed))

	population := make([]*domain.Chromosome, 0, populationSize)

	for i := 0; i < populationSize; i++ {
		studentsCopy := make([]*domain.Student, len(students))
		copy(studentsCopy, students)

		rng.Seed(seed + int64(i))
		shuffleStudents(studentsCopy, rng)

		chromosome, buildErr := buildChromosomeFromStudents(studentsCopy, distribution)
		if buildErr != nil {
			return nil, buildErr
		}
		chromosome.Stats["seed_offset"] = i

		population = append(population, chromosome)
	}

	return population, nil
}

// shuffleStudents performs Fisher-Yates shuffle on student slice using provided RNG.
// This ensures deterministic shuffling with seeded random generator.
func shuffleStudents(students []*domain.Student, rng *rand.Rand) {
	n := len(students)
	for i := n - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		students[i], students[j] = students[j], students[i]
	}
}

// checkGroupContents validates that a single group has no nil students and no duplicates.
// It updates the seen map in place and returns an error on the first violation.
func checkGroupContents(group domain.Group, seen map[string]bool, idx int) error {
	if len(group) == 0 {
		return fmt.Errorf("group %d is empty", idx)
	}
	for _, student := range group {
		if student == nil {
			return fmt.Errorf("group %d contains nil student", idx)
		}
		if seen[student.ID] {
			return fmt.Errorf("student %s appears multiple times", student.ID)
		}
		seen[student.ID] = true
	}
	return nil
}

// ValidateChromosome checks that a chromosome has valid group assignments.
// It verifies:
// - No student appears more than once
// - All students are assigned
// - No empty groups
func ValidateChromosome(chromosome *domain.Chromosome, expectedStudents []*domain.Student) error {
	if chromosome == nil {
		return fmt.Errorf("chromosome is nil")
	}

	if len(chromosome.Groups) == 0 {
		return fmt.Errorf("chromosome has no groups")
	}

	// Track seen students
	seen := make(map[string]bool)
	totalAssigned := 0

	for i, group := range chromosome.Groups {
		if err := checkGroupContents(group, seen, i); err != nil {
			return err
		}
		totalAssigned += len(group)
	}

	// Verify all expected students are present
	if totalAssigned != len(expectedStudents) {
		return fmt.Errorf("chromosome has %d students, expected %d", totalAssigned, len(expectedStudents))
	}

	for _, student := range expectedStudents {
		if !seen[student.ID] {
			return fmt.Errorf("student %s is missing from chromosome", student.ID)
		}
	}

	return nil
}

// GetDistributionStats returns statistics about a group distribution.
func GetDistributionStats(distribution []int) map[string]interface{} {
	if len(distribution) == 0 {
		return map[string]interface{}{
			"num_groups":  0,
			"min_size":    0,
			"max_size":    0,
			"mean_size":   0.0,
			"variance":    0.0,
			"total_slots": 0,
		}
	}

	min := distribution[0]
	max := distribution[0]
	total := 0

	for _, size := range distribution {
		if size < min {
			min = size
		}
		if size > max {
			max = size
		}
		total += size
	}

	mean := float64(total) / float64(len(distribution))

	// Calculate variance
	variance := 0.0
	for _, size := range distribution {
		diff := float64(size) - mean
		variance += diff * diff
	}
	variance /= float64(len(distribution))

	return map[string]interface{}{
		"num_groups":  len(distribution),
		"min_size":    min,
		"max_size":    max,
		"mean_size":   mean,
		"variance":    variance,
		"total_slots": total,
		"is_balanced": max-min <= 1, // Considered balanced if sizes differ by at most 1
	}
}
