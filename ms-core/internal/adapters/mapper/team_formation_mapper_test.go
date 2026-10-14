package mapper

import (
	"testing"

	"ms-core/internal/adapters/dto"
	"ms-core/internal/core/domain"
)

func TestNewTeamFormationMapper(t *testing.T) {
	mapper := NewTeamFormationMapper()

	if mapper == nil {
		t.Fatal("Expected mapper to be created, got nil")
	}

	if mapper.numericCriteria == nil {
		t.Error("Expected numericCriteria to be initialized")
	}

	if mapper.categoricalCriteria == nil {
		t.Error("Expected categoricalCriteria to be initialized")
	}

	if mapper.socialCriteria == nil {
		t.Error("Expected socialCriteria to be initialized")
	}

	// Verify some known criteria are registered
	if !mapper.numericCriteria["age"] {
		t.Error("Expected 'age' to be in numeric criteria catalog")
	}

	if !mapper.categoricalCriteria["sex"] {
		t.Error("Expected 'sex' to be in categorical criteria catalog")
	}

	if !mapper.socialCriteria["social"] {
		t.Error("Expected 'social' to be in social criteria catalog")
	}
}

func TestDetectCriterionType(t *testing.T) {
	mapper := NewTeamFormationMapper()

	tests := []struct {
		name          string
		criterionName string
		expectedType  CriterionType
		shouldError   bool
	}{
		// Numeric criteria
		{name: "age", criterionName: "age", expectedType: CriterionTypeNumeric, shouldError: false},
		{name: "edad", criterionName: "edad", expectedType: CriterionTypeNumeric, shouldError: false},
		{name: "gpa", criterionName: "gpa", expectedType: CriterionTypeNumeric, shouldError: false},
		{name: "openness", criterionName: "openness", expectedType: CriterionTypeNumeric, shouldError: false},
		{name: "conscientiousness", criterionName: "conscientiousness", expectedType: CriterionTypeNumeric, shouldError: false},
		{name: "extraversion", criterionName: "extraversion", expectedType: CriterionTypeNumeric, shouldError: false},
		{name: "agreeableness", criterionName: "agreeableness", expectedType: CriterionTypeNumeric, shouldError: false},
		{name: "neuroticism", criterionName: "neuroticism", expectedType: CriterionTypeNumeric, shouldError: false},

		// Categorical criteria
		{name: "sex", criterionName: "sex", expectedType: CriterionTypeCategorical, shouldError: false},
		{name: "gender", criterionName: "gender", expectedType: CriterionTypeCategorical, shouldError: false},
		{name: "career_code", criterionName: "career_code", expectedType: CriterionTypeCategorical, shouldError: false},
		{name: "major", criterionName: "major", expectedType: CriterionTypeCategorical, shouldError: false},
		{name: "language", criterionName: "language", expectedType: CriterionTypeCategorical, shouldError: false},
		{name: "mbti", criterionName: "mbti", expectedType: CriterionTypeCategorical, shouldError: false},
		{name: "enneagram", criterionName: "enneagram", expectedType: CriterionTypeCategorical, shouldError: false},

		// Social criteria
		{name: "social", criterionName: "social", expectedType: CriterionTypeSocial, shouldError: false},
		{name: "sociometric", criterionName: "sociometric", expectedType: CriterionTypeSocial, shouldError: false},
		{name: "sociogram", criterionName: "sociogram", expectedType: CriterionTypeSocial, shouldError: false},
		{name: "affinity", criterionName: "affinity", expectedType: CriterionTypeSocial, shouldError: false},

		// Case insensitivity
		{name: "AGE uppercase", criterionName: "AGE", expectedType: CriterionTypeNumeric, shouldError: false},
		{name: "Sex mixed case", criterionName: "Sex", expectedType: CriterionTypeCategorical, shouldError: false},
		{name: "SOCIAL uppercase", criterionName: "SOCIAL", expectedType: CriterionTypeSocial, shouldError: false},

		// Whitespace handling
		{name: "age with spaces", criterionName: "  age  ", expectedType: CriterionTypeNumeric, shouldError: false},

		// Unknown criterion
		{name: "unknown", criterionName: "unknown_criterion", expectedType: "", shouldError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			criterionType, err := mapper.detectCriterionType(tt.criterionName)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for criterion '%s', got nil", tt.criterionName)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for criterion '%s': %v", tt.criterionName, err)
				}

				if criterionType != tt.expectedType {
					t.Errorf("Expected type %s for criterion '%s', got %s", tt.expectedType, tt.criterionName, criterionType)
				}
			}
		})
	}
}

