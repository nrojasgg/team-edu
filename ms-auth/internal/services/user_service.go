package services

import (
	"errors"
	"ms-registro-curricular/internal/models"
	"ms-registro-curricular/internal/repositories"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Create(user *models.User) (*models.User, error)
	GetByID(id int64) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetByRUT(rut string) (*models.User, error)
	Update(user *models.User) (*models.User, error)
	Delete(id int64) error
	List(limit, offset int) ([]*models.User, error)
	GenerateJWT(user *models.User) (string, error)
	GetUserPermissions(userID int64) ([]*models.Permission, error)
	HasPermission(userID int64, permissionName string) (bool, error)
	GetUsersBatch(ids []int64) ([]*models.User, error)
	Search(query string, roleID int, limit, offset int) ([]*models.User, error)
}

type userService struct {
	userRepo  repositories.UserRepository
	jwtSecret string
}

// NewUserService creates a new UserService and accepts the JWT secret used for signing tokens
func NewUserService(userRepo repositories.UserRepository, jwtSecret string) UserService {
	return &userService{userRepo: userRepo, jwtSecret: jwtSecret}
}

// GenerateJWT generates a JWT token for the given user
func (s *userService) GenerateJWT(user *models.User) (string, error) {
	// Get user permissions
	permissions, err := s.GetUserPermissions(user.ID)
	if err != nil {
		return "", err
	}

	// Extract permission names
	var permissionNames []string
	for _, perm := range permissions {
		permissionNames = append(permissionNames, perm.Name)
	}

	claims := jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"role": func() string {
			if user.Role != nil {
				return user.Role.Name
			} else {
				return ""
			}
		}(),
		"permissions": permissionNames,
		"exp":         time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *userService) Create(user *models.User) (*models.User, error) {
	// Check if user with this email already exists
	existingUser, _ := s.userRepo.GetByEmail(user.Email)
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Check if user with this RUT already exists
	existingUser, _ = s.userRepo.GetByRUT(user.RUT)
	if existingUser != nil {
		return nil, errors.New("user with this RUT already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}
	user.Password = string(hashedPassword)

	// Create user
	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	createdUser, err := s.userRepo.GetByID(user.ID)
	if err != nil {
		return nil, err
	}

	createdUser.Password = ""
	return createdUser, nil
}

func (s *userService) GetByID(id int64) (*models.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	user.Password = ""
	return user, nil
}

func (s *userService) GetByEmail(email string) (*models.User, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	user.Password = ""
	return user, nil
}

func (s *userService) GetByRUT(rut string) (*models.User, error) {
	user, err := s.userRepo.GetByRUT(rut)
	if err != nil {
		return nil, err
	}

	user.Password = ""
	return user, nil
}

func (s *userService) Update(user *models.User) (*models.User, error) {
	err := s.userRepo.Update(user)
	if err != nil {
		return nil, err
	}

	updatedUser, err := s.userRepo.GetByID(user.ID)
	if err != nil {
		return nil, err
	}

	updatedUser.Password = ""
	return updatedUser, nil
}

func (s *userService) Delete(id int64) error {
	return s.userRepo.Delete(id)
}

func (s *userService) List(limit, offset int) ([]*models.User, error) {
	users, err := s.userRepo.List(limit, offset)
	if err != nil {
		return nil, err
	}

	// Clear passwords
	for _, user := range users {
		user.Password = ""
	}

	return users, nil
}

func (s *userService) GetUserPermissions(userID int64) ([]*models.Permission, error) {
	return s.userRepo.GetUserPermissions(userID)
}

func (s *userService) HasPermission(userID int64, permissionName string) (bool, error) {
	permissions, err := s.userRepo.GetUserPermissions(userID)
	if err != nil {
		return false, err
	}

	for _, perm := range permissions {
		if perm.Name == permissionName {
			return true, nil
		}
	}

	return false, nil
}

func (s *userService) GetUsersBatch(ids []int64) ([]*models.User, error) {
	users, err := s.userRepo.GetUsersBatch(ids)
	if err != nil {
		return nil, err
	}

	// Clear passwords
	for _, user := range users {
		user.Password = ""
	}

	return users, nil
}

func (s *userService) Search(query string, roleID int, limit, offset int) ([]*models.User, error) {
	users, err := s.userRepo.Search(query, roleID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Clear passwords
	for _, user := range users {
		user.Password = ""
	}

	return users, nil
}
