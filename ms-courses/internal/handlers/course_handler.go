package handlers

import (
	"ms-courses/internal/models"
	"ms-courses/internal/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type CourseHandler struct {
	service *services.CourseService
}

func NewCourseHandler(service *services.CourseService) *CourseHandler {
	return &CourseHandler{service: service}
}

// GetAllCourses obtiene todos los cursos
// @Summary Obtener todos los cursos
// @Description Obtiene la lista completa de cursos
// @Tags courses
// @Accept json
// @Produce json
// @Success 200 {array} models.Course
// @Router /courses [get]
func (h *CourseHandler) GetAllCourses(c *fiber.Ctx) error {
	courses, err := h.service.GetAllCourses()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get courses",
		})
	}

	return c.JSON(courses)
}

// GetCourseByID obtiene un curso por ID
// @Summary Obtener un curso por ID
// @Description Obtiene un curso específico por su ID
// @Tags courses
// @Accept json
// @Produce json
// @Param id path int true "Course ID"
// @Success 200 {object} models.Course
// @Router /courses/{id} [get]
func (h *CourseHandler) GetCourseByID(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid course ID",
		})
	}

	course, err := h.service.GetCourseByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Course not found",
		})
	}

	return c.JSON(course)
}

// CreateCourse crea un nuevo curso
// @Summary Crear un nuevo curso
// @Description Crea un nuevo curso en el sistema
// @Tags courses
// @Accept json
// @Produce json
// @Param course body models.Course true "Course data"
// @Success 201 {object} models.Course
// @Router /courses [post]
func (h *CourseHandler) CreateCourse(c *fiber.Ctx) error {
	var course models.Course
	if err := c.BodyParser(&course); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	createdCourse, err := h.service.CreateCourse(&course)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create course",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(createdCourse)
}

// UpdateCourse actualiza un curso existente
// @Summary Actualizar un curso
// @Description Actualiza un curso existente
// @Tags courses
// @Accept json
// @Produce json
// @Param id path int true "Course ID"
// @Param course body models.Course true "Course data"
// @Success 200 {object} models.Course
// @Router /courses/{id} [put]
func (h *CourseHandler) UpdateCourse(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid course ID",
		})
	}

	var course models.Course
	if err := c.BodyParser(&course); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	course.ID = id
	updatedCourse, err := h.service.UpdateCourse(&course)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update course",
		})
	}

	return c.JSON(updatedCourse)
}

// DeleteCourse elimina un curso
// @Summary Eliminar un curso
// @Description Elimina un curso del sistema y todas sus secciones relacionadas
// @Tags courses
// @Accept json
// @Produce json
// @Param id path int true "Course ID"
// @Success 204
// @Router /courses/{id} [delete]
func (h *CourseHandler) DeleteCourse(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid course ID",
		})
	}

	err = h.service.DeleteCourse(id)
	if err != nil {
		// Log the actual error for debugging
		c.Locals("error", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to delete course",
			"details": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
