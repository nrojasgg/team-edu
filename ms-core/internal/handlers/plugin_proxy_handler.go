package handlers

import (
	"ms-core/internal/core"

	"github.com/gofiber/fiber/v2"
)

type GenericPluginProxy struct {
	base psychometricTestHandler
}

func NewGenericPluginProxy(pluginManager *core.PluginManager, pluginName string) *GenericPluginProxy {
	return &GenericPluginProxy{
		base: newPsychometricTestHandler(pluginManager, pluginName),
	}
}

func (h *GenericPluginProxy) CreateTest(c *fiber.Ctx) error {
	return h.base.createTest(c)
}

func (h *GenericPluginProxy) GetTests(c *fiber.Ctx) error {
	return h.base.getTests(c)
}

func (h *GenericPluginProxy) GetSurveys(c *fiber.Ctx) error {
	sectionID := c.Params("section_id")
	if sectionID == "" {
		sectionID = c.Query("section_id")
	}
	if sectionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "section_id is required",
		})
	}

	return h.base.execute(c, map[string]interface{}{
		"action":     "get_surveys",
		"section_id": sectionID,
	})
}

func (h *GenericPluginProxy) GetTest(c *fiber.Ctx) error {
	return h.base.getTest(c)
}

func (h *GenericPluginProxy) UpdateTest(c *fiber.Ctx) error {
	return h.base.updateTest(c)
}

func (h *GenericPluginProxy) DeleteTest(c *fiber.Ctx) error {
	return h.base.deleteTest(c)
}

func (h *GenericPluginProxy) SubmitResponse(c *fiber.Ctx) error {
	return h.base.submitResponse(c)
}

func (h *GenericPluginProxy) GetResponses(c *fiber.Ctx) error {
	return h.base.getResponses(c)
}

func (h *GenericPluginProxy) GetSectionResponses(c *fiber.Ctx) error {
	return h.base.getSectionResponses(c)
}

func (h *GenericPluginProxy) GetSectionResponsesGroupCompat(c *fiber.Ctx) error {
	sectionID := c.Params("sectionId")
	if sectionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "section_id is required",
		})
	}

	return h.base.execute(c, map[string]interface{}{
		"action":     "get_section_responses",
		"section_id": sectionID,
	})
}

func (h *GenericPluginProxy) GetSurveyResponses(c *fiber.Ctx) error {
	return h.base.getSurveyResponses(c)
}

func (h *GenericPluginProxy) GetStudentSurveyResponse(c *fiber.Ctx) error {
	surveyID := c.Params("surveyId")
	studentID := c.Params("studentId")
	if surveyID == "" || studentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "survey_id and student_id are required",
		})
	}

	return h.base.execute(c, map[string]interface{}{
		"action":     "get_student_response",
		"survey_id":  surveyID,
		"student_id": studentID,
	})
}

func (h *GenericPluginProxy) GetStudentEneatypes(c *fiber.Ctx) error {
	sectionID := c.Params("section_id")
	if sectionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "section_id is required",
		})
	}

	return h.base.execute(c, map[string]interface{}{
		"action":     "get_student_eneatypes",
		"section_id": sectionID,
	})
}

func (h *GenericPluginProxy) GetResults(c *fiber.Ctx) error {
	return h.base.getResults(c)
}

func (h *GenericPluginProxy) GetStudentTestResponse(c *fiber.Ctx) error {
	testID := c.Params("test_id")
	studentID := c.Params("studentId")
	if testID == "" || studentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "test_id and student_id are required",
		})
	}

	return h.base.execute(c, map[string]interface{}{
		"action":     "get_student_response",
		"test_id":    testID,
		"student_id": studentID,
	})
}

func (h *GenericPluginProxy) GetSurveyPeers(c *fiber.Ctx) error {
	request := map[string]interface{}{}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request body",
		})
	}
	request["action"] = "get_survey_peers"
	return h.base.execute(c, request)
}

