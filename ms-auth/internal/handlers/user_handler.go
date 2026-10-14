package handlers

import (
	"fmt"
	"ms-registro-curricular/internal/dto"
	"ms-registro-curricular/internal/models"
	"ms-registro-curricular/internal/services"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// GetByID proxy para el servicio de usuario (para el middleware)
func (h *UserHandler) GetByID(id int64) (*models.User, error) {
	return h.userService.GetByID(id)
}

// ImpersonateUser allows an admin to simulate another user and returns a JWT for that user
// @Summary Impersonate user
// @Description Admin can simulate another user and get a JWT for that user
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} dto.ApiResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /users/{id}/impersonate [post]
func (h *UserHandler) ImpersonateUser(c *fiber.Ctx) error {
	// Obtener el usuario actual desde el contexto (middleware debe ponerlo)
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok || currentUser == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Unauthorized",
		})
	}
	// Solo admin puede simular
	if currentUser.Role == nil || currentUser.Role.Name != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Only admin can impersonate users",
		})
	}
	// Obtener el usuario objetivo
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid user ID",
			Error:   err.Error(),
		})
	}
	user, err := h.userService.GetByID(id)
	if err != nil || user == nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Success: false,
			Message: "User not found",
		})
	}
	// Generar JWT para el usuario objetivo
	token, err := h.userService.GenerateJWT(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Failed to generate token",
			Error:   err.Error(),
		})
	}
	return c.JSON(dto.ApiResponse{
		Success: true,
		Message: "Impersonation successful",
		Data: fiber.Map{
			"token": token,
			"user":  h.userToResponse(user),
		},
	})
}

// BatchUserRequest representa la petición para obtener usuarios por IDs
type BatchUserRequest struct {
	IDs []int64 `json:"ids"`
}

// BatchUserResponse representa la respuesta con los usuarios encontrados
type BatchUserResponse struct {
	Success bool               `json:"success"`
	Users   []dto.UserResponse `json:"users"`
}

// GetUsersBatch obtiene usuarios por una lista de IDs
func (h *UserHandler) GetUsersBatch(c *fiber.Ctx) error {
	// Esperar {ids: [...]} en el body
	var raw map[string]interface{}
	if err := c.BodyParser(&raw); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}
	idsRaw, ok := raw["ids"]
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Missing 'ids' field",
		})
	}
	idsIface, ok := idsRaw.([]interface{})
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "'ids' must be an array",
		})
	}
	ids := make([]int64, 0, len(idsIface))
	for _, v := range idsIface {
		switch val := v.(type) {
		case float64:
			ids = append(ids, int64(val))
		case int:
			ids = append(ids, int64(val))
		}
	}

	// Buscar usuarios por los IDs
	users := make([]dto.UserResponse, 0, len(ids))
	for _, id := range ids {
		user, err := h.userService.GetByID(id)
		if err == nil && user != nil {
			users = append(users, h.userToResponse(user))
		}
	}
	return c.JSON(BatchUserResponse{
		Success: true,
		Users:   users,
	})
}

type UserHandler struct {
	userService services.UserService
	roleService services.RoleService
	validator   *validator.Validate
}

func NewUserHandler(userService services.UserService, roleService services.RoleService) *UserHandler {
	return &UserHandler{
		userService: userService,
		roleService: roleService,
		validator:   validator.New(),
	}
}

// checkUserPermission verifica que el usuario autenticado tenga el permiso
// requerido para operar sobre el rol destino (ej: "Crear estudiante").
func (h *UserHandler) checkUserPermission(c *fiber.Ctx, roleID int64, action string) error {
	role, err := h.roleService.GetByID(roleID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid role ID",
			Error:   err.Error(),
		})
	}

	requiredPerm := fmt.Sprintf("%s %s", action, role.Name)

	perms, _ := c.Locals("userPermissions").([]string)
	for _, p := range perms {
		if p == requiredPerm {
			return nil
		}
	}

	return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
		Success: false,
		Message: fmt.Sprintf("Access denied: requires permission '%s'", requiredPerm),
	})
}

// Helper function to convert User model to UserResponse DTO
func (h *UserHandler) userToResponse(user *models.User) dto.UserResponse {
	fullName := user.Nombres
	if user.Paterno != "" {
		fullName += " " + user.Paterno
	}
	if user.Materno != "" {
		fullName += " " + user.Materno
	}

	response := dto.UserResponse{
		ID:         user.ID,
		Nombres:    user.Nombres,
		Paterno:    user.Paterno,
		Materno:    user.Materno,
		FullName:   fullName, // Campo calculado para compatibilidad
		RUT:        user.RUT,
		Email:      user.Email,
		PictureURL: user.PictureURL,
		RoleID:     user.RoleID,
		CareerCode: user.CareerCode,
	}

	if user.Role != nil {
		response.RoleName = user.Role.Name
	}

	return response
}

// CreateUser creates a new user
// @Summary Create user
// @Description Create a new user
// @Tags users
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "User data"
// @Success 201 {object} dto.ApiResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /users [post]
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req dto.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	if err := h.validator.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
	}

	// Admin can create any user without granular permission checks.
	callerIsAdmin := false
	if caller, ok := c.Locals("user").(*models.User); ok && caller != nil && caller.Role != nil {
		callerIsAdmin = caller.Role.Name == "admin"
	}
	if !callerIsAdmin {
		if permErr := h.checkUserPermission(c, req.RoleID, "Crear"); permErr != nil {
			return permErr
		}
	}

	user := &models.User{
		Nombres:    req.Nombres,
		Paterno:    req.Paterno,
		Materno:    req.Materno,
		RUT:        req.RUT,
		Email:      req.Email,
		Password:   req.Password,
		RoleID:     req.RoleID,
		CareerCode: req.CareerCode,
	}

	createdUser, err := h.userService.Create(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Failed to create user",
			Error:   err.Error(),
		})
	}

	response := h.userToResponse(createdUser)

	return c.Status(fiber.StatusCreated).JSON(dto.ApiResponse{
		Success: true,
		Message: "User created successfully",
		Data:    response,
	})
}

