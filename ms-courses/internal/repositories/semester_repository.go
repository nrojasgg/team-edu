package repositories

import (
	"database/sql"
	"ms-courses/internal/models"
)

type SemesterRepository struct {
	db *sql.DB
}

func NewSemesterRepository(db *sql.DB) *SemesterRepository {
	return &SemesterRepository{db: db}
}

func (r *SemesterRepository) GetAll() ([]models.Semester, error) {
	query := "SELECT id, year, term, is_active FROM semesters ORDER BY year DESC, term DESC"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var semesters []models.Semester
	for rows.Next() {
		var semester models.Semester
		err := rows.Scan(&semester.ID, &semester.Year, &semester.Term, &semester.IsActive)
		if err != nil {
			return nil, err
		}
		semesters = append(semesters, semester)
	}

	return semesters, nil
}

func (r *SemesterRepository) GetByID(id int64) (*models.Semester, error) {
	query := "SELECT id, year, term, is_active FROM semesters WHERE id = $1"
	row := r.db.QueryRow(query, id)

	var semester models.Semester
	err := row.Scan(&semester.ID, &semester.Year, &semester.Term, &semester.IsActive)
	if err != nil {
		return nil, err
	}

	return &semester, nil
}

func (r *SemesterRepository) Create(semester *models.Semester) (*models.Semester, error) {
	query := "INSERT INTO semesters (year, term, is_active) VALUES ($1, $2, $3) RETURNING id"
	err := r.db.QueryRow(query, semester.Year, semester.Term, semester.IsActive).Scan(&semester.ID)
	if err != nil {
		return nil, err
	}

	return semester, nil
}

func (r *SemesterRepository) Update(semester *models.Semester) (*models.Semester, error) {
	query := "UPDATE semesters SET year = $1, term = $2, is_active = $3 WHERE id = $4"
	_, err := r.db.Exec(query, semester.Year, semester.Term, semester.IsActive, semester.ID)
	if err != nil {
		return nil, err
	}

	return semester, nil
}

func (r *SemesterRepository) Delete(id int64) error {
	query := "DELETE FROM semesters WHERE id = $1"
	_, err := r.db.Exec(query, id)
	return err
}

func (r *SemesterRepository) GetCurrent() (*models.Semester, error) {
	query := "SELECT id, year, term, is_active FROM semesters WHERE is_active = true LIMIT 1"
	row := r.db.QueryRow(query)

	var semester models.Semester
	err := row.Scan(&semester.ID, &semester.Year, &semester.Term, &semester.IsActive)
	if err != nil {
		return nil, err
	}

	return &semester, nil
}

func (r *SemesterRepository) GetByYearAndTerm(year int, term int) (*models.Semester, error) {
	query := "SELECT id, year, term, is_active FROM semesters WHERE year = $1 AND term = $2"
	row := r.db.QueryRow(query, year, term)

	var semester models.Semester
	err := row.Scan(&semester.ID, &semester.Year, &semester.Term, &semester.IsActive)
	if err != nil {
		return nil, err
	}

	return &semester, nil
}

func (r *SemesterRepository) SetActive(id int64) (*models.Semester, error) {
	// Primero desactivar todos los semestres
	_, err := r.db.Exec("UPDATE semesters SET is_active = false")
	if err != nil {
		return nil, err
	}

	// Luego activar el semestre específico
	query := "UPDATE semesters SET is_active = true WHERE id = $1"
	_, err = r.db.Exec(query, id)
	if err != nil {
		return nil, err
	}

	// Retornar el semestre activado
	return r.GetByID(id)
}

func (r *SemesterRepository) CreateAndActivate(semester *models.Semester) (*models.Semester, error) {
	// Comenzar una transacción
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Desactivar todos los semestres existentes
	_, err = tx.Exec("UPDATE semesters SET is_active = false")
	if err != nil {
		return nil, err
	}

	// Crear el nuevo semestre como activo
	semester.IsActive = true
	query := "INSERT INTO semesters (year, term, is_active) VALUES ($1, $2, $3) RETURNING id"
	err = tx.QueryRow(query, semester.Year, semester.Term, semester.IsActive).Scan(&semester.ID)
	if err != nil {
		return nil, err
	}

	// Confirmar la transacción
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return semester, nil
}
