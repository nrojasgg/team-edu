package proxy

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// ProxyToNotifications proxies requests to ms-notis microservice
func (p *ProxyService) ProxyToNotifications(c *fiber.Ctx) error {
	originalPath := c.OriginalURL()
	targetPath := strings.TrimPrefix(originalPath, "/api/v1/notifications")
	if targetPath == "" {
		targetPath = "/"
	}

	targetURL := fmt.Sprintf("%s/api/v1/notifications%s", p.NotificationsURL, targetPath)
	log.Printf("Proxying request from %s to %s", originalPath, targetURL)

	var bodyReader io.Reader
	if c.Body() != nil {
		bodyReader = bytes.NewReader(c.Body())
	}

	req, err := http.NewRequest(c.Method(), targetURL, bodyReader)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create proxy request",
		})
	}

	for key, values := range c.GetReqHeaders() {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		log.Printf("Error making request to ms-notis: %v", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "Failed to connect to notifications service",
		})
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read notifications service response",
		})
	}

	for key, values := range resp.Header {
		for _, value := range values {
			c.Set(key, value)
		}
	}

	return c.Status(resp.StatusCode).Send(body)
}
