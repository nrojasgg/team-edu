package handlers

import (
	"ms-courses/internal/models"
	"ms-courses/internal/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type SectionHandler struct {
	service *services.SectionService
}

func NewSectionHandler(service *services.SectionService) *SectionHandler {
	return &SectionHandler{service: service}
}

func (h *SectionHandler) GetAllSections(c *fiber.Ctx) error {
	sections, err := h.service.GetAllSections()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get sections",
		})
	}

	return c.JSON(sections)
}

func (h *SectionHandler) GetSectionByID(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid section ID",
		})
	}

	section, err := h.service.GetSectionByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Section not found",
		})
	}

	return c.JSON(section)
}

func (h *SectionHandler) GetSectionsByCourseOffering(c *fiber.Ctx) error {
	courseOfferingID, err := strconv.ParseInt(c.Params("courseOfferingId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid course offering ID",
		})
	}

	sections, err := h.service.GetSectionsByCourseOffering(courseOfferingID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get sections",
		})
	}

	return c.JSON(sections)
}

func (h *SectionHandler) CreateSection(c *fiber.Ctx) error {
	var section models.Section
	if err := c.BodyParser(&section); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	createdSection, err := h.service.CreateSection(&section)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create section",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(createdSection)
}

func (h *SectionHandler) UpdateSection(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid section ID",
		})
	}

	var section models.Section
	if err := c.BodyParser(&section); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	section.ID = id
	updatedSection, err := h.service.UpdateSection(&section)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update section",
		})
	}

	return c.JSON(updatedSection)
}

func (h *SectionHandler) DeleteSection(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid section ID",
		})
	}

	err = h.service.DeleteSection(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete section",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateSectionStaff updates the professor and assistant assignments for a section
func (h *SectionHandler) UpdateSectionStaff(c *fiber.Ctx) error {
	sectionID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid section ID",
		})
	}

	var req struct {
		ProfessorID  *int64  `json:"professor_id"`
		AssistantID  *int64  `json:"assistant_id"`
		ProfessorIDs []int64 `json:"professor_ids"`
		AssistantIDs []int64 `json:"assistant_ids"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	profIDs := req.ProfessorIDs
	if len(profIDs) == 0 && req.ProfessorID != nil && *req.ProfessorID > 0 {
		profIDs = []int64{*req.ProfessorID}
	} else if profIDs == nil {
		profIDs = make([]int64, 0)
	}

	astIDs := req.AssistantIDs
	if len(astIDs) == 0 && req.AssistantID != nil && *req.AssistantID > 0 {
		astIDs = []int64{*req.AssistantID}
	} else if astIDs == nil {
		astIDs = make([]int64, 0)
	}

	section, err := h.service.UpdateStaff(sectionID, profIDs, astIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update staff",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Staff assigned successfully",
		"data":    section,
	})
}
