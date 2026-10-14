package services

import (
	"ms-registro-curricular/internal/models"
	"ms-registro-curricular/internal/repositories"
)

type RoleService interface {
	Create(role *models.Role) (*models.Role, error)
	GetByID(id int64) (*models.Role, error)
	Update(role *models.Role) (*models.Role, error)
	Delete(id int64) error
	List() ([]*models.Role, error)
}

type roleService struct {
	roleRepo repositories.RoleRepository
}

func NewRoleService(roleRepo repositories.RoleRepository) RoleService {
	return &roleService{roleRepo: roleRepo}
}

func (s *roleService) Create(role *models.Role) (*models.Role, error) {
	err := s.roleRepo.Create(role)
	if err != nil {
		return nil, err
	}

	return s.roleRepo.GetByID(role.ID)
}

func (s *roleService) GetByID(id int64) (*models.Role, error) {
	return s.roleRepo.GetByID(id)
}

func (s *roleService) Update(role *models.Role) (*models.Role, error) {
	err := s.roleRepo.Update(role)
	if err != nil {
		return nil, err
	}

	return s.roleRepo.GetByID(role.ID)
}

func (s *roleService) Delete(id int64) error {
	return s.roleRepo.Delete(id)
}

func (s *roleService) List() ([]*models.Role, error) {
	return s.roleRepo.List()
}
