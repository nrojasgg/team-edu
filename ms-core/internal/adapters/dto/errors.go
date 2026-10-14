package dto

import "fmt"

// Validation errors
var (
	ErrMissingSectionID   = fmt.Errorf("section_id is required")
	ErrNoCandidates       = fmt.Errorf("candidates list cannot be empty")
	ErrInvalidTeamSize    = fmt.Errorf("team_size must be at least 2")
	ErrNoCriteria         = fmt.Errorf("criteria_parameters cannot be empty")
	ErrNoDemographicData  = fmt.Errorf("demographic_data cannot be empty")
	ErrInvalidCriterionID = fmt.Errorf("criterion ID cannot be empty")
)

// ErrInvalidCriterionWeight representa un error cuando el peso de un criterio está fuera de rango
type ErrInvalidCriterionWeight struct {
	CriterionName string
	Weight        float64
}

func (e ErrInvalidCriterionWeight) Error() string {
	return fmt.Sprintf("criterion '%s' has invalid weight %.2f (must be between 0.0 and 1.0)", e.CriterionName, e.Weight)
}

func NewErrInvalidCriterionWeight(name string, weight float64) error {
	return ErrInvalidCriterionWeight{
		CriterionName: name,
		Weight:        weight,
	}
}

// ErrInvalidCriterionTarget representa un error cuando el target de un criterio está fuera de rango
type ErrInvalidCriterionTarget struct {
	CriterionName string
	Target        float64
}

func (e ErrInvalidCriterionTarget) Error() string {
	return fmt.Sprintf("criterion '%s' has invalid target %.2f (must be between 0.0 and 1.0)", e.CriterionName, e.Target)
}

func NewErrInvalidCriterionTarget(name string, target float64) error {
	return ErrInvalidCriterionTarget{
		CriterionName: name,
		Target:        target,
	}
}

// ErrMissingStudentData representa un error cuando falta información de un estudiante
type ErrMissingStudentData struct {
	StudentID string
}

func (e ErrMissingStudentData) Error() string {
	return fmt.Sprintf("missing demographic data for student '%s'", e.StudentID)
}

func NewErrMissingStudentData(studentID string) error {
	return ErrMissingStudentData{StudentID: studentID}
}

// ErrUnknownCriterionType representa un error cuando el tipo de criterio no puede ser detectado
type ErrUnknownCriterionType struct {
	CriterionName string
}

func (e ErrUnknownCriterionType) Error() string {
	return fmt.Sprintf("cannot determine type for criterion '%s' (not numeric, categorical, or social)", e.CriterionName)
}

func NewErrUnknownCriterionType(name string) error {
	return ErrUnknownCriterionType{CriterionName: name}
}

// ErrInvalidCriterionGoal representa un error cuando el goal de un criterio no es válido
type ErrInvalidCriterionGoal struct {
	CriterionName string
	Goal          string
}

func (e ErrInvalidCriterionGoal) Error() string {
	return fmt.Sprintf("criterion '%s' has invalid goal '%s' (valid: minimize, maximize, target, heterogeneity, homogeneity)", e.CriterionName, e.Goal)
}

func NewErrInvalidCriterionGoal(name string, goal string) error {
	return ErrInvalidCriterionGoal{CriterionName: name, Goal: goal}
}
