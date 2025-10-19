package handler

import (
    "strconv"

    "github.com/gofiber/fiber/v2"
    "github.com/rodolfodpk/instagrano/internal/domain"
    "github.com/rodolfodpk/instagrano/internal/service"
)

type PostHandler struct {
    postService *service.PostService
}

func NewPostHandler(postService *service.PostService) *PostHandler {
    return &PostHandler{postService: postService}
}

func (h *PostHandler) CreatePost(c *fiber.Ctx) error {
    userID := c.Locals("userID").(uint)

    var req struct {
        Title     string           `json:"title"`
        Caption   string           `json:"caption"`
        MediaType domain.MediaType `json:"media_type"`
    }

    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
    }

    post, err := h.postService.CreatePost(userID, req.Title, req.Caption, req.MediaType)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }

    return c.Status(201).JSON(post)
}

func (h *PostHandler) GetPost(c *fiber.Ctx) error {
    id, err := strconv.ParseUint(c.Params("id"), 10, 32)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "invalid post id"})
    }

    post, err := h.postService.GetPost(uint(id))
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "post not found"})
    }

    return c.JSON(post)
}
