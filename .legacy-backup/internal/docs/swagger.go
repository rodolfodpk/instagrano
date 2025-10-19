package docs

import (
	"os"

	"github.com/gofiber/fiber/v2"
)

func SetupSwagger(app *fiber.App) {
	// Serve the complete swagger.json file
	app.Get("/docs/doc.json", func(c *fiber.Ctx) error {
		// Read the complete swagger.json file
		swaggerData, err := os.ReadFile("docs/swagger.json")
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to read swagger file: " + err.Error()})
		}

		c.Set("Content-Type", "application/json")
		return c.Send(swaggerData)
	})

	// Serve a simple Swagger UI HTML page
	app.Get("/docs", func(c *fiber.Ctx) error {
		html := `<!DOCTYPE html>
<html>
<head>
    <title>Instagrano API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui.css" />
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-bundle.js"></script>
    <script>
        SwaggerUIBundle({
            url: '/docs/doc.json',
            dom_id: '#swagger-ui',
            presets: [
                SwaggerUIBundle.presets.apis,
                SwaggerUIBundle.presets.standalone
            ]
        });
    </script>
</body>
</html>`
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	})
}
