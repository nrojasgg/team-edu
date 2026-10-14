package handlers

import (
	"encoding/json"
	"fmt"
	"ms-courses/internal/services"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type StudentSectionHandler struct {
	service *services.StudentSectionService
}

func NewStudentSectionHandler(service *services.StudentSectionService) *StudentSectionHandler {
	return &StudentSectionHandler{service: service}
}

// GetStudentsBySectionID obtiene todos los estudiantes de una sección específica
// @Summary Get students by section ID
// @Description Get all students enrolled in a specific section
// @Tags student-sections
// @Accept json
// @Produce json
// @Param id path int true "Section ID"
// @Success 200 {array} models.StudentSection
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /sections/{id}/students [get]
func (h *StudentSectionHandler) GetStudentsBySectionID(c *fiber.Ctx) error {
	sectionIDStr := c.Params("id")
	sectionID, err := strconv.ParseInt(sectionIDStr, 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid section ID",
		})
	}

	studentSections, err := h.service.GetStudentsBySectionID(sectionID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Error getting students for section: %v", err),
		})
	}

	// Extraer solo los IDs
	ids := make([]int64, 0, len(studentSections))
	for _, ss := range studentSections {
		ids = append(ids, ss.StudentID)
	}
	return c.JSON(fiber.Map{"ids": ids})
}

// GetSectionsByStudentID obtiene todas las secciones de un estudiante específico
// @Summary Get sections by student ID
// @Description Get all sections where a student is enrolled
// @Tags student-sections
// @Accept json
// @Produce json
// @Param id path int true "Student ID"
// @Success 200 {array} models.StudentSection
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /students/{id}/sections [get]
func (h *StudentSectionHandler) GetSectionsByStudentID(c *fiber.Ctx) error {
	studentIDStr := c.Params("id")
	studentID, err := strconv.ParseInt(studentIDStr, 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid student ID",
		})
	}

	studentSections, err := h.service.GetSectionsByStudentID(studentID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Error getting sections for student: %v", err),
		})
	}

	return c.JSON(studentSections)
}

// AddStudentToSection agrega un estudiante a una sección
// @Summary Add student to section
// @Description Enroll a student in a section
// @Tags student-sections
// @Accept json
// @Produce json
// @Param body body map[string]int64 true "Student and Section IDs"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /student-sections [post]
func (h *StudentSectionHandler) AddStudentToSection(c *fiber.Ctx) error {
	var req struct {
		StudentID int64 `json:"student_id"`
		SectionID int64 `json:"section_id"`
	}

	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.StudentID == 0 || req.SectionID == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Student ID and Section ID are required",
		})
	}

	err := h.service.AddStudentToSection(req.StudentID, req.SectionID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Error adding student to section: %v", err),
		})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"message": "Student added to section successfully",
	})
}

// RemoveStudentFromSection remueve un estudiante de una sección
// @Summary Remove student from section
// @Description Remove a student from a section
// @Tags student-sections
// @Accept json
// @Produce json
// @Param body body map[string]int64 true "Student and Section IDs"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /student-sections [delete]
func (h *StudentSectionHandler) RemoveStudentFromSection(c *fiber.Ctx) error {
	var req struct {
		StudentID int64 `json:"student_id"`
		SectionID int64 `json:"section_id"`
	}

	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.StudentID == 0 || req.SectionID == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Student ID and Section ID are required",
		})
	}

	err := h.service.RemoveStudentFromSection(req.StudentID, req.SectionID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Error removing student from section: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Student removed from section successfully",
	})
}

// RemoveStudentFromSectionByParams remueve un estudiante de una sección usando parámetros de ruta
// @Summary Remove student from section by path params
// @Description Remove a student from a section using path parameters
// @Tags student-sections
// @Produce json
// @Param sectionId path int true "Section ID"
// @Param studentId path int true "Student ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /sections/{sectionId}/students/{studentId} [delete]
func (h *StudentSectionHandler) RemoveStudentFromSectionByParams(c *fiber.Ctx) error {
	sectionIDStr := c.Params("sectionId")
	studentIDStr := c.Params("studentId")

	sectionID, err := strconv.ParseInt(sectionIDStr, 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid section ID",
		})
	}

	studentID, err := strconv.ParseInt(studentIDStr, 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid student ID",
		})
	}

	err = h.service.RemoveStudentFromSection(studentID, sectionID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Error removing student from section: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Student removed from section successfully",
	})
}
