package services

import (
	"ms-registro-curricular/internal/models"
	"ms-registro-curricular/internal/repositories"
)

type PermissionService interface {
	Create(permission *models.Permission) (*models.Permission, error)
	GetByID(id int64) (*models.Permission, error)
	Update(permission *models.Permission) (*models.Permission, error)
	Delete(id int64) error
	List() ([]*models.Permission, error)
	GetByRoleID(roleID int64) ([]*models.Permission, error)
	AssignToRole(roleID, permissionID int64) error
	RemoveFromRole(roleID, permissionID int64) error
}

type permissionService struct {
	permissionRepo repositories.PermissionRepository
}

func NewPermissionService(permissionRepo repositories.PermissionRepository) PermissionService {
	return &permissionService{permissionRepo: permissionRepo}
}

func (s *permissionService) Create(permission *models.Permission) (*models.Permission, error) {
	err := s.permissionRepo.Create(permission)
	if err != nil {
		return nil, err
	}

	return s.permissionRepo.GetByID(permission.ID)
}

func (s *permissionService) GetByID(id int64) (*models.Permission, error) {
	return s.permissionRepo.GetByID(id)
}

func (s *permissionService) Update(permission *models.Permission) (*models.Permission, error) {
	err := s.permissionRepo.Update(permission)
	if err != nil {
		return nil, err
	}

	return s.permissionRepo.GetByID(permission.ID)
}

func (s *permissionService) Delete(id int64) error {
	return s.permissionRepo.Delete(id)
}

func (s *permissionService) List() ([]*models.Permission, error) {
	return s.permissionRepo.List()
}

func (s *permissionService) GetByRoleID(roleID int64) ([]*models.Permission, error) {
	return s.permissionRepo.GetByRoleID(roleID)
}

func (s *permissionService) AssignToRole(roleID, permissionID int64) error {
	return s.permissionRepo.AssignToRole(roleID, permissionID)
}

func (s *permissionService) RemoveFromRole(roleID, permissionID int64) error {
	return s.permissionRepo.RemoveFromRole(roleID, permissionID)
}
