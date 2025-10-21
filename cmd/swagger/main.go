package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

func main() {
	app := fiber.New()

	// Get port from environment or use default 8081
	port := os.Getenv("SWAGGER_PORT")
	if port == "" {
		port = "8081"
	}

	// Serve static docs directory
	app.Static("/docs", "./docs")

	// Swagger UI
	app.Get("/swagger/*", swagger.New(swagger.Config{
		URL:         "/docs/swagger.json",
		DeepLinking: false,
	}))

	// Redirect root to Swagger UI
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/swagger/")
	})

	log.Printf("Swagger UI starting on port %s", port)
	log.Printf("Visit http://localhost:%s/swagger/", port)
	log.Fatal(app.Listen(":" + port))
}
