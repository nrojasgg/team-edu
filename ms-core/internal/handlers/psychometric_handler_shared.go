package handlers

import (
	"ms-core/internal/core"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type psychometricTestHandler struct {
	pluginManager *core.PluginManager
	pluginName    string
}

func newPsychometricTestHandler(pluginManager *core.PluginManager, pluginName string) psychometricTestHandler {
	return psychometricTestHandler{
		pluginManager: pluginManager,
		pluginName:    pluginName,
	}
}

func unwrapPluginResponse(result interface{}) interface{} {
	if m, ok := result.(map[string]interface{}); ok {
		// If the map has both "success" and "data" keys, let's unwrap it!
		if _, hasSuccess := m["success"]; hasSuccess {
			if dataVal, hasData := m["data"]; hasData {
				return dataVal
			}
		}
	}
	return result
}

func (h *psychometricTestHandler) parseBody(c *fiber.Ctx) (map[string]interface{}, error) {
	request := map[string]interface{}{}
	if err := c.BodyParser(&request); err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, errInvalidRequestBody)
	}
	return request, nil
}

func (h *psychometricTestHandler) execute(c *fiber.Ctx, request map[string]interface{}) error {
	result, err := h.pluginManager.ExecutePlugin(h.pluginName, request)
	if err != nil {
		return pluginExecError(c, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    unwrapPluginResponse(result),
	})
}

func (h *psychometricTestHandler) createTest(c *fiber.Ctx) error {
	request, err := h.parseBody(c)
	if err != nil {
		return err
	}

	request["action"] = "create_survey"
	result, err := h.pluginManager.ExecutePlugin(h.pluginName, request)
	if err != nil {
		if strings.Contains(err.Error(), "unknown action") {
			request["action"] = "create_test"
			return h.execute(c, request)
		}
		return pluginExecError(c, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    unwrapPluginResponse(result),
	})
}

func (h *psychometricTestHandler) getTests(c *fiber.Ctx) error {
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

	request := map[string]interface{}{
		"action":     "get_surveys",
		"section_id": sectionID,
	}
	result, err := h.pluginManager.ExecutePlugin(h.pluginName, request)
	if err != nil {
		if strings.Contains(err.Error(), "unknown action") {
			request["action"] = "get_tests"
			return h.execute(c, request)
		}
		return pluginExecError(c, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    unwrapPluginResponse(result),
	})
}

func (h *psychometricTestHandler) getTest(c *fiber.Ctx) error {
	testID := c.Params("id")
	if testID == "" {
		testID = c.Params("test_id")
	}
	if testID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": errTestIDRequired,
		})
	}

	request := map[string]interface{}{
		"action":    "get_survey",
		"survey_id": testID,
		"test_id":   testID,
	}
	result, err := h.pluginManager.ExecutePlugin(h.pluginName, request)
	if err != nil {
		if strings.Contains(err.Error(), "unknown action") {
			request["action"] = "get_test"
			return h.execute(c, request)
		}
		return pluginExecError(c, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    unwrapPluginResponse(result),
	})
}

func (h *psychometricTestHandler) submitResponse(c *fiber.Ctx) error {
	testID := c.Params("test_id")
	if testID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Test ID is required",
		})
	}

	request, err := h.parseBody(c)
	if err != nil {
		return err
	}

	request["action"] = "submit_response"
	request["test_id"] = testID
	return h.execute(c, request)
}

func (h *psychometricTestHandler) getResponses(c *fiber.Ctx) error {
	testID := c.Params("test_id")
	if testID == "" {
		testID = c.Params("surveyId")
	}
	if testID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": errTestIDRequired,
		})
	}

	request := map[string]interface{}{
		"action":    "get_responses",
		"test_id":   testID,
		"survey_id": testID,
	}
	for k, v := range c.Queries() {
		request[k] = v
	}

	return h.execute(c, request)
}

func (h *psychometricTestHandler) getResults(c *fiber.Ctx) error {
	testID := c.Params("test_id")
	if testID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": errTestIDRequired,
		})
	}

	request := map[string]interface{}{
		"action":    "get_results",
		"test_id":   testID,
		"survey_id": testID,
	}
	for k, v := range c.Queries() {
		request[k] = v
	}

	return h.execute(c, request)
}

func (h *psychometricTestHandler) getSectionResponses(c *fiber.Ctx) error {
	sectionID := c.Query("section_id")
	if sectionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "section_id query parameter is required",
		})
	}

	request := map[string]interface{}{
		"action":     "get_section_responses",
		"section_id": sectionID,
	}
	for k, v := range c.Queries() {
		request[k] = v
	}

	return h.execute(c, request)
}

func (h *psychometricTestHandler) getSurveyResponses(c *fiber.Ctx) error {
	surveyID := c.Params("surveyId")
	if surveyID == "" {
		surveyID = c.Params("test_id")
	}
	if surveyID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "survey_id is required",
		})
	}

	request := map[string]interface{}{
		"action":    "get_responses",
		"survey_id": surveyID,
		"test_id":   surveyID,
	}
	for k, v := range c.Queries() {
		request[k] = v
	}

	return h.execute(c, request)
}

func (h *psychometricTestHandler) updateTest(c *fiber.Ctx) error {
	testID := c.Params("id")
	if testID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": errTestIDRequired,
		})
	}

	request, err := h.parseBody(c)
	if err != nil {
		return err
	}

	request["action"] = "update_survey"
	request["survey_id"] = testID
	request["test_id"] = testID
	result, err := h.pluginManager.ExecutePlugin(h.pluginName, request)
	if err != nil {
		if strings.Contains(err.Error(), "unknown action") {
			request["action"] = "update_test"
			return h.execute(c, request)
		}
		return pluginExecError(c, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    unwrapPluginResponse(result),
	})
}

func (h *psychometricTestHandler) deleteTest(c *fiber.Ctx) error {
	testID := c.Params("id")
	if testID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": errTestIDRequired,
		})
	}

	request := map[string]interface{}{
		"action":    "delete_survey",
		"survey_id": testID,
		"test_id":   testID,
	}
	result, err := h.pluginManager.ExecutePlugin(h.pluginName, request)
	if err != nil {
		if strings.Contains(err.Error(), "unknown action") {
			request["action"] = "delete_test"
			return h.execute(c, request)
		}
		return pluginExecError(c, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    unwrapPluginResponse(result),
	})
}
