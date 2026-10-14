package handlers

import (
	"ms-courses/internal/models"
	"ms-courses/internal/repositories"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type InvitationCodeHandler struct {
	repo *repositories.InvitationCodeRepository
}

func NewInvitationCodeHandler(repo *repositories.InvitationCodeRepository) *InvitationCodeHandler {
	return &InvitationCodeHandler{repo: repo}
}

func (h *InvitationCodeHandler) CreateInvitationCode(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	professorID, ok := userID.(int64)
	if !ok {
		professorID = int64(userID.(float64))
	}

	var req struct {
		SectionID interface{} `json:"section_id"`
		ExpiresAt time.Time   `json:"expires_at"`
		MaxUses   *int        `json:"max_uses,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var sectionID int64
	switch v := req.SectionID.(type) {
	case float64:
		sectionID = int64(v)
	case string:
		sectionID, _ = strconv.ParseInt(v, 10, 64)
	default:
		sectionID = 0
	}

	if sectionID == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Section ID is required",
		})
	}

	if req.ExpiresAt.IsZero() {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Expiration date is required",
		})
	}

	ic, err := h.repo.Create(&models.CreateInvitationCodeRequest{
		SectionID: sectionID,
		ExpiresAt: req.ExpiresAt,
		MaxUses:   req.MaxUses,
	}, professorID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(ic)
}

func (h *InvitationCodeHandler) GetSectionInvitationCodes(c *fiber.Ctx) error {
	sectionIDStr := c.Params("id")
	sectionID, err := strconv.ParseInt(sectionIDStr, 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid section ID",
		})
	}

	codes, err := h.repo.GetActiveBySection(sectionID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(codes)
}

func (h *InvitationCodeHandler) DeactivateInvitationCode(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	err = h.repo.Deactivate(id)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Invitation code deactivated successfully",
	})
}

func (h *InvitationCodeHandler) UseInvitationCode(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	studentID, ok := userID.(int64)
	if !ok {
		studentID = int64(userID.(float64))
	}

	var req struct {
		Code string `json:"code"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Code == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Code is required",
		})
	}

	ic, err := h.repo.GetByCode(req.Code)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Invitation code not found",
		})
	}

	if !ic.IsActive {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invitation code is not active",
		})
	}

	if time.Now().After(ic.ExpiresAt) {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invitation code has expired",
		})
	}

	if ic.MaxUses != nil && ic.UsesCount >= *ic.MaxUses {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invitation code has reached maximum uses",
		})
	}

	studentName, _ := c.Locals("name").(string)
	studentEmail, _ := c.Locals("email").(string)
	studentRUT, _ := c.Locals("rut").(string)

	err = h.repo.AddStudentToSection(studentID, ic.SectionID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	err = h.repo.IncrementUses(ic.ID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	err = h.repo.RecordUse(ic.ID, studentID, studentName, studentEmail, studentRUT)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Invitation code used successfully",
	})
}

func (h *InvitationCodeHandler) GetInvitationCodeUses(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	uses, err := h.repo.GetUsesByCodeID(id)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(uses)
}
