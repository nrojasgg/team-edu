package handlers

import (
	"errors"
	"strings"

	"ms-core/internal/core"

	"github.com/gofiber/fiber/v2"
)

// pluginExecError inspecciona el error retornado por PluginManager.ExecutePlugin
// y devuelve la respuesta HTTP apropiada:
//
//   - ErrPluginDisabled → 503 Service Unavailable
//   - "plugin not found" → 404 Not Found
//   - cualquier otro error → 500 Internal Server Error
//
// Uso típico en un handler:
//
//	result, err := h.pluginManager.ExecutePlugin("enneagram", request)
//	if err != nil {
//	    return pluginExecError(c, err)
//	}
func pluginExecError(c *fiber.Ctx, err error) error {
	if errors.Is(err, core.ErrPluginDisabled) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error":   true,
			"message": "plugin is currently disabled",
			"detail":  err.Error(),
		})
	}

	if strings.Contains(err.Error(), "not found") {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "plugin not found",
			"detail":  err.Error(),
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error":   true,
		"message": err.Error(),
	})
}
