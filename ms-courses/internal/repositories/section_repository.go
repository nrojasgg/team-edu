package repositories

import (
	"database/sql"
	"ms-courses/internal/models"
)

type SectionRepository struct {
	db *sql.DB
}

func NewSectionRepository(db *sql.DB) *SectionRepository {
	return &SectionRepository{db: db}
}

func (r *SectionRepository) populateStaffForSections(sections []models.Section) ([]models.Section, error) {
	if len(sections) == 0 {
		return sections, nil
	}

	for i := range sections {
		sections[i].ProfessorIDs = make([]int64, 0)
		sections[i].AssistantIDs = make([]int64, 0)
	}

	idxMap := make(map[int64]int, len(sections))
	for i, s := range sections {
		idxMap[s.ID] = i
	}

	profRows, err := r.db.Query("SELECT section_id, professor_id FROM section_professors")
	if err == nil {
		defer profRows.Close()
		for profRows.Next() {
			var secID, profID int64
			if err := profRows.Scan(&secID, &profID); err == nil {
				if idx, ok := idxMap[secID]; ok {
					sections[idx].ProfessorIDs = append(sections[idx].ProfessorIDs, profID)
				}
			}
		}
	}

	astRows, err := r.db.Query("SELECT section_id, assistant_id FROM section_assistants")
	if err == nil {
		defer astRows.Close()
		for astRows.Next() {
			var secID, astID int64
			if err := astRows.Scan(&secID, &astID); err == nil {
				if idx, ok := idxMap[secID]; ok {
					sections[idx].AssistantIDs = append(sections[idx].AssistantIDs, astID)
				}
			}
		}
	}

	// Fallback for legacy rows
	for i := range sections {
		if len(sections[i].ProfessorIDs) == 0 && sections[i].ProfessorID != nil && *sections[i].ProfessorID > 0 {
			sections[i].ProfessorIDs = append(sections[i].ProfessorIDs, *sections[i].ProfessorID)
		}
		if len(sections[i].AssistantIDs) == 0 && sections[i].AssistantID != nil && *sections[i].AssistantID > 0 {
			sections[i].AssistantIDs = append(sections[i].AssistantIDs, *sections[i].AssistantID)
		}
	}

	return sections, nil
}

func (r *SectionRepository) populateStaffForSection(section *models.Section) (*models.Section, error) {
	if section == nil {
		return nil, nil
	}
	section.ProfessorIDs = make([]int64, 0)
	section.AssistantIDs = make([]int64, 0)

	profRows, err := r.db.Query("SELECT professor_id FROM section_professors WHERE section_id = $1", section.ID)
	if err == nil {
		defer profRows.Close()
		for profRows.Next() {
			var profID int64
			if err := profRows.Scan(&profID); err == nil {
				section.ProfessorIDs = append(section.ProfessorIDs, profID)
			}
		}
	}

	astRows, err := r.db.Query("SELECT assistant_id FROM section_assistants WHERE section_id = $1", section.ID)
	if err == nil {
		defer astRows.Close()
		for astRows.Next() {
			var astID int64
			if err := astRows.Scan(&astID); err == nil {
				section.AssistantIDs = append(section.AssistantIDs, astID)
			}
		}
	}

	if len(section.ProfessorIDs) == 0 && section.ProfessorID != nil && *section.ProfessorID > 0 {
		section.ProfessorIDs = append(section.ProfessorIDs, *section.ProfessorID)
	}
	if len(section.AssistantIDs) == 0 && section.AssistantID != nil && *section.AssistantID > 0 {
		section.AssistantIDs = append(section.AssistantIDs, *section.AssistantID)
	}

	return section, nil
}

