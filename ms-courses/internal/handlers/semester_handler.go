package handlers

import (
	"ms-courses/internal/models"
	"ms-courses/internal/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type SemesterHandler struct {
	service *services.SemesterService
}

func NewSemesterHandler(service *services.SemesterService) *SemesterHandler {
	return &SemesterHandler{service: service}
}

func (h *SemesterHandler) GetAllSemesters(c *fiber.Ctx) error {
	semesters, err := h.service.GetAllSemesters()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get semesters",
		})
	}

	return c.JSON(semesters)
}

func (h *SemesterHandler) GetSemesterByID(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid semester ID",
		})
	}

	semester, err := h.service.GetSemesterByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Semester not found",
		})
	}

	return c.JSON(semester)
}

func (h *SemesterHandler) CreateSemester(c *fiber.Ctx) error {
	var semester models.Semester
	if err := c.BodyParser(&semester); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	createdSemester, err := h.service.CreateSemester(&semester)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create semester",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(createdSemester)
}

func (h *SemesterHandler) UpdateSemester(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid semester ID",
		})
	}

	var semester models.Semester
	if err := c.BodyParser(&semester); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	semester.ID = id
	updatedSemester, err := h.service.UpdateSemester(&semester)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update semester",
		})
	}

	return c.JSON(updatedSemester)
}

func (h *SemesterHandler) DeleteSemester(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid semester ID",
		})
	}

	err = h.service.DeleteSemester(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete semester",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *SemesterHandler) GetCurrentSemester(c *fiber.Ctx) error {
	semester, err := h.service.GetCurrentSemester()
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No active semester found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    semester,
	})
}

func (h *SemesterHandler) SetActiveSemester(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid semester ID",
		})
	}

	semester, err := h.service.SetActiveSemester(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to set active semester",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Semester activated successfully",
		"data":    semester,
	})
}

type CreateOrActivateSemesterRequest struct {
	Year int `json:"year" validate:"required,min=2020,max=2050"`
	Term int `json:"term" validate:"required,min=1,max=2"`
}

func (h *SemesterHandler) CreateOrActivateSemester(c *fiber.Ctx) error {
	var req CreateOrActivateSemesterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	semester, err := h.service.CreateOrActivateSemester(req.Year, req.Term)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create or activate semester",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Semester created/activated successfully",
		"data":    semester,
	})
}
