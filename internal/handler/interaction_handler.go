package handler

import (
    "strconv"

    "github.com/gofiber/fiber/v2"
    "github.com/rodolfodpk/instagrano/internal/service"
)

type InteractionHandler struct {
    interactionService *service.InteractionService
}

func NewInteractionHandler(interactionService *service.InteractionService) *InteractionHandler {
    return &InteractionHandler{interactionService: interactionService}
}

func (h *InteractionHandler) LikePost(c *fiber.Ctx) error {
    userID := c.Locals("userID").(uint)
    postID, err := strconv.ParseUint(c.Params("id"), 10, 32)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "invalid post id"})
    }

    err = h.interactionService.LikePost(userID, uint(postID))
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }

func (h *InteractionHandler) CommentPost(c *fiber.Ctx) error {
    userID := c.Locals("userID").(uint)
    postID, err := strconv.ParseUint(c.Params("id"), 10, 32)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "invalid post id"})
    }

    var req struct {
        Text string `json:"text"`
    }

    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
    }

    err = h.interactionService.CommentPost(userID, uint(postID), req.Text)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(fiber.Map{"message": "comment added"})
}