func TestMapStudents(t *testing.T) {
	mapper := NewTeamFormationMapper()

	t.Run("successful mapping with complete data", func(t *testing.T) {
		dto := &dto.TeamFormationRequestDTO{
			Candidates: []string{"student1", "student2"},
			DemographicData: map[string]dto.StudentAttributesDTO{
				"student1": {
					NumericAttrs: map[string]float64{
						"age":          20.0,
						"openness":     0.8,
						"extraversion": 0.6,
					},
					CategoricalAttrs: map[string]string{
						"sex":         "M",
						"career_code": "ING_INF",
					},
				},
				"student2": {
					NumericAttrs: map[string]float64{
						"age":      22.0,
						"openness": 0.7,
					},
					CategoricalAttrs: map[string]string{
						"sex": "F",
					},
				},
			},
		}

		students, err := mapper.mapStudents(dto)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(students) != 2 {
			t.Fatalf("Expected 2 students, got %d", len(students))
		}

		// Verify student1
		if students[0].ID != "student1" {
			t.Errorf("Expected student ID 'student1', got '%s'", students[0].ID)
		}

		if students[0].NumericAttrs["age"] != 20.0 {
			t.Errorf("Expected age 20.0, got %.1f", students[0].NumericAttrs["age"])
		}

		if students[0].CategoricalAttrs["sex"] != "M" {
			t.Errorf("Expected sex 'M', got '%s'", students[0].CategoricalAttrs["sex"])
		}

		// Verify student2
		if students[1].ID != "student2" {
			t.Errorf("Expected student ID 'student2', got '%s'", students[1].ID)
		}
	})

	t.Run("missing demographic data creates empty student", func(t *testing.T) {
		dto := &dto.TeamFormationRequestDTO{
			Candidates: []string{"student1", "student2"},
			DemographicData: map[string]dto.StudentAttributesDTO{
				"student1": {
					NumericAttrs: map[string]float64{"age": 20.0},
				},
				// student2 missing
			},
		}

		students, err := mapper.mapStudents(dto)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(students) != 2 {
			t.Fatalf("Expected 2 students, got %d", len(students))
		}

		// student2 should exist with empty maps
		if students[1].ID != "student2" {
			t.Errorf("Expected student ID 'student2', got '%s'", students[1].ID)
		}

		if len(students[1].NumericAttrs) != 0 {
			t.Errorf("Expected empty NumericAttrs for missing student, got %d entries", len(students[1].NumericAttrs))
		}
	})

	t.Run("empty candidates returns error", func(t *testing.T) {
		dto := &dto.TeamFormationRequestDTO{
			Candidates:      []string{},
			DemographicData: map[string]dto.StudentAttributesDTO{},
		}

		students, err := mapper.mapStudents(dto)

		if err == nil {
			t.Error("Expected error for empty candidates, got nil")
		}

		if students != nil {
			t.Errorf("Expected nil students for error case, got %d students", len(students))
		}
	})
}

