package repositories

import (
	"database/sql"
	"ms-courses/internal/models"
)

type StudentSectionRepository struct {
	db *sql.DB
}

func NewStudentSectionRepository(db *sql.DB) *StudentSectionRepository {
	return &StudentSectionRepository{db: db}
}

// GetStudentsBySectionID obtiene todos los estudiantes de una sección específica
func (r *StudentSectionRepository) GetStudentsBySectionID(sectionID int64) ([]models.StudentSection, error) {
	query := `
		SELECT student_id, section_id 
		FROM student_sections 
		WHERE section_id = $1
	`

	rows, err := r.db.Query(query, sectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var studentSections []models.StudentSection
	for rows.Next() {
		var ss models.StudentSection
		err := rows.Scan(&ss.StudentID, &ss.SectionID)
		if err != nil {
			return nil, err
		}
		studentSections = append(studentSections, ss)
	}

	return studentSections, nil
}

// GetSectionsByStudentID obtiene todas las secciones de un estudiante específico
func (r *StudentSectionRepository) GetSectionsByStudentID(studentID int64) ([]models.StudentSection, error) {
	query := `
		SELECT student_id, section_id 
		FROM student_sections 
		WHERE student_id = $1
	`

	rows, err := r.db.Query(query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var studentSections []models.StudentSection
	for rows.Next() {
		var ss models.StudentSection
		err := rows.Scan(&ss.StudentID, &ss.SectionID)
		if err != nil {
			return nil, err
		}
		studentSections = append(studentSections, ss)
	}

	return studentSections, nil
}

// AddStudentToSection agrega un estudiante a una sección
func (r *StudentSectionRepository) AddStudentToSection(studentID, sectionID int64) error {
	query := `
		INSERT INTO student_sections (student_id, section_id) 
		VALUES ($1, $2)
		ON CONFLICT (student_id, section_id) DO NOTHING
	`

	_, err := r.db.Exec(query, studentID, sectionID)
	return err
}

// RemoveStudentFromSection remueve un estudiante de una sección
func (r *StudentSectionRepository) RemoveStudentFromSection(studentID, sectionID int64) error {
	query := `
		DELETE FROM student_sections 
		WHERE student_id = $1 AND section_id = $2
	`

	_, err := r.db.Exec(query, studentID, sectionID)
	return err
}