func (r *SectionRepository) GetAll() ([]models.Section, error) {
	query := `
		SELECT s.id, s.name, s.course_offering_id, s.professor_id, s.assistant_id
		FROM sections s
		ORDER BY s.name
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sections []models.Section
	for rows.Next() {
		var section models.Section
		err := rows.Scan(&section.ID, &section.Name, &section.CourseOfferingID, &section.ProfessorID, &section.AssistantID)
		if err != nil {
			return nil, err
		}
		sections = append(sections, section)
	}

	return r.populateStaffForSections(sections)
}

func (r *SectionRepository) GetByID(id int64) (*models.Section, error) {
	query := "SELECT id, name, course_offering_id, professor_id, assistant_id FROM sections WHERE id = $1"
	row := r.db.QueryRow(query, id)

	var section models.Section
	err := row.Scan(&section.ID, &section.Name, &section.CourseOfferingID, &section.ProfessorID, &section.AssistantID)
	if err != nil {
		return nil, err
	}

	return r.populateStaffForSection(&section)
}

func (r *SectionRepository) GetByCourseOffering(courseOfferingID int64) ([]models.Section, error) {
	query := "SELECT id, name, course_offering_id, professor_id, assistant_id FROM sections WHERE course_offering_id = $1"
	rows, err := r.db.Query(query, courseOfferingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sections []models.Section
	for rows.Next() {
		var section models.Section
		err := rows.Scan(&section.ID, &section.Name, &section.CourseOfferingID, &section.ProfessorID, &section.AssistantID)
		if err != nil {
			return nil, err
		}
		sections = append(sections, section)
	}

	return r.populateStaffForSections(sections)
}

func (r *SectionRepository) Create(section *models.Section) (*models.Section, error) {
	if len(section.ProfessorIDs) > 0 && section.ProfessorID == nil {
		p := section.ProfessorIDs[0]
		section.ProfessorID = &p
	} else if section.ProfessorID != nil && len(section.ProfessorIDs) == 0 {
		section.ProfessorIDs = []int64{*section.ProfessorID}
	}

	if len(section.AssistantIDs) > 0 && section.AssistantID == nil {
		a := section.AssistantIDs[0]
		section.AssistantID = &a
	} else if section.AssistantID != nil && len(section.AssistantIDs) == 0 {
		section.AssistantIDs = []int64{*section.AssistantID}
	}

	query := `
		INSERT INTO sections (name, course_offering_id, professor_id, assistant_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, course_offering_id, professor_id, assistant_id
	`
	err := r.db.QueryRow(query, section.Name, section.CourseOfferingID, section.ProfessorID, section.AssistantID).Scan(
		&section.ID, &section.Name, &section.CourseOfferingID, &section.ProfessorID, &section.AssistantID,
	)
	if err != nil {
		return nil, err
	}

	for _, pid := range section.ProfessorIDs {
		if pid > 0 {
			_, _ = r.db.Exec("INSERT INTO section_professors (section_id, professor_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", section.ID, pid)
		}
	}
	for _, aid := range section.AssistantIDs {
		if aid > 0 {
			_, _ = r.db.Exec("INSERT INTO section_assistants (section_id, assistant_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", section.ID, aid)
		}
	}

	if section.ProfessorIDs == nil {
		section.ProfessorIDs = make([]int64, 0)
	}
	if section.AssistantIDs == nil {
		section.AssistantIDs = make([]int64, 0)
	}

	return section, nil
}

func (r *SectionRepository) Update(section *models.Section) (*models.Section, error) {
	if len(section.ProfessorIDs) > 0 {
		p := section.ProfessorIDs[0]
		section.ProfessorID = &p
		_, _ = r.db.Exec("DELETE FROM section_professors WHERE section_id = $1", section.ID)
		for _, pid := range section.ProfessorIDs {
			if pid > 0 {
				_, _ = r.db.Exec("INSERT INTO section_professors (section_id, professor_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", section.ID, pid)
			}
		}
	}
	if len(section.AssistantIDs) > 0 {
		a := section.AssistantIDs[0]
		section.AssistantID = &a
		_, _ = r.db.Exec("DELETE FROM section_assistants WHERE section_id = $1", section.ID)
		for _, aid := range section.AssistantIDs {
			if aid > 0 {
				_, _ = r.db.Exec("INSERT INTO section_assistants (section_id, assistant_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", section.ID, aid)
			}
		}
	}

	query := `
		UPDATE sections
		SET name = $1, course_offering_id = $2, professor_id = $3, assistant_id = $4
		WHERE id = $5
		RETURNING id, name, course_offering_id, professor_id, assistant_id
	`
	err := r.db.QueryRow(query, section.Name, section.CourseOfferingID, section.ProfessorID, section.AssistantID, section.ID).Scan(
		&section.ID, &section.Name, &section.CourseOfferingID, &section.ProfessorID, &section.AssistantID,
	)
	if err != nil {
		return nil, err
	}

	return r.populateStaffForSection(section)
}

func (r *SectionRepository) Delete(id int64) error {
	query := "DELETE FROM sections WHERE id = $1"
	_, err := r.db.Exec(query, id)
	return err
}

// UpdateStaff updates professor and assistant assignments for a section (supporting multiple)
func (r *SectionRepository) UpdateStaff(sectionID int64, professorIDs, assistantIDs []int64) (*models.Section, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM section_professors WHERE section_id = $1", sectionID); err != nil {
		return nil, err
	}
	for _, pid := range professorIDs {
		if pid > 0 {
			if _, err := tx.Exec("INSERT INTO section_professors (section_id, professor_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", sectionID, pid); err != nil {
				return nil, err
			}
		}
	}

	if _, err := tx.Exec("DELETE FROM section_assistants WHERE section_id = $1", sectionID); err != nil {
		return nil, err
	}
	for _, aid := range assistantIDs {
		if aid > 0 {
			if _, err := tx.Exec("INSERT INTO section_assistants (section_id, assistant_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", sectionID, aid); err != nil {
				return nil, err
			}
		}
	}

	var primaryProfID, primaryAstID *int64
	if len(professorIDs) > 0 && professorIDs[0] > 0 {
		p := professorIDs[0]
		primaryProfID = &p
	}
	if len(assistantIDs) > 0 && assistantIDs[0] > 0 {
		a := assistantIDs[0]
		primaryAstID = &a
	}

	query := `
		UPDATE sections 
		SET professor_id = $1, assistant_id = $2
		WHERE id = $3
		RETURNING id, name, course_offering_id, professor_id, assistant_id
	`

	var section models.Section
	err = tx.QueryRow(query, primaryProfID, primaryAstID, sectionID).Scan(
		&section.ID,
		&section.Name,
		&section.CourseOfferingID,
		&section.ProfessorID,
		&section.AssistantID,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	section.ProfessorIDs = professorIDs
	if section.ProfessorIDs == nil {
		section.ProfessorIDs = make([]int64, 0)
	}
	section.AssistantIDs = assistantIDs
	if section.AssistantIDs == nil {
		section.AssistantIDs = make([]int64, 0)
	}

	return &section, nil
}