func TestBuildConfig(t *testing.T) {
	mapper := NewTeamFormationMapper()

	t.Run("successful config with normalized weights", func(t *testing.T) {
		criteriaParams := map[string]dto.CriterionParameterDTO{
			"age": {
				Target: 0.5,
				Weight: 0.3,
			},
			"sex": {
				Target: 0.6,
				Weight: 0.5,
			},
			"social": {
				Target: 0.7,
				Weight: 0.2,
			},
		}

		config, err := mapper.buildConfig(criteriaParams)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Check that weights are normalized (sum to 1.0)
		totalWeight := 0.0
		for _, weight := range config.Weights {
			totalWeight += weight
		}

		if totalWeight < 0.99 || totalWeight > 1.01 {
			t.Errorf("Expected weights to sum to ~1.0, got %.2f", totalWeight)
		}

		// Check individual normalized weights
		// Original: age=0.3, sex=0.5, social=0.2, total=1.0
		// Normalized: age=0.3, sex=0.5, social=0.2 (already normalized)
		if config.Weights["age"] < 0.29 || config.Weights["age"] > 0.31 {
			t.Errorf("Expected age weight ~0.3, got %.2f", config.Weights["age"])
		}

		// Check targets are preserved
		if config.Targets["age"] != 0.5 {
			t.Errorf("Expected age target 0.5, got %.1f", config.Targets["age"])
		}

		if config.Targets["sex"] != 0.6 {
			t.Errorf("Expected sex target 0.6, got %.1f", config.Targets["sex"])
		}
	})

	t.Run("zero total weight returns error", func(t *testing.T) {
		criteriaParams := map[string]dto.CriterionParameterDTO{
			"age": {Target: 0.5, Weight: 0.0},
			"sex": {Target: 0.6, Weight: 0.0},
		}

		config, err := mapper.buildConfig(criteriaParams)

		if err == nil {
			t.Error("Expected error for zero total weight, got nil")
		}

		if config != nil {
			t.Error("Expected nil config for error case")
		}
	})
}

func TestBuildCriterionFactory(t *testing.T) {
	mapper := NewTeamFormationMapper()

	t.Run("successful factory creation with mixed criteria types", func(t *testing.T) {
		criteriaParams := map[string]dto.CriterionParameterDTO{
			"age":    {Target: 0.5, Weight: 0.3},
			"sex":    {Target: 0.6, Weight: 0.4},
			"social": {Target: 0.7, Weight: 0.3},
		}

		factory, err := mapper.buildCriterionFactory(criteriaParams)

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if factory == nil {
			t.Fatal("Expected factory to be created, got nil")
		}

		// Verify criteria are registered (by trying to get them)
		ageCriterion := factory.Get("age")
		if ageCriterion == nil {
			t.Error("Expected 'age' criterion to be registered")
		}

		sexCriterion := factory.Get("sex")
		if sexCriterion == nil {
			t.Error("Expected 'sex' criterion to be registered")
		}

		socialCriterion := factory.Get("social")
		if socialCriterion == nil {
			t.Error("Expected 'social' criterion to be registered")
		}
	})

	t.Run("unknown criterion type returns error", func(t *testing.T) {
		criteriaParams := map[string]dto.CriterionParameterDTO{
			"unknown_criterion": {Target: 0.5, Weight: 0.5},
		}

		factory, err := mapper.buildCriterionFactory(criteriaParams)

		if err == nil {
			t.Error("Expected error for unknown criterion type, got nil")
		}

		if factory != nil {
			t.Error("Expected nil factory for error case")
		}
	})
}

func TestBuildContext(t *testing.T) {
	mapper := NewTeamFormationMapper()

	dto := &dto.TeamFormationRequestDTO{
		SectionID: "section123",
		TeamSize:  4,
		Algorithm: "genetic",
		DemographicData: map[string]dto.StudentAttributesDTO{
			"student1": {
				SocialConnections: map[string]float64{
					"student2": 0.8,
					"student3": 0.6,
				},
			},
			"student2": {
				SocialConnections: map[string]float64{
					"student1": 0.7,
				},
			},
		},
	}

	config := &domain.Config{
		Targets: map[string]float64{
			"age": 0.5,
			"sex": 0.6,
		},
	}

	ctx := mapper.buildContext(dto, config)

	// Check social graph is built
	if len(ctx.SocialGraph) != 2 {
		t.Errorf("Expected 2 entries in SocialGraph, got %d", len(ctx.SocialGraph))
	}

	if ctx.SocialGraph["student1"]["student2"] != 0.8 {
		t.Errorf("Expected student1->student2 affinity 0.8, got %.1f", ctx.SocialGraph["student1"]["student2"])
	}

	// Check targets are copied
	if ctx.Targets["age"] != 0.5 {
		t.Errorf("Expected age target 0.5, got %.1f", ctx.Targets["age"])
	}

	// Check constraints are set
	if ctx.Constraints["section_id"] != "section123" {
		t.Errorf("Expected section_id 'section123', got %v", ctx.Constraints["section_id"])
	}

	if ctx.Constraints["team_size"] != 4 {
		t.Errorf("Expected team_size 4, got %v", ctx.Constraints["team_size"])
	}
}

