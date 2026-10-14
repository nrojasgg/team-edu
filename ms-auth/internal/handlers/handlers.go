package handlers

import (
	"ms-registro-curricular/internal/dto"
	"ms-registro-curricular/internal/models"
	"ms-registro-curricular/internal/services"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// Simplified handlers for other entities
type RoleHandler struct {
	roleService services.RoleService
	validator   *validator.Validate
}

func NewRoleHandler(roleService services.RoleService) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
		validator:   validator.New(),
	}
}

// GetRoles retrieves all roles
// @Summary Get all roles
// @Description Get all available roles
// @Tags roles
// @Produce json
// @Success 200 {array} models.Role
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /roles [get]
func (h *RoleHandler) GetRoles(c *fiber.Ctx) error {
	roles, err := h.roleService.List()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve roles",
			Error:   err.Error(),
		})
	}

	return c.JSON(roles)
}

// GetRole retrieves a role by ID
// @Summary Get role by ID
// @Description Get role by ID
// @Tags roles
// @Produce json
// @Param id path int true "Role ID"
// @Success 200 {object} dto.ApiResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /roles/{id} [get]
func (h *RoleHandler) GetRole(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid role ID",
			Error:   err.Error(),
		})
	}

	role, err := h.roleService.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Role not found",
			Error:   err.Error(),
		})
	}

	return c.JSON(dto.ApiResponse{
		Success: true,
		Message: "Role retrieved successfully",
		Data:    role,
	})
}

// CreateRole creates a new role
// @Summary Create role
// @Description Create a new role
// @Tags roles
// @Accept json
// @Produce json
// @Param role body models.Role true "Role data"
// @Success 201 {object} dto.ApiResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /roles [post]
func (h *RoleHandler) CreateRole(c *fiber.Ctx) error {
	var role models.Role
	if err := c.BodyParser(&role); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	createdRole, err := h.roleService.Create(&role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Failed to create role",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(dto.ApiResponse{
		Success: true,
		Message: "Role created successfully",
		Data:    createdRole,
	})
}

// UpdateRole updates a role
// @Summary Update role
// @Description Update an existing role
// @Tags roles
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Param role body models.Role true "Role data"
// @Success 200 {object} dto.ApiResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /roles/{id} [put]
func (h *RoleHandler) UpdateRole(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid role ID",
			Error:   err.Error(),
		})
	}

	var role models.Role
	if err := c.BodyParser(&role); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}
	role.ID = id

	updatedRole, err := h.roleService.Update(&role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Failed to update role",
			Error:   err.Error(),
		})
	}

	return c.JSON(dto.ApiResponse{
		Success: true,
		Message: "Role updated successfully",
		Data:    updatedRole,
	})
}

// DeleteRole deletes a role
// @Summary Delete role
// @Description Delete a role
// @Tags roles
// @Produce json
// @Param id path int true "Role ID"
// @Success 200 {object} dto.ApiResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /roles/{id} [delete]
func (h *RoleHandler) DeleteRole(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid role ID",
			Error:   err.Error(),
		})
	}

	err = h.roleService.Delete(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Failed to delete role",
			Error:   err.Error(),
		})
	}

	return c.JSON(dto.ApiResponse{
		Success: true,
		Message: "Role deleted successfully",
	})
}

type PermissionHandler struct {
	permissionService services.PermissionService
}

func NewPermissionHandler(permissionService services.PermissionService) *PermissionHandler {
	return &PermissionHandler{permissionService: permissionService}
}

// GetPermissions retrieves all permissions
// @Summary Get all permissions
// @Description Get all available permissions
// @Tags permissions
// @Produce json
// @Success 200 {array} models.Permission
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /permissions [get]
func (h *PermissionHandler) GetPermissions(c *fiber.Ctx) error {
	permissions, err := h.permissionService.List()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve permissions",
			Error:   err.Error(),
		})
	}

	return c.JSON(permissions)
}

// GetPermission retrieves a permission by ID
// @Summary Get permission by ID
// @Description Get permission by ID
// @Tags permissions
// @Produce json
// @Param id path int true "Permission ID"
// @Success 200 {object} dto.ApiResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /permissions/{id} [get]
func (h *PermissionHandler) GetPermission(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid permission ID",
			Error:   err.Error(),
		})
	}

	permission, err := h.permissionService.GetByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Permission not found",
			Error:   err.Error(),
		})
	}

	return c.JSON(dto.ApiResponse{
		Success: true,
		Message: "Permission retrieved successfully",
		Data:    permission,
	})
}

