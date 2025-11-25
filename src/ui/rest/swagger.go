package rest

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

//go:embed swagger-ui/*
//go:embed openapi.yaml
var swaggerUIFS embed.FS

// InitSwagger initializes the Swagger UI routes
func InitSwagger(app fiber.Router) {
	// Serve the OpenAPI specification
	app.Get("/docs/openapi.yaml", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/x-yaml")
		data, err := swaggerUIFS.ReadFile("openapi.yaml")
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "OpenAPI specification not found",
				"details": err.Error(),
			})
		}
		return c.Send(data)
	})

	// Create a sub-filesystem for swagger-ui files
	swaggerFS, _ := fs.Sub(swaggerUIFS, "swagger-ui")

	// Serve Swagger UI
	app.Use("/docs/swagger", filesystem.New(filesystem.Config{
		Root:   http.FS(swaggerFS),
		Browse: true,
	}))

	// Redirect /docs to Swagger UI
	app.Get("/docs", func(c *fiber.Ctx) error {
		return c.Redirect("/docs/swagger/index.html")
	})

	// Main swagger index route
	app.Get("/swagger", func(c *fiber.Ctx) error {
		return c.Redirect("/docs/swagger/index.html")
	})
}