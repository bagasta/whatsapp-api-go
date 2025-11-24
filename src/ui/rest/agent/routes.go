package agent

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/agent"
	"github.com/gofiber/fiber/v2"
)

func InitRoutes(app fiber.Router, usecase agent.IAgentUsecase) {
	handler := NewHandler(usecase)

	// Agent routes
	app.Post("/agents/:agentId/run", handler.ExecuteRun)
	app.Post("/agents/:agentId/messages", handler.SendMessage)
	app.Post("/agents/:agentId/media", handler.SendMedia)
}