func TestMapToDomain_Integration(t *testing.T) {
	mapper := NewTeamFormationMapper()

	dto := &dto.TeamFormationRequestDTO{
		SectionID:  "section123",
		Candidates: []string{"student1", "student2", "student3"},
		TeamSize:   3,
		CriteriaParameters: map[string]dto.CriterionParameterDTO{
			"age": {
				Target: 0.5,
				Weight: 0.3,
			},
			"sex": {
				Target: 0.6,
				Weight: 0.4,
			},
			"social": {
				Target: 0.7,
				Weight: 0.3,
			},
		},
		DemographicData: map[string]dto.StudentAttributesDTO{
			"student1": {
				NumericAttrs: map[string]float64{
					"age": 20.0,
				},
				CategoricalAttrs: map[string]string{
					"sex": "M",
				},
				SocialConnections: map[string]float64{
					"student2": 0.8,
				},
			},
			"student2": {
				NumericAttrs: map[string]float64{
					"age": 22.0,
				},
				CategoricalAttrs: map[string]string{
					"sex": "F",
				},
			},
			"student3": {
				NumericAttrs: map[string]float64{
					"age": 21.0,
				},
				CategoricalAttrs: map[string]string{
					"sex": "M",
				},
			},
		},
	}

	students, config, factory, ctx, err := mapper.MapToDomain(dto)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify students
	if len(students) != 3 {
		t.Errorf("Expected 3 students, got %d", len(students))
	}

	// Verify config
	if config == nil {
		t.Fatal("Expected config to be created")
	}

	totalWeight := 0.0
	for _, weight := range config.Weights {
		totalWeight += weight
	}

	if totalWeight < 0.99 || totalWeight > 1.01 {
		t.Errorf("Expected weights to sum to ~1.0, got %.2f", totalWeight)
	}

	// Verify factory
	if factory == nil {
		t.Fatal("Expected factory to be created")
	}

	if factory.Get("age") == nil {
		t.Error("Expected 'age' criterion to be registered")
	}

	// Verify context
	if len(ctx.SocialGraph) == 0 {
		t.Error("Expected social graph to be populated")
	}
}

func TestMapToDomain_HomogeneityTarget(t *testing.T) {
	mapper := NewTeamFormationMapper()

	// Test with target = 0.0 (want homogeneity, all same sex)
	homogeneousDTO := &dto.TeamFormationRequestDTO{
		SectionID:  "section_homogeneity_test",
		Candidates: []string{"s1", "s2", "s3", "s4"},
		TeamSize:   4,
		CriteriaParameters: map[string]dto.CriterionParameterDTO{
			"sex": dto.CriterionParameterDTO{
				Target: 0.0,
				Weight: 1.0,
			},
		},
		DemographicData: map[string]dto.StudentAttributesDTO{
			"s1": dto.StudentAttributesDTO{CategoricalAttrs: map[string]string{"sex": "M"}},
			"s2": dto.StudentAttributesDTO{CategoricalAttrs: map[string]string{"sex": "M"}},
			"s3": dto.StudentAttributesDTO{CategoricalAttrs: map[string]string{"sex": "M"}},
			"s4": dto.StudentAttributesDTO{CategoricalAttrs: map[string]string{"sex": "M"}},
		},
	}

	students, _, factory, ctx, err := mapper.MapToDomain(homogeneousDTO)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify target is set correctly
	if ctx.Targets["sex"] != 0.0 {
		t.Errorf("Expected sex target to be 0.0, got %f", ctx.Targets["sex"])
	}

	// Create homogeneous group and check score
	criterion := factory.Get("sex")
	homogeneousGroup := domain.Group{students[0], students[1], students[2], students[3]}
	homogeneousScore := criterion.Calculate(&homogeneousGroup, ctx)

	if homogeneousScore < 0.9 {
		t.Errorf("Expected high score for homogeneous group with target=0.0, got %f", homogeneousScore)
	}
}

