package services

import (
	"ms-courses/internal/models"
	"ms-courses/internal/repositories"
)

type StudentSectionService struct {
	repo *repositories.StudentSectionRepository
}

func NewStudentSectionService(repo *repositories.StudentSectionRepository) *StudentSectionService {
	return &StudentSectionService{repo: repo}
}

// GetStudentsBySectionID obtiene todos los estudiantes de una sección
func (s *StudentSectionService) GetStudentsBySectionID(sectionID int64) ([]models.StudentSection, error) {
	return s.repo.GetStudentsBySectionID(sectionID)
}

// GetSectionsByStudentID obtiene todas las secciones de un estudiante
func (s *StudentSectionService) GetSectionsByStudentID(studentID int64) ([]models.StudentSection, error) {
	return s.repo.GetSectionsByStudentID(studentID)
}

// AddStudentToSection agrega un estudiante a una sección
func (s *StudentSectionService) AddStudentToSection(studentID, sectionID int64) error {
	return s.repo.AddStudentToSection(studentID, sectionID)
}

// RemoveStudentFromSection remueve un estudiante de una sección
func (s *StudentSectionService) RemoveStudentFromSection(studentID, sectionID int64) error {
	return s.repo.RemoveStudentFromSection(studentID, sectionID)
}
