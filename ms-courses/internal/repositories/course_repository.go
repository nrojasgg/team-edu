package repositories

import (
	"database/sql"
	"ms-courses/internal/models"
)

type CourseRepository struct {
	db *sql.DB
}

func NewCourseRepository(db *sql.DB) *CourseRepository {
	return &CourseRepository{db: db}
}

func (r *CourseRepository) GetAll() ([]models.Course, error) {
	query := "SELECT id, code, name FROM courses ORDER BY code"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []models.Course
	for rows.Next() {
		var course models.Course
		err := rows.Scan(&course.ID, &course.Code, &course.Name)
		if err != nil {
			return nil, err
		}
		courses = append(courses, course)
	}

	return courses, nil
}

func (r *CourseRepository) GetByID(id int64) (*models.Course, error) {
	query := "SELECT id, code, name FROM courses WHERE id = $1"
	row := r.db.QueryRow(query, id)

	var course models.Course
	err := row.Scan(&course.ID, &course.Code, &course.Name)
	if err != nil {
		return nil, err
	}

	return &course, nil
}

func (r *CourseRepository) Create(course *models.Course) (*models.Course, error) {
	query := "INSERT INTO courses (code, name) VALUES ($1, $2) RETURNING id"
	err := r.db.QueryRow(query, course.Code, course.Name).Scan(&course.ID)
	if err != nil {
		return nil, err
	}

	return course, nil
}

func (r *CourseRepository) Update(course *models.Course) (*models.Course, error) {
	query := "UPDATE courses SET code = $1, name = $2 WHERE id = $3"
	_, err := r.db.Exec(query, course.Code, course.Name, course.ID)
	if err != nil {
		return nil, err
	}

	return course, nil
}

func (r *CourseRepository) Delete(id int64) error {
	// Start a transaction to delete course and all related data
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// First, delete student_sections for all sections related to this course
	_, err = tx.Exec(`
		DELETE FROM student_sections 
		WHERE section_id IN (
			SELECT s.id FROM sections s
			JOIN course_offerings co ON s.course_offering_id = co.id
			WHERE co.course_id = $1
		)
	`, id)
	if err != nil {
		return err
	}

	// Then, delete sections related to this course
	_, err = tx.Exec(`
		DELETE FROM sections 
		WHERE course_offering_id IN (
			SELECT id FROM course_offerings WHERE course_id = $1
		)
	`, id)
	if err != nil {
		return err
	}

	// Then, delete course_offerings
	_, err = tx.Exec("DELETE FROM course_offerings WHERE course_id = $1", id)
	if err != nil {
		return err
	}

	// Finally, delete the course itself
	_, err = tx.Exec("DELETE FROM courses WHERE id = $1", id)
	if err != nil {
		return err
	}

	// Commit the transaction
	return tx.Commit()
}
