package main

import (
    "log"

    "github.com/gofiber/fiber/v2"
    "github.com/rodolfodpk/instagrano/internal/config"
    "github.com/rodolfodpk/instagrano/internal/repository/postgres"
)

func main() {
    cfg := config.Load()

    db, err := postgres.Connect(cfg.DatabaseURL)
    if err != nil {
        log.Fatal("Database connection failed:", err)
    }
    defer db.Close()

    app := fiber.New()

    app.Get("/health", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "status":   "ok",
            "database": "connected",
        })
    })

    log.Printf("🚀 Server starting on port %s", cfg.Port)
    log.Fatal(app.Listen(":" + cfg.Port))
}
