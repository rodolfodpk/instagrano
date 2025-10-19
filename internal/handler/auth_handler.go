package handler

import (
    "github.com/gofiber/fiber/v2"
    "github.com/rodolfodpk/instagrano/internal/service"
)

type AuthHandler struct {
    authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
    return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
    var req struct {
        Username string `json:"username"`
        Email    string `json:"email"`
        Password string `json:"password"`
    }

    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
    }

    user, err := h.authService.Register(req.Username, req.Email, req.Password)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }

    user.Password = ""
    return c.JSON(fiber.Map{"user": user})
}
