package session

import (
	"encoding/json"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/session"
	"github.com/gofiber/fiber/v2"
	"io"
	"strings"
)

type Handler struct {
	usecase session.ISessionUsecase
}

func NewHandler(usecase session.ISessionUsecase) *Handler {
	return &Handler{usecase: usecase}
}

// POST /sessions
func (h *Handler) CreateSession(c *fiber.Ctx) error {
	var req session.CreateSessionRequest
	if err := c.BodyParser(&req); err != nil {
		if err != io.EOF { // allow empty body to be validated below
			// Try to provide clearer parse error
			return c.Status(400).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "INVALID_PAYLOAD",
					"message": "Invalid request body: " + err.Error(),
				},
			})
		}
	}

	// fallback: if BodyParser failed to decode but raw body exists, attempt manual unmarshal for clearer error
	if req.UserID == "" && req.AgentID == "" && len(c.Body()) > 0 {
		var alt session.CreateSessionRequest
		if err := json.Unmarshal(c.Body(), &alt); err == nil {
			req = alt
		} else {
			return c.Status(400).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "INVALID_PAYLOAD",
					"message": "Cannot parse JSON: " + err.Error(),
				},
			})
		}
	}

	// basic validation
	if strings.TrimSpace(req.UserID) == "" || strings.TrimSpace(req.AgentID) == "" || strings.TrimSpace(req.AgentName) == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_PAYLOAD",
				"message": "userId, agentId, agentName are required",
			},
		})
	}

	// Validate required fields
	if req.UserID == "" || req.AgentID == "" || req.AgentName == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_PAYLOAD",
				"message": "userId, agentId, and agentName are required",
			},
		})
	}

	resp, err := h.usecase.CreateSession(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": err.Error(),
			},
		})
	}

	return c.JSON(resp)
}

// GET /sessions/:agentId
func (h *Handler) GetSession(c *fiber.Ctx) error {
	agentID := c.Params("agentId")
	if agentID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_PAYLOAD",
				"message": "agentId is required",
			},
		})
	}

	resp, err := h.usecase.GetSession(agentID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "SESSION_NOT_FOUND",
				"message": err.Error(),
			},
		})
	}

	return c.JSON(resp)
}

// DELETE /sessions/:agentId
func (h *Handler) DeleteSession(c *fiber.Ctx) error {
	agentID := c.Params("agentId")
	if agentID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_PAYLOAD",
				"message": "agentId is required",
			},
		})
	}

	err := h.usecase.DeleteSession(agentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": err.Error(),
			},
		})
	}

	return c.JSON(fiber.Map{
		"deleted": true,
	})
}

// POST /sessions/:agentId/reconnect
func (h *Handler) ReconnectSession(c *fiber.Ctx) error {
	agentID := c.Params("agentId")
	if agentID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_PAYLOAD",
				"message": "agentId is required",
			},
		})
	}

	resp, err := h.usecase.ReconnectSession(agentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": err.Error(),
			},
		})
	}

	return c.JSON(resp)
}

// POST /sessions/:agentId/qr
func (h *Handler) GetQR(c *fiber.Ctx) error {
	agentID := c.Params("agentId")
	if agentID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INVALID_PAYLOAD",
				"message": "agentId is required",
			},
		})
	}

	resp, err := h.usecase.GetQR(agentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": err.Error(),
			},
		})
	}

	return c.JSON(resp)
}