func TestRegisterCustomCriterion(t *testing.T) {
	mapper := NewTeamFormationMapper()

	// Register a custom numeric criterion
	mapper.RegisterCustomCriterion("custom_score", CriterionTypeNumeric)

	criterionType, err := mapper.GetCriterionType("custom_score")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if criterionType != CriterionTypeNumeric {
		t.Errorf("Expected type %s, got %s", CriterionTypeNumeric, criterionType)
	}

	// Register a custom categorical criterion
	mapper.RegisterCustomCriterion("custom_category", CriterionTypeCategorical)

	criterionType, err = mapper.GetCriterionType("custom_category")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if criterionType != CriterionTypeCategorical {
		t.Errorf("Expected type %s, got %s", CriterionTypeCategorical, criterionType)
	}
}

func TestCatalogCompleteness(t *testing.T) {
	mapper := NewTeamFormationMapper()

	// Test that common English and Spanish criteria are recognized
	commonCriteria := []struct {
		name         string
		expectedType CriterionType
	}{
		{"age", CriterionTypeNumeric},
		{"edad", CriterionTypeNumeric},
		{"sex", CriterionTypeCategorical},
		{"sexo", CriterionTypeCategorical},
		{"gender", CriterionTypeCategorical},
		{"genero", CriterionTypeCategorical},
		{"social", CriterionTypeSocial},
		{"openness", CriterionTypeNumeric},
		{"apertura", CriterionTypeNumeric},
		{"career_code", CriterionTypeCategorical},
		{"codigo_carrera", CriterionTypeCategorical},
	}

	for _, tc := range commonCriteria {
		t.Run(tc.name, func(t *testing.T) {
			criterionType, err := mapper.detectCriterionType(tc.name)

			if err != nil {
				t.Errorf("Expected criterion '%s' to be recognized, got error: %v", tc.name, err)
			}

			if criterionType != tc.expectedType {
				t.Errorf("Expected type %s for '%s', got %s", tc.expectedType, tc.name, criterionType)
			}
		})
	}
}

// ─── Goal field propagation tests ────────────────────────────────────────────

func TestNumericStrategy(t *testing.T) {
	tests := []struct {
		goal     string
		expected string
	}{
		{"minimize", "minimize"},
		{"maximize", "maximize"},
		{"target", "target"},
		{"homogeneity", "homogeneity"},
		{"heterogeneity", "heterogeneity"},
		{"", "target"},         // empty → default
		{"unknown", "target"},  // unknown → default
		{"MINIMIZE", "target"}, // case-sensitive: not recognized → default
	}

	for _, tc := range tests {
		t.Run("goal="+tc.goal, func(t *testing.T) {
			got := numericStrategy(tc.goal)
			if got != tc.expected {
				t.Errorf("numericStrategy(%q) = %q, want %q", tc.goal, got, tc.expected)
			}
		})
	}
}

func TestBuildCriterionFactory_GoalPropagated(t *testing.T) {
	mapper := NewTeamFormationMapper()

	tests := []struct {
		name         string
		criterionKey string
		goal         string
		// We verify the criterion was created (factory returns non-nil) and
		// that the strategy was applied by calculating scores on a known group.
		// For minimize/maximize we check score ordering; for homogeneity/
		// heterogeneity/target we only verify non-nil creation.
	}{
		{name: "minimize goal accepted", criterionKey: "neuroticism", goal: "minimize"},
		{name: "maximize goal accepted", criterionKey: "agreeableness", goal: "maximize"},
		{name: "target goal accepted", criterionKey: "openness", goal: "target"},
		{name: "homogeneity goal accepted", criterionKey: "conscientiousness", goal: "homogeneity"},
		{name: "heterogeneity goal accepted", criterionKey: "extraversion", goal: "heterogeneity"},
		{name: "empty goal defaults to target", criterionKey: "openness", goal: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			params := map[string]dto.CriterionParameterDTO{
				tc.criterionKey: {Target: 0.5, Weight: 1.0, Goal: tc.goal},
			}
			factory, err := mapper.buildCriterionFactory(params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			criterion := factory.Get(tc.criterionKey)
			if criterion == nil {
				t.Fatalf("expected criterion %q to be registered, got nil", tc.criterionKey)
			}
		})
	}
}