// GetUser retrieves a user by ID
// @Summary Get user
// @Description Get user by ID
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} dto.ApiResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid user ID",
			Error:   err.Error(),
		})
	}

	user, err := h.userService.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Success: false,
			Message: "User not found",
			Error:   err.Error(),
		})
	}

	response := h.userToResponse(user)

	return c.JSON(dto.ApiResponse{
		Success: true,
		Message: "User retrieved successfully",
		Data:    response,
	})
}

// GetUserByRUT retrieves a user by RUT
// @Summary Get user by RUT
// @Description Get user by RUT
// @Tags users
// @Produce json
// @Param rut query string true "User RUT"
// @Success 200 {object} dto.ApiResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /users/by-rut [get]
func (h *UserHandler) GetUserByRUT(c *fiber.Ctx) error {
	rut := c.Query("rut")
	if rut == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "RUT parameter is required",
			Error:   "rut query parameter cannot be empty",
		})
	}

	user, err := h.userService.GetByRUT(rut)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Success: false,
			Message: "User not found",
			Error:   err.Error(),
		})
	}

	response := h.userToResponse(user)

	return c.JSON(dto.ApiResponse{
		Success: true,
		Message: "User retrieved successfully",
		Data:    response,
	})
}

// GetUserByEmail retrieves a user by email
// @Summary Get user by email
// @Description Get user by email address
// @Tags users
// @Produce json
// @Param email query string true "User email"
// @Success 200 {object} dto.ApiResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /users/by-email [get]
func (h *UserHandler) GetUserByEmail(c *fiber.Ctx) error {
	email := c.Query("email")
	if email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Email parameter is required",
			Error:   "email query parameter cannot be empty",
		})
	}

	user, err := h.userService.GetByEmail(email)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Success: false,
			Message: "User not found",
			Error:   err.Error(),
		})
	}

	response := h.userToResponse(user)

	return c.JSON(dto.ApiResponse{
		Success: true,
		Message: "User retrieved successfully",
		Data:    response,
	})
}

// UpdateUser updates a user
// @Summary Update user
// @Description Update user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body dto.UpdateUserRequest true "User data"
// @Success 200 {object} dto.ApiResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid user ID",
			Error:   err.Error(),
		})
	}

	var req dto.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	if err := h.validator.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
	}

	// Admin can edit any user without granular permission checks.
	// Only apply the role-based permission check for non-admin callers.
	callerIsAdmin := false
	if caller, ok := c.Locals("user").(*models.User); ok && caller != nil && caller.Role != nil {
		callerIsAdmin = caller.Role.Name == "admin"
	}
	if !callerIsAdmin {
		if permErr := h.checkUserPermission(c, req.RoleID, "Editar"); permErr != nil {
			return permErr
		}
	}

	user := &models.User{
		ID:         id,
		Nombres:    req.Nombres,
		Paterno:    req.Paterno,
		Materno:    req.Materno,
		RUT:        req.RUT,
		Email:      req.Email,
		RoleID:     req.RoleID,
		CareerCode: req.CareerCode,
	}

	updatedUser, err := h.userService.Update(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Failed to update user",
			Error:   err.Error(),
		})
	}

	response := h.userToResponse(updatedUser)

	return c.JSON(dto.ApiResponse{
		Success: true,
		Message: "User updated successfully",
		Data:    response,
	})
}

// DeleteUser deletes a user
// @Summary Delete user
// @Description Delete user by ID
// @Tags users
// @Param id path int true "User ID"
// @Success 200 {object} dto.ApiResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid user ID",
			Error:   err.Error(),
		})
	}

	err = h.userService.Delete(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Failed to delete user",
			Error:   err.Error(),
		})
	}

	return c.JSON(dto.ApiResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}

// ListUsers lists all users with pagination
// @Summary List users
// @Description Get all users with pagination
// @Tags users
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} dto.ApiResponse
// @Failure 400 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /users [get]
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	offset := c.QueryInt("offset", 0)
	query := c.Query("q", "")
	roleID := c.QueryInt("role_id", 0)

	var users []*models.User
	var err error

	if query != "" || roleID > 0 {
		users, err = h.userService.Search(query, roleID, limit, offset)
	} else {
		users, err = h.userService.List(limit, offset)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Failed to get users",
			Error:   err.Error(),
		})
	}

	var responses []dto.UserResponse
	for _, user := range users {
		response := h.userToResponse(user)
		responses = append(responses, response)
	}

	return c.JSON(dto.ApiResponse{
		Success: true,
		Message: "Users retrieved successfully",
		Data:    responses,
	})
}

// GetUserPermissions retrieves permissions for a specific user
// @Summary Get user permissions
// @Description Get all permissions assigned to a user through their role
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {array} models.Permission
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /users/{id}/permissions [get]
func (h *UserHandler) GetUserPermissions(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid user ID",
			Error:   err.Error(),
		})
	}

	permissions, err := h.userService.GetUserPermissions(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve user permissions",
			Error:   err.Error(),
		})
	}

	return c.JSON(permissions)
}
