package handlers

import (
	"ms-registro-curricular/internal/dto"
	"ms-registro-curricular/internal/models"
	"ms-registro-curricular/internal/services"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService services.AuthService
	validator   *validator.Validate
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator.New(),
	}
}

// Helper function to convert User model to UserResponse DTO
func (h *AuthHandler) userToResponse(user *models.User) dto.UserResponse {
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

// Login authenticates a user
// @Summary User login
// @Description Authenticate user with email and password (DEPRECATED: Use Google OAuth instead)
// @Tags auth
// @Accept json
// @Produce json
// @Param loginRequest body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.ApiResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Email and password are required",
			Error:   "missing_credentials",
		})
	}

	token, user, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Correo o contraseña incorrectos",
			Error:   "invalid_credentials",
		})
	}

	return c.JSON(dto.ApiResponse{
		Success: true,
		Message: "Login successful",
		Data: fiber.Map{
			"token": token,
			"user":  user,
		},
	})
}

// Register creates a new user account
// @Summary User registration
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param registerRequest body dto.RegisterRequest true "User registration data"
// @Success 201 {object} dto.ApiResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Email and password are required",
			Error:   "missing_required_fields",
		})
	}

	user, err := h.authService.Register(&models.User{
		Nombres:    getStringValue(req.Nombres),
		Paterno:    getStringValue(req.Paterno),
		Materno:    getStringValue(req.Materno),
		RUT:        getStringValue(req.RUT),
		Email:      req.Email,
		Password:   req.Password,
		RoleID:     getInt64Value(req.RoleID),
		CareerCode: req.CareerCode,
	})

	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return c.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{
				Success: false,
				Message: "User with this email already exists",
				Error:   "user_exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Failed to create user",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(dto.ApiResponse{
		Success: true,
		Message: "User registered successfully",
		Data:    user,
	})
}

// GoogleAuth initiates Google OAuth flow
// @Summary Google OAuth URL
// @Description Get Google OAuth authorization URL
// @Tags auth
// @Produce json
// @Success 200 {object} dto.ApiResponse
// @Router /auth/google [get]
func (h *AuthHandler) GoogleAuth(c *fiber.Ctx) error {
	url := h.authService.GetGoogleAuthURL()

	return c.JSON(dto.ApiResponse{
		Success: true,
		Message: "Google OAuth URL generated",
		Data: fiber.Map{
			"auth_url": url,
		},
	})
}

// GoogleCallback handles Google OAuth callback
// @Summary Google OAuth callback
// @Description Handle Google OAuth callback and authenticate/register user
// @Tags auth
// @Accept json
// @Produce json
// @Param googleAuthRequest body dto.GoogleAuthRequest true "Google OAuth code"
// @Success 200 {object} dto.ApiResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/google/callback [post]
func (h *AuthHandler) GoogleCallback(c *fiber.Ctx) error {
	var req dto.GoogleAuthRequest
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

	token, user, err := h.authService.GoogleLogin(req.Code)
	if err != nil {
		// Check if the error is related to email domain validation
		if err.Error() == "only @usach.cl email addresses are allowed" {
			return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
				Success: false,
				Message: "Solo se permiten correos electrónicos con dominio @usach.cl",
				Error:   "invalid_email_domain",
			})
		}

		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Google authentication failed",
			Error:   err.Error(),
		})
	}

	response := dto.LoginResponse{
		Token: token,
		User:  h.userToResponse(user),
	}

	return c.JSON(dto.ApiResponse{
		Success: true,
		Message: "Autenticación con Google exitosa",
		Data:    response,
	})
}

// Helper functions to handle pointer values
func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func getInt64Value(ptr *int64) int64 {
	if ptr == nil {
		return 3 // Default role ID for students
	}
	return *ptr
}

// ForgotPassword handles the request to send a password reset email
// @Summary Forgot password
// @Description Send instructions to reset password
// @Tags auth
// @Accept json
// @Produce json
// @Param forgotRequest body dto.ForgotPasswordRequest true "Email address and reset URL"
// @Success 200 {object} dto.ApiResponse
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req dto.ForgotPasswordRequest
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

	err := h.authService.ForgotPassword(req.Email, req.ResetURL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Failed to process forgot password request",
			Error:   err.Error(),
		})
	}

	return c.JSON(dto.ApiResponse{
		Success: true,
		Message: "Si el correo electrónico está registrado, recibirás instrucciones en breve.",
	})
}

// ResetPassword handles the request to change password via reset token
// @Summary Reset password
// @Description Reset password using a valid token
// @Tags auth
// @Accept json
// @Produce json
// @Param resetRequest body dto.ResetPasswordRequest true "Token and new password"
// @Success 200 {object} dto.ApiResponse
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req dto.ResetPasswordRequest
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

	err := h.authService.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Error al restablecer la contraseña",
			Error:   err.Error(),
		})
	}

	return c.JSON(dto.ApiResponse{
		Success: true,
		Message: "Contraseña restablecida exitosamente. Ahora puedes iniciar sesión.",
	})
}
