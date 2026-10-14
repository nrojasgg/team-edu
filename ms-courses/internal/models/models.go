package models

// Course representa un curso
type Course struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// Semester representa un semestre
type Semester struct {
	ID       int64 `json:"id"`
	Year     int   `json:"year"`
	Term     int   `json:"term"`
	IsActive bool  `json:"is_active"`
}

// CourseOffering representa una oferta de curso en un semestre específico
type CourseOffering struct {
	ID         int64     `json:"id"`
	CourseID   int64     `json:"course_id"`
	SemesterID int64     `json:"semester_id"`
	Course     *Course   `json:"course,omitempty"`
	Semester   *Semester `json:"semester,omitempty"`
}

// Section representa una sección de un curso
type Section struct {
	ID               int64           `json:"id"`
	Name             string          `json:"name"`
	CourseOfferingID int64           `json:"course_offering_id"`
	ProfessorID      *int64          `json:"professor_id,omitempty"`
	AssistantID      *int64          `json:"assistant_id,omitempty"`
	ProfessorIDs     []int64         `json:"professor_ids"`
	AssistantIDs     []int64         `json:"assistant_ids"`
	CourseOffering   *CourseOffering `json:"course_offering,omitempty"`
}

// StudentSection representa la relación entre un estudiante y una sección
type StudentSection struct {
	StudentID int64    `json:"student_id"`
	SectionID int64    `json:"section_id"`
	Section   *Section `json:"section,omitempty"`
}
