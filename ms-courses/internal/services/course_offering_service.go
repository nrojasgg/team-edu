package services

import (
	"ms-courses/internal/models"
	"ms-courses/internal/repositories"
)

type CourseOfferingService struct {
	repo         *repositories.CourseOfferingRepository
	courseRepo   *repositories.CourseRepository
	semesterRepo *repositories.SemesterRepository
}

func NewCourseOfferingService(repo *repositories.CourseOfferingRepository, courseRepo *repositories.CourseRepository, semesterRepo *repositories.SemesterRepository) *CourseOfferingService {
	return &CourseOfferingService{
		repo:         repo,
		courseRepo:   courseRepo,
		semesterRepo: semesterRepo,
	}
}

func (s *CourseOfferingService) GetAllCourseOfferings() ([]models.CourseOffering, error) {
	return s.repo.GetAll()
}

func (s *CourseOfferingService) GetCourseOfferingByID(id int64) (*models.CourseOffering, error) {
	return s.repo.GetByID(id)
}

func (s *CourseOfferingService) CreateCourseOffering(courseOffering *models.CourseOffering) (*models.CourseOffering, error) {
	// Verificar que el curso existe
	_, err := s.courseRepo.GetByID(courseOffering.CourseID)
	if err != nil {
		return nil, err
	}

	// Verificar que el semestre existe
	_, err = s.semesterRepo.GetByID(courseOffering.SemesterID)
	if err != nil {
		return nil, err
	}

	return s.repo.Create(courseOffering)
}

func (s *CourseOfferingService) DeleteCourseOffering(id int64) error {
	return s.repo.Delete(id)
}
