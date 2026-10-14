package services

import (
	"ms-courses/internal/models"
	"ms-courses/internal/repositories"
)

type SemesterService struct {
	repo *repositories.SemesterRepository
}

func NewSemesterService(repo *repositories.SemesterRepository) *SemesterService {
	return &SemesterService{repo: repo}
}

func (s *SemesterService) GetAllSemesters() ([]models.Semester, error) {
	return s.repo.GetAll()
}

func (s *SemesterService) GetSemesterByID(id int64) (*models.Semester, error) {
	return s.repo.GetByID(id)
}

func (s *SemesterService) CreateSemester(semester *models.Semester) (*models.Semester, error) {
	return s.repo.Create(semester)
}

func (s *SemesterService) UpdateSemester(semester *models.Semester) (*models.Semester, error) {
	return s.repo.Update(semester)
}

func (s *SemesterService) DeleteSemester(id int64) error {
	return s.repo.Delete(id)
}

func (s *SemesterService) GetCurrentSemester() (*models.Semester, error) {
	return s.repo.GetCurrent()
}

func (s *SemesterService) SetActiveSemester(id int64) (*models.Semester, error) {
	return s.repo.SetActive(id)
}

func (s *SemesterService) CreateOrActivateSemester(year int, term int) (*models.Semester, error) {
	// Primero verificar si el semestre ya existe
	semester, err := s.repo.GetByYearAndTerm(year, term)
	if err == nil && semester != nil {
		// El semestre existe, solo activarlo
		return s.repo.SetActive(semester.ID)
	}

	// El semestre no existe, crearlo y activarlo
	newSemester := &models.Semester{
		Year:     year,
		Term:     term,
		IsActive: true,
	}

	// Crear el semestre (esto automáticamente lo activará)
	return s.repo.CreateAndActivate(newSemester)
}
