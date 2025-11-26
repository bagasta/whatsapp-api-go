package admin

import (
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/webhook"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	usecase webhook.IWebhookConfigUsecase
}

func NewHandler(usecase webhook.IWebhookConfigUsecase) *Handler {
	return &Handler{usecase: usecase}
}

// GET /admin/webhook-config (default/fallback)
func (h *Handler) GetConfig(c *fiber.Ctx) error {
	cfg, err := h.usecase.GetDefault()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": err.Error(),
			},
		})
	}

	if cfg == nil {
		return c.JSON(fiber.Map{
			"url":    "",
			"secret": "",
		})
	}

	return c.JSON(cfg)
}

type saveRequest struct {
	URL    string `json:"url"`
	Secret string `json:"secret"`
}

// POST /admin/webhook-config (default/fallback)
func (h *Handler) SaveConfig(c *fiber.Ctx) error {
	var req saveRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_PAYLOAD",
				"message": "Invalid request body",
			},
		})
	}

	if strings.TrimSpace(req.URL) == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_PAYLOAD",
				"message": "url is required",
			},
		})
	}

	cfg, err := h.usecase.SaveDefault(req.URL, req.Secret)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": err.Error(),
			},
		})
	}

	return c.JSON(cfg)
}

// GET /admin/sessions
func (h *Handler) ListSessions(c *fiber.Ctx) error {
	sessions, err := h.usecase.ListSessions()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": err.Error(),
			},
		})
	}
	return c.JSON(fiber.Map{"sessions": sessions})
}

// GET /admin/sessions/:agentId/webhook
func (h *Handler) GetAgentConfig(c *fiber.Ctx) error {
	agentID := c.Params("agentId")
	if agentID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_PAYLOAD",
				"message": "agentId is required",
			},
		})
	}

	cfg, err := h.usecase.GetForAgent(agentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": err.Error(),
			},
		})
	}
	if cfg == nil {
		return c.JSON(fiber.Map{
			"agentId": agentID,
			"url":     "",
			"secret":  "",
		})
	}
	return c.JSON(cfg)
}

// POST /admin/sessions/:agentId/webhook
func (h *Handler) SaveAgentConfig(c *fiber.Ctx) error {
	agentID := c.Params("agentId")
	if agentID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_PAYLOAD",
				"message": "agentId is required",
			},
		})
	}

	var req saveRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_PAYLOAD",
				"message": "Invalid request body",
			},
		})
	}

	if strings.TrimSpace(req.URL) == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_PAYLOAD",
				"message": "url is required",
			},
		})
	}

	cfg, err := h.usecase.SaveForAgent(agentID, req.URL, req.Secret)
	if err != nil {
		status := 500
		code := "INTERNAL_ERROR"
		msg := err.Error()
		if strings.Contains(strings.ToLower(msg), "session not found") {
			status = 404
			code = "SESSION_NOT_FOUND"
		}
		return c.Status(status).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    code,
				"message": msg,
			},
		})
	}

	return c.JSON(cfg)
}
