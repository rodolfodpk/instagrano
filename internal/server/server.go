package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/instagrano/internal/docs"
	"github.com/yourusername/instagrano/internal/handlers"
)

func Start() {
	app := fiber.New()
	docs.SetupSwagger(app)
	app.Get("/", handlers.Home)
	app.Post("/register", handlers.Register)
	app.Post("/login", handlers.Login)
	app.Post("/upload", handlers.Upload)
	app.Get("/feed", handlers.Feed)
	app.Post("/posts/:id/like", handlers.Like)
	app.Post("/posts/:id/comments", handlers.Comment)
	app.Listen(":3000")
}
