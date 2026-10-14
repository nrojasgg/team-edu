package services

import (
	"ms-courses/internal/models"
	"ms-courses/internal/repositories"
)

type CourseService struct {
	repo *repositories.CourseRepository
}

func NewCourseService(repo *repositories.CourseRepository) *CourseService {
	return &CourseService{repo: repo}
}

func (s *CourseService) GetAllCourses() ([]models.Course, error) {
	return s.repo.GetAll()
}

func (s *CourseService) GetCourseByID(id int64) (*models.Course, error) {
	return s.repo.GetByID(id)
}

func (s *CourseService) CreateCourse(course *models.Course) (*models.Course, error) {
	return s.repo.Create(course)
}

func (s *CourseService) UpdateCourse(course *models.Course) (*models.Course, error) {
	return s.repo.Update(course)
}

func (s *CourseService) DeleteCourse(id int64) error {
	return s.repo.Delete(id)
}
