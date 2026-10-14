package handlers

import (
	"ms-courses/internal/models"
	"ms-courses/internal/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type CourseOfferingHandler struct {
	service *services.CourseOfferingService
}

func NewCourseOfferingHandler(service *services.CourseOfferingService) *CourseOfferingHandler {
	return &CourseOfferingHandler{service: service}
}

func (h *CourseOfferingHandler) GetAllCourseOfferings(c *fiber.Ctx) error {
	courseOfferings, err := h.service.GetAllCourseOfferings()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get course offerings",
		})
	}

	return c.JSON(courseOfferings)
}

func (h *CourseOfferingHandler) GetCourseOfferingByID(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid course offering ID",
		})
	}

	courseOffering, err := h.service.GetCourseOfferingByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Course offering not found",
		})
	}

	return c.JSON(courseOffering)
}

func (h *CourseOfferingHandler) CreateCourseOffering(c *fiber.Ctx) error {
	var courseOffering models.CourseOffering
	if err := c.BodyParser(&courseOffering); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	createdCourseOffering, err := h.service.CreateCourseOffering(&courseOffering)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create course offering",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(createdCourseOffering)
}

func (h *CourseOfferingHandler) DeleteCourseOffering(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid course offering ID",
		})
	}

	err = h.service.DeleteCourseOffering(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete course offering",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
