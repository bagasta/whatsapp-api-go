package dashboard

import (
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/dashboard"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	usecase dashboard.IDashboardUsecase
}

func NewHandler(usecase dashboard.IDashboardUsecase) *Handler {
	return &Handler{usecase: usecase}
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req struct {
		AgentID string `json:"agent_id"`
		ApiKey  string `json:"api_key"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}
	token, err := h.usecase.Login(req.AgentID, req.ApiKey)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"token": token, "agent_id": req.AgentID})
}

func (h *Handler) GetAnalytics(c *fiber.Ctx) error {
	agentID := c.Params("agentId")
	auth := c.Get("Authorization")
	parts := strings.Fields(auth)
	apiKey := ""
	if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
		apiKey = parts[1]
	}

	// Verify access (reuse Login logic for verification)
	_, err := h.usecase.Login(agentID, apiKey)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	analytics, err := h.usecase.GetAnalytics(agentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(analytics)
}