func TestBuildCriterionFactory_MinimizeScoresLowerMeanHigher(t *testing.T) {
	mapper := NewTeamFormationMapper()
	params := map[string]dto.CriterionParameterDTO{
		"neuroticism": {Target: 0.5, Weight: 1.0, Goal: "minimize"},
	}
	factory, err := mapper.buildCriterionFactory(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	criterion := factory.Get("neuroticism")

	ctx := domain.Context{}

	lowGroup := domain.Group{
		{ID: "1", NumericAttrs: map[string]float64{"neuroticism": 0.1}},
		{ID: "2", NumericAttrs: map[string]float64{"neuroticism": 0.1}},
	}
	highGroup := domain.Group{
		{ID: "3", NumericAttrs: map[string]float64{"neuroticism": 0.9}},
		{ID: "4", NumericAttrs: map[string]float64{"neuroticism": 0.9}},
	}

	scoreLow := criterion.Calculate(&lowGroup, ctx)
	scoreHigh := criterion.Calculate(&highGroup, ctx)

	if scoreLow <= scoreHigh {
		t.Errorf("minimize: low-mean group score %.4f should be > high-mean group score %.4f", scoreLow, scoreHigh)
	}
}

func TestBuildCriterionFactory_MaximizeScoresHigherMeanHigher(t *testing.T) {
	mapper := NewTeamFormationMapper()
	params := map[string]dto.CriterionParameterDTO{
		"agreeableness": {Target: 0.5, Weight: 1.0, Goal: "maximize"},
	}
	factory, err := mapper.buildCriterionFactory(params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	criterion := factory.Get("agreeableness")

	ctx := domain.Context{}

	lowGroup := domain.Group{
		{ID: "1", NumericAttrs: map[string]float64{"agreeableness": 0.2}},
		{ID: "2", NumericAttrs: map[string]float64{"agreeableness": 0.2}},
	}
	highGroup := domain.Group{
		{ID: "3", NumericAttrs: map[string]float64{"agreeableness": 0.8}},
		{ID: "4", NumericAttrs: map[string]float64{"agreeableness": 0.8}},
	}

	scoreLow := criterion.Calculate(&lowGroup, ctx)
	scoreHigh := criterion.Calculate(&highGroup, ctx)

	if scoreHigh <= scoreLow {
		t.Errorf("maximize: high-mean group score %.4f should be > low-mean group score %.4f", scoreHigh, scoreLow)
	}
}

func TestBuildCriterionFactory_CategoricalIgnoresGoal(t *testing.T) {
	// Goal is only meaningful for numeric criteria.
	// Categorical criteria must be registered and work correctly regardless of
	// what goal value is present in the DTO (they simply ignore it).
	mapper := NewTeamFormationMapper()
	params := map[string]dto.CriterionParameterDTO{
		"career_code": {Target: 0.8, Weight: 1.0, Goal: "minimize"}, // goal is irrelevant here
	}
	factory, err := mapper.buildCriterionFactory(params)
	if err != nil {
		t.Fatalf("unexpected error building factory with categorical criterion: %v", err)
	}
	criterion := factory.Get("career_code")
	if criterion == nil {
		t.Fatal("expected career_code criterion to be registered")
	}

	// Verify it still behaves as a CategoricalCriterion (Blau-based).
	// A fully homogeneous group should score close to 1.0 when target = 0.0,
	// and a fully heterogeneous group should score close to 1.0 when target = 1.0.
	ctx := domain.Context{Targets: map[string]float64{"career_code": 0.0}}
	homogeneous := domain.Group{
		{ID: "1", CategoricalAttrs: map[string]string{"career_code": "CS"}},
		{ID: "2", CategoricalAttrs: map[string]string{"career_code": "CS"}},
		{ID: "3", CategoricalAttrs: map[string]string{"career_code": "CS"}},
	}
	score := criterion.Calculate(&homogeneous, ctx)
	if score < 0.9 {
		t.Errorf("expected high score for homogeneous group with target=0.0, got %.4f", score)
	}
}

func TestMapToDomain_GoalFlowsToFactory(t *testing.T) {
	// End-to-end: Goal set in CriterionParameterDTO reaches the factory criterion.
	// We verify that minimize/maximize produce different scores than target would
	// for the same demographic data, proving the strategy was actually applied.
	m := NewTeamFormationMapper()

	baseDemographic := map[string]dto.StudentAttributesDTO{
		"s1": {NumericAttrs: map[string]float64{"neuroticism": 0.1}},
		"s2": {NumericAttrs: map[string]float64{"neuroticism": 0.1}},
	}

	makeDTO := func(goal string) *dto.TeamFormationRequestDTO {
		return &dto.TeamFormationRequestDTO{
			SectionID:  "sec-1",
			Candidates: []string{"s1", "s2"},
			TeamSize:   2,
			CriteriaParameters: map[string]dto.CriterionParameterDTO{
				"neuroticism": {Target: 0.5, Weight: 1.0, Goal: goal},
			},
			DemographicData: baseDemographic,
		}
	}

	_, _, factoryMinimize, ctxMin, err := m.MapToDomain(makeDTO("minimize"))
	if err != nil {
		t.Fatalf("MapToDomain(minimize) error: %v", err)
	}
	_, _, factoryTarget, ctxTgt, err := m.MapToDomain(makeDTO("target"))
	if err != nil {
		t.Fatalf("MapToDomain(target) error: %v", err)
	}

	group := domain.Group{
		{ID: "s1", NumericAttrs: map[string]float64{"neuroticism": 0.1}},
		{ID: "s2", NumericAttrs: map[string]float64{"neuroticism": 0.1}},
	}

	scoreMinimize := factoryMinimize.Get("neuroticism").Calculate(&group, ctxMin)
	scoreTarget := factoryTarget.Get("neuroticism").Calculate(&group, ctxTgt)

	// minimize: mean=0.1 → score=0.9
	// target(0.5): mean=0.1, distance=0.4 → score=0.6
	// They must differ.
	if scoreMinimize == scoreTarget {
		t.Errorf("expected different scores for minimize vs target goals, both got %.4f", scoreMinimize)
	}
	if scoreMinimize < 0.85 {
		t.Errorf("minimize score for low-neuroticism group should be high, got %.4f", scoreMinimize)
	}
}

func TestResolveCriterionTypeWithExplicitType(t *testing.T) {
	m := NewTeamFormationMapper()

	// 1. Criterio con nombre desconocido pero Type explícito "statistical" o "numeric"
	cTypeNum, err := m.resolveCriterionType("custom_sensor_metric", dto.CriterionParameterDTO{Type: "statistical"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cTypeNum != CriterionTypeNumeric {
		t.Errorf("expected CriterionTypeNumeric, got %v", cTypeNum)
	}

	// 2. Criterio con nombre desconocido pero Type explícito "relational" o "social"
	cTypeSoc, err := m.resolveCriterionType("custom_peer_network", dto.CriterionParameterDTO{Type: "relational"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cTypeSoc != CriterionTypeSocial {
		t.Errorf("expected CriterionTypeSocial, got %v", cTypeSoc)
	}

	// 3. Criterio con nombre desconocido pero Type explícito "categorical"
	cTypeCat, err := m.resolveCriterionType("shift_preference", dto.CriterionParameterDTO{Type: "categorical"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cTypeCat != CriterionTypeCategorical {
		t.Errorf("expected CriterionTypeCategorical, got %v", cTypeCat)
	}

	// 4. MapToDomain con criterios explícitos
	reqDTO := &dto.TeamFormationRequestDTO{
		SectionID: "sec-1",
		TeamSize:  2,
		Candidates: []string{"s1", "s2"},
		Students: []dto.StudentDTO{
			{ID: "s1", Attributes: dto.StudentAttributesDTO{NumericAttrs: map[string]float64{"custom_metric": 0.5}}},
			{ID: "s2", Attributes: dto.StudentAttributesDTO{NumericAttrs: map[string]float64{"custom_metric": 0.5}}},
		},
		CriteriaParameters: map[string]dto.CriterionParameterDTO{
			"custom_metric": {Target: 0.5, Weight: 0.8, Type: "numeric"},
			"custom_peer":   {Target: 0.0, Weight: 0.5, Type: "social"},
		},
	}

	_, _, factory, _, err := m.MapToDomain(reqDTO)
	if err != nil {
		t.Fatalf("MapToDomain failed: %v", err)
	}

	if factory.Get("custom_metric") == nil {
		t.Errorf("factory should contain custom_metric")
	}
	if factory.Get("custom_peer") == nil {
		t.Errorf("factory should contain custom_peer")
	}
	if !m.IsSocialCriterion("custom_peer") {
		t.Errorf("custom_peer should be recognized as social criterion")
	}
}

