package repositories

import (
	"database/sql"
	"ms-courses/internal/models"
)

type CourseOfferingRepository struct {
	db *sql.DB
}

func NewCourseOfferingRepository(db *sql.DB) *CourseOfferingRepository {
	return &CourseOfferingRepository{db: db}
}

func (r *CourseOfferingRepository) GetAll() ([]models.CourseOffering, error) {
	query := `
		SELECT co.id, co.course_id, co.semester_id,
			   c.id, c.code, c.name,
			   s.id, s.year, s.term
		FROM course_offerings co
		JOIN courses c ON co.course_id = c.id
		JOIN semesters s ON co.semester_id = s.id
		ORDER BY s.year DESC, s.term DESC, c.code
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courseOfferings []models.CourseOffering
	for rows.Next() {
		var co models.CourseOffering
		var course models.Course
		var semester models.Semester

		err := rows.Scan(
			&co.ID, &co.CourseID, &co.SemesterID,
			&course.ID, &course.Code, &course.Name,
			&semester.ID, &semester.Year, &semester.Term,
		)
		if err != nil {
			return nil, err
		}

		co.Course = &course
		co.Semester = &semester
		courseOfferings = append(courseOfferings, co)
	}

	return courseOfferings, nil
}

func (r *CourseOfferingRepository) GetByID(id int64) (*models.CourseOffering, error) {
	query := `
		SELECT co.id, co.course_id, co.semester_id,
			   c.id, c.code, c.name,
			   s.id, s.year, s.term
		FROM course_offerings co
		JOIN courses c ON co.course_id = c.id
		JOIN semesters s ON co.semester_id = s.id
		WHERE co.id = $1
	`
	row := r.db.QueryRow(query, id)

	var co models.CourseOffering
	var course models.Course
	var semester models.Semester

	err := row.Scan(
		&co.ID, &co.CourseID, &co.SemesterID,
		&course.ID, &course.Code, &course.Name,
		&semester.ID, &semester.Year, &semester.Term,
	)
	if err != nil {
		return nil, err
	}

	co.Course = &course
	co.Semester = &semester

	return &co, nil
}

func (r *CourseOfferingRepository) Create(courseOffering *models.CourseOffering) (*models.CourseOffering, error) {
	query := "INSERT INTO course_offerings (course_id, semester_id) VALUES ($1, $2) RETURNING id"
	err := r.db.QueryRow(query, courseOffering.CourseID, courseOffering.SemesterID).Scan(&courseOffering.ID)
	if err != nil {
		return nil, err
	}

	return courseOffering, nil
}

func (r *CourseOfferingRepository) Delete(id int64) error {
	query := "DELETE FROM course_offerings WHERE id = $1"
	_, err := r.db.Exec(query, id)
	return err
}
