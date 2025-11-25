package agent

import (
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/agent"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	usecase agent.IAgentUsecase
}

func NewHandler(usecase agent.IAgentUsecase) *Handler {
	return &Handler{usecase: usecase}
}

// authMiddleware extracts Bearer token from Authorization header
func (h *Handler) authMiddleware(c *fiber.Ctx) (string, error) {
	auth := c.Get("Authorization")
	if auth == "" {
		// Fallbacks for environments that strip Authorization
		if alt := c.Get("X-Api-Key"); alt != "" {
			return alt, nil
		}
		if qp := c.Query("token"); qp != "" {
			return qp, nil
		}
	}
	if auth == "" {
		return "", fiber.NewError(401, "UNAUTHORIZED: Missing Authorization header")
	}

	parts := strings.Fields(auth)
	if len(parts) < 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", fiber.NewError(401, "UNAUTHORIZED: Invalid Authorization format")
	}
	// Support users who typed "Bearer <token>" or accidentally "Bearer Bearer <token>"
	return parts[len(parts)-1], nil
}

// POST /agents/:agentId/run
func (h *Handler) ExecuteRun(c *fiber.Ctx) error {
	agentID := c.Params("agentId")
	if agentID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_PAYLOAD",
				"message": "agentId is required",
			},
		})
	}

	apiKey, err := h.authMiddleware(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": err.Error(),
			},
		})
	}

	var req agent.RunRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_PAYLOAD",
				"message": "Invalid request body",
			},
		})
	}

	resp, err := h.usecase.ExecuteRun(agentID, apiKey, req)
	if err != nil {
		statusCode := 500
		errorCode := "INTERNAL_ERROR"

		if strings.Contains(err.Error(), "UNAUTHORIZED") {
			statusCode = 401
			errorCode = "UNAUTHORIZED"
		} else if strings.Contains(err.Error(), "SESSION_NOT_FOUND") {
			statusCode = 404
			errorCode = "SESSION_NOT_FOUND"
		} else if strings.Contains(err.Error(), "SESSION_NOT_READY") {
			statusCode = 409
			errorCode = "SESSION_NOT_READY"
		} else if strings.Contains(err.Error(), "AI_TIMEOUT") {
			statusCode = 504
			errorCode = "AI_TIMEOUT"
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    errorCode,
				"message": err.Error(),
			},
		})
	}

	return c.JSON(resp)
}

// POST /agents/:agentId/messages
func (h *Handler) SendMessage(c *fiber.Ctx) error {
	agentID := c.Params("agentId")
	if agentID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_PAYLOAD",
				"message": "agentId is required",
			},
		})
	}

	apiKey, err := h.authMiddleware(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": err.Error(),
			},
		})
	}

	var req agent.SendMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_PAYLOAD",
				"message": "Invalid request body",
			},
		})
	}

	resp, err := h.usecase.SendMessage(agentID, apiKey, req)
	if err != nil {
		statusCode := 500
		errorCode := "INTERNAL_ERROR"

		if strings.Contains(err.Error(), "SESSION_NOT_READY") {
			statusCode = 409
			errorCode = "SESSION_NOT_READY"
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    errorCode,
				"message": err.Error(),
			},
		})
	}

	return c.JSON(resp)
}

// POST /agents/:agentId/media
func (h *Handler) SendMedia(c *fiber.Ctx) error {
	agentID := c.Params("agentId")
	if agentID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_PAYLOAD",
				"message": "agentId is required",
			},
		})
	}

	apiKey, err := h.authMiddleware(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "UNAUTHORIZED",
				"message": err.Error(),
			},
		})
	}

	var req agent.SendMediaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_PAYLOAD",
				"message": "Invalid request body",
			},
		})
	}

	resp, err := h.usecase.SendMedia(agentID, apiKey, req)
	if err != nil {
		statusCode := 500
		errorCode := "INTERNAL_ERROR"

		if strings.Contains(err.Error(), "SESSION_NOT_READY") {
			statusCode = 409
			errorCode = "SESSION_NOT_READY"
		} else if strings.Contains(err.Error(), "MEDIA_TOO_LARGE") {
			statusCode = 413
			errorCode = "MEDIA_TOO_LARGE"
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    errorCode,
				"message": err.Error(),
			},
		})
	}

	return c.JSON(resp)
}
