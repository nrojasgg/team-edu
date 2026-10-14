package services

import (
	"ms-courses/internal/models"
	"ms-courses/internal/repositories"
)

type SectionService struct {
	repo               *repositories.SectionRepository
	courseOfferingRepo *repositories.CourseOfferingRepository
}

func NewSectionService(repo *repositories.SectionRepository, courseOfferingRepo *repositories.CourseOfferingRepository) *SectionService {
	return &SectionService{
		repo:               repo,
		courseOfferingRepo: courseOfferingRepo,
	}
}

func (s *SectionService) GetAllSections() ([]models.Section, error) {
	return s.repo.GetAll()
}

func (s *SectionService) GetSectionByID(id int64) (*models.Section, error) {
	return s.repo.GetByID(id)
}

func (s *SectionService) GetSectionsByCourseOffering(courseOfferingID int64) ([]models.Section, error) {
	return s.repo.GetByCourseOffering(courseOfferingID)
}

func (s *SectionService) CreateSection(section *models.Section) (*models.Section, error) {
	// Verificar que la oferta de curso existe
	_, err := s.courseOfferingRepo.GetByID(section.CourseOfferingID)
	if err != nil {
		return nil, err
	}

	return s.repo.Create(section)
}

func (s *SectionService) UpdateSection(section *models.Section) (*models.Section, error) {
	// Verificar que la oferta de curso existe
	_, err := s.courseOfferingRepo.GetByID(section.CourseOfferingID)
	if err != nil {
		return nil, err
	}

	return s.repo.Update(section)
}

func (s *SectionService) DeleteSection(id int64) error {
	return s.repo.Delete(id)
}

// UpdateStaff updates the professor and assistant assignments for a section (multiple supported)
func (s *SectionService) UpdateStaff(sectionID int64, professorIDs, assistantIDs []int64) (*models.Section, error) {
	return s.repo.UpdateStaff(sectionID, professorIDs, assistantIDs)
}
