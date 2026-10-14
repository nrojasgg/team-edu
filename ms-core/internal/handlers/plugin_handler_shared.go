package handlers

import (
	"ms-core/internal/core"

	"github.com/gofiber/fiber/v2"
)

func parseRequestBody(c *fiber.Ctx) (map[string]interface{}, error) {
	request := map[string]interface{}{}
	if err := c.BodyParser(&request); err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, errInvalidRequestBody)
	}
	return request, nil
}

func executePluginData(c *fiber.Ctx, pluginManager *core.PluginManager, pluginName string, request map[string]interface{}) error {
	result, err := pluginManager.ExecutePlugin(pluginName, request)
	if err != nil {
		return pluginExecError(c, err)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

func executePluginRaw(c *fiber.Ctx, pluginManager *core.PluginManager, pluginName string, request map[string]interface{}) (interface{}, error) {
	result, err := pluginManager.ExecutePlugin(pluginName, request)
	if err != nil {
		return nil, pluginExecError(c, err)
	}

	return result, nil
}

func requiredParam(c *fiber.Ctx, key string, msg string) (string, error) {
	v := c.Params(key)
	if v == "" {
		return "", fiber.NewError(fiber.StatusBadRequest, msg)
	}
	return v, nil
}

func requiredQuery(c *fiber.Ctx, key string, msg string) (string, error) {
	v := c.Query(key)
	if v == "" {
		return "", fiber.NewError(fiber.StatusBadRequest, msg)
	}
	return v, nil
}
