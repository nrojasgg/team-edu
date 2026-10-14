package handlers

import (
	"ms-notis/internal/dto"
	"ms-notis/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type NotificationHandler struct {
	emailService services.EmailService
	validator    *validator.Validate
}

func NewNotificationHandler(emailService services.EmailService) *NotificationHandler {
	return &NotificationHandler{
		emailService: emailService,
		validator:    validator.New(),
	}
}

// SendPasswordReset sends a password reset email
// @Summary Send password reset email
// @Description Send an email with password reset link
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body dto.PasswordResetRequest true "Password reset data"
// @Success 200 {object} dto.ApiResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /notifications/password-reset [post]
func (h *NotificationHandler) SendPasswordReset(c *fiber.Ctx) error {
	var req dto.PasswordResetRequest
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

	if err := h.emailService.SendPasswordResetEmail(req.Email, req.Name, req.ResetToken, req.ResetURL); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Failed to send password reset email",
			Error:   err.Error(),
		})
	}

	return c.JSON(dto.ApiResponse{
		Success: true,
		Message: "Password reset email sent successfully",
	})
}

// SendTestReminder sends a test reminder email
// @Summary Send test reminder email
// @Description Send an email reminding student about a pending test
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body dto.TestReminderRequest true "Test reminder data"
// @Success 200 {object} dto.ApiResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /notifications/test-reminder [post]
func (h *NotificationHandler) SendTestReminder(c *fiber.Ctx) error {
	var req dto.TestReminderRequest
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

	if err := h.emailService.SendTestReminderEmail(req.Email, req.Name, req.TestName, req.TestURL, req.DueDate); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false,
			Message: "Failed to send test reminder email",
			Error:   err.Error(),
		})
	}

	return c.JSON(dto.ApiResponse{
		Success: true,
		Message: "Test reminder email sent successfully",
	})
}
