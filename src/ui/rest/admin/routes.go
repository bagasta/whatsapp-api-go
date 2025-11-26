package admin

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/webhook"
	"github.com/gofiber/fiber/v2"
)

func InitRoutes(app fiber.Router, usecase webhook.IWebhookConfigUsecase) {
	handler := NewHandler(usecase)

	adminGroup := app.Group("/admin")
	adminGroup.Get("/webhook-config", handler.GetConfig)
	adminGroup.Post("/webhook-config", handler.SaveConfig)
	adminGroup.Get("/sessions", handler.ListSessions)
	adminGroup.Get("/sessions/:agentId/webhook", handler.GetAgentConfig)
	adminGroup.Post("/sessions/:agentId/webhook", handler.SaveAgentConfig)
}
