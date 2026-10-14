package dto

import (
	"strings"
	"testing"
)

// ─── TeamFormationRequestDTO.Validate ────────────────────────────────────────

func TestValidate_Valid(t *testing.T) {
	d := &TeamFormationRequestDTO{
		SectionID:  "s1",
		Candidates: []string{"a", "b"},
		TeamSize:   2,
		CriteriaParameters: map[string]CriterionParameterDTO{
			"gender": {Target: 0.5, Weight: 0.8},
		},
		DemographicData: map[string]StudentAttributesDTO{
			"a": {CategoricalAttrs: map[string]string{"gender": "M"}},
		},
	}
	if err := d.Validate(); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func TestValidate_MissingSectionID(t *testing.T) {
	d := &TeamFormationRequestDTO{
		Candidates: []string{"a"},
		TeamSize:   2,
		CriteriaParameters: map[string]CriterionParameterDTO{
			"x": {Target: 0.5, Weight: 0.5},
		},
		DemographicData: map[string]StudentAttributesDTO{"a": {}},
	}
	if err := d.Validate(); err != ErrMissingSectionID {
		t.Fatalf("expected ErrMissingSectionID, got: %v", err)
	}
}

func TestValidate_NoCandidates(t *testing.T) {
	d := &TeamFormationRequestDTO{
		SectionID: "s1",
		TeamSize:  2,
		CriteriaParameters: map[string]CriterionParameterDTO{
			"x": {Target: 0.5, Weight: 0.5},
		},
		DemographicData: map[string]StudentAttributesDTO{"a": {}},
	}
	if err := d.Validate(); err != ErrNoCandidates {
		t.Fatalf("expected ErrNoCandidates, got: %v", err)
	}
}

func TestValidate_InvalidTeamSize(t *testing.T) {
	d := &TeamFormationRequestDTO{
		SectionID:  "s1",
		Candidates: []string{"a"},
		TeamSize:   1,
		CriteriaParameters: map[string]CriterionParameterDTO{
			"x": {Target: 0.5, Weight: 0.5},
		},
		DemographicData: map[string]StudentAttributesDTO{"a": {}},
	}
	if err := d.Validate(); err != ErrInvalidTeamSize {
		t.Fatalf("expected ErrInvalidTeamSize, got: %v", err)
	}
}

func TestValidate_NoCriteria(t *testing.T) {
	d := &TeamFormationRequestDTO{
		SectionID:       "s1",
		Candidates:      []string{"a"},
		TeamSize:        2,
		DemographicData: map[string]StudentAttributesDTO{"a": {}},
	}
	if err := d.Validate(); err != ErrNoCriteria {
		t.Fatalf("expected ErrNoCriteria, got: %v", err)
	}
}

func TestValidate_NoDemographicData(t *testing.T) {
	d := &TeamFormationRequestDTO{
		SectionID:  "s1",
		Candidates: []string{"a"},
		TeamSize:   2,
		CriteriaParameters: map[string]CriterionParameterDTO{
			"x": {Target: 0.5, Weight: 0.5},
		},
	}
	if err := d.Validate(); err != ErrNoDemographicData {
		t.Fatalf("expected ErrNoDemographicData, got: %v", err)
	}
}

func TestValidate_InvalidWeight(t *testing.T) {
	d := &TeamFormationRequestDTO{
		SectionID:  "s1",
		Candidates: []string{"a"},
		TeamSize:   2,
		CriteriaParameters: map[string]CriterionParameterDTO{
			"x": {Target: 0.5, Weight: 1.5}, // invalid
		},
		DemographicData: map[string]StudentAttributesDTO{"a": {}},
	}
	err := d.Validate()
	if err == nil {
		t.Fatal("expected error for invalid weight")
	}
	if !strings.Contains(err.Error(), "invalid weight") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_InvalidTarget(t *testing.T) {
	d := &TeamFormationRequestDTO{
		SectionID:  "s1",
		Candidates: []string{"a"},
		TeamSize:   2,
		CriteriaParameters: map[string]CriterionParameterDTO{
			"x": {Target: -0.1, Weight: 0.5}, // invalid target
		},
		DemographicData: map[string]StudentAttributesDTO{"a": {}},
	}
	err := d.Validate()
	if err == nil {
		t.Fatal("expected error for invalid target")
	}
	if !strings.Contains(err.Error(), "invalid target") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ─── GetAlgorithm ────────────────────────────────────────────────────────────

func TestGetAlgorithm_Default(t *testing.T) {
	d := &TeamFormationRequestDTO{}
	if got := d.GetAlgorithm(); got != "genetic" {
		t.Fatalf("expected 'genetic', got %q", got)
	}
}

func TestGetAlgorithm_Custom(t *testing.T) {
	d := &TeamFormationRequestDTO{Algorithm: "greedy"}
	if got := d.GetAlgorithm(); got != "greedy" {
		t.Fatalf("expected 'greedy', got %q", got)
	}
}

// ─── GetMinTeamSize ───────────────────────────────────────────────────────────

func TestGetMinTeamSize_Explicit(t *testing.T) {
	d := &TeamFormationRequestDTO{TeamSize: 5, MinTeamSize: 3}
	if got := d.GetMinTeamSize(); got != 3 {
		t.Fatalf("expected 3, got %d", got)
	}
}

func TestGetMinTeamSize_Default(t *testing.T) {
	d := &TeamFormationRequestDTO{TeamSize: 5}
	if got := d.GetMinTeamSize(); got != 4 {
		t.Fatalf("expected 4, got %d", got)
	}
}

// ─── GetMaxTeamSize ───────────────────────────────────────────────────────────

func TestGetMaxTeamSize_Explicit(t *testing.T) {
	d := &TeamFormationRequestDTO{TeamSize: 5, MaxTeamSize: 7}
	if got := d.GetMaxTeamSize(); got != 7 {
		t.Fatalf("expected 7, got %d", got)
	}
}

func TestGetMaxTeamSize_Default(t *testing.T) {
	d := &TeamFormationRequestDTO{TeamSize: 5}
	if got := d.GetMaxTeamSize(); got != 6 {
		t.Fatalf("expected 6, got %d", got)
	}
}

// ─── Error types ─────────────────────────────────────────────────────────────

func TestErrInvalidCriterionWeight_Error(t *testing.T) {
	e := ErrInvalidCriterionWeight{CriterionName: "gender", Weight: 1.5}
	msg := e.Error()
	if !strings.Contains(msg, "gender") || !strings.Contains(msg, "1.50") {
		t.Fatalf("unexpected error message: %q", msg)
	}
}

func TestNewErrInvalidCriterionWeight(t *testing.T) {
	err := NewErrInvalidCriterionWeight("age", 2.0)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "age") {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestErrInvalidCriterionTarget_Error(t *testing.T) {
	e := ErrInvalidCriterionTarget{CriterionName: "gpa", Target: -0.5}
	msg := e.Error()
	if !strings.Contains(msg, "gpa") || !strings.Contains(msg, "-0.50") {
		t.Fatalf("unexpected error message: %q", msg)
	}
}

func TestNewErrInvalidCriterionTarget(t *testing.T) {
	err := NewErrInvalidCriterionTarget("gpa", -0.1)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "gpa") {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestErrMissingStudentData_Error(t *testing.T) {
	e := ErrMissingStudentData{StudentID: "stu42"}
	msg := e.Error()
	if !strings.Contains(msg, "stu42") {
		t.Fatalf("unexpected error message: %q", msg)
	}
}

func TestNewErrMissingStudentData(t *testing.T) {
	err := NewErrMissingStudentData("stu99")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "stu99") {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestErrUnknownCriterionType_Error(t *testing.T) {
	e := ErrUnknownCriterionType{CriterionName: "mystery"}
	msg := e.Error()
	if !strings.Contains(msg, "mystery") {
		t.Fatalf("unexpected error message: %q", msg)
	}
}

func TestNewErrUnknownCriterionType(t *testing.T) {
	err := NewErrUnknownCriterionType("unknown_field")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unknown_field") {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestSentinelErrors(t *testing.T) {
	errors := []error{
		ErrMissingSectionID,
		ErrNoCandidates,
		ErrInvalidTeamSize,
		ErrNoCriteria,
		ErrNoDemographicData,
		ErrInvalidCriterionID,
	}
	for _, e := range errors {
		if e == nil {
			t.Fatal("expected non-nil sentinel error")
		}
		if e.Error() == "" {
			t.Fatal("expected non-empty error message")
		}
	}
}
