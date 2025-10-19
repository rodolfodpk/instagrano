package docs

import (
	"github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"
)

func SetupSwagger(app *fiber.App) {
	app.Get("/docs", func(c *fiber.Ctx) {
		c.Set("Content-Type", "text/html")
		c.Send(swagger.New(swagger.Config{URL: "doc.json"}).HTML())
	})
}
