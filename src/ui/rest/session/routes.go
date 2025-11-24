package session

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/session"
	"github.com/gofiber/fiber/v2"
)

func InitRoutes(app fiber.Router, usecase session.ISessionUsecase) {
	handler := NewHandler(usecase)

	// Session management routes
	app.Post("/sessions", handler.CreateSession)
	app.Get("/sessions/:agentId", handler.GetSession)
	app.Delete("/sessions/:agentId", handler.DeleteSession)
	app.Post("/sessions/:agentId/reconnect", handler.ReconnectSession)
	app.Post("/sessions/:agentId/qr", handler.GetQR)
}
