package dashboard

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/domains/dashboard"
	"github.com/gofiber/fiber/v2"
)

func InitRoutes(app fiber.Router, usecase dashboard.IDashboardUsecase) {
	h := NewHandler(usecase)
	group := app.Group("/dashboard")
	group.Post("/login", h.Login)
	group.Get("/analytics/:agentId", h.GetAnalytics)
}