// CreatePermission creates a new permission
// @Summary Create permission
// @Description Create a new permission
// @Tags permissions
// @Accept json
// @Produce json
// @Param permission body models.Permission true "Permission data"
// @Success 201 {object} dto.ApiResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /permissions [post]
func (h *PermissionHandler) CreatePermission(c *fiber.Ctx) error {
	var permission models.Permission
	if err := c.BodyParser(&permission); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	createdPermission, err := h.permissionService.Create(&permission)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Failed to create permission",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(dto.ApiResponse{
		Success: true,
		Message: "Permission created successfully",
		Data:    createdPermission,
	})
}

// UpdatePermission updates a permission
// @Summary Update permission
// @Description Update an existing permission
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path int true "Permission ID"
// @Param permission body models.Permission true "Permission data"
// @Success 200 {object} dto.ApiResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /permissions/{id} [put]
func (h *PermissionHandler) UpdatePermission(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid permission ID",
			Error:   err.Error(),
		})
	}

	var permission models.Permission
	if err := c.BodyParser(&permission); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}
	permission.ID = id

	updatedPermission, err := h.permissionService.Update(&permission)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Failed to update permission",
			Error:   err.Error(),
		})
	}

	return c.JSON(dto.ApiResponse{
		Success: true,
		Message: "Permission updated successfully",
		Data:    updatedPermission,
	})
}

// DeletePermission deletes a permission
// @Summary Delete permission
// @Description Delete a permission
// @Tags permissions
// @Produce json
// @Param id path int true "Permission ID"
// @Success 200 {object} dto.ApiResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /permissions/{id} [delete]
func (h *PermissionHandler) DeletePermission(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid permission ID",
			Error:   err.Error(),
		})
	}

	err = h.permissionService.Delete(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Failed to delete permission",
			Error:   err.Error(),
		})
	}

	return c.JSON(dto.ApiResponse{
		Success: true,
		Message: "Permission deleted successfully",
	})
}

// GetRolePermissions retrieves permissions for a specific role
// @Summary Get role permissions
// @Description Get all permissions assigned to a role
// @Tags roles
// @Produce json
// @Param id path int true "Role ID"
// @Success 200 {array} models.Permission
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /roles/{id}/permissions [get]
func (h *PermissionHandler) GetRolePermissions(c *fiber.Ctx) error {
	roleID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid role ID",
			Error:   err.Error(),
		})
	}

	permissions, err := h.permissionService.GetByRoleID(roleID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Failed to retrieve role permissions",
			Error:   err.Error(),
		})
	}

	return c.JSON(permissions)
}

// AssignPermissionToRole assigns a permission to a role
// @Summary Assign permission to role
// @Description Assign a specific permission to a role
// @Tags roles
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Param permission body map[string]int64 true "Permission assignment data"
// @Success 200 {object} dto.ApiResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /roles/{id}/permissions [post]
func (h *PermissionHandler) AssignPermissionToRole(c *fiber.Ctx) error {
	roleID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid role ID",
			Error:   err.Error(),
		})
	}

	var req struct {
		PermissionID int64 `json:"permission_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	err = h.permissionService.AssignToRole(roleID, req.PermissionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Failed to assign permission to role",
			Error:   err.Error(),
		})
	}

	return c.JSON(dto.ApiResponse{
		Success: true,
		Message: "Permission assigned to role successfully",
	})
}

// RemovePermissionFromRole removes a permission from a role
// @Summary Remove permission from role
// @Description Remove a specific permission from a role
// @Tags roles
// @Produce json
// @Param roleId path int true "Role ID"
// @Param permissionId path int true "Permission ID"
// @Success 200 {object} dto.ApiResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Security BearerAuth
// @Router /roles/{roleId}/permissions/{permissionId} [delete]
func (h *PermissionHandler) RemovePermissionFromRole(c *fiber.Ctx) error {
	roleID, err := strconv.ParseInt(c.Params("roleId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid role ID",
			Error:   err.Error(),
		})
	}

	permissionID, err := strconv.ParseInt(c.Params("permissionId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid permission ID",
			Error:   err.Error(),
		})
	}

	err = h.permissionService.RemoveFromRole(roleID, permissionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Failed to remove permission from role",
			Error:   err.Error(),
		})
	}

	return c.JSON(dto.ApiResponse{
		Success: true,
		Message: "Permission removed from role successfully",
	})
}
