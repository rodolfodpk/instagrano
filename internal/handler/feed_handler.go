package handler

import (
    "github.com/gofiber/fiber/v2"
    "github.com/rodolfodpk/instagrano/internal/service"
)

type FeedHandler struct {
    feedService *service.FeedService
}

func NewFeedHandler(feedService *service.FeedService) *FeedHandler {
    return &FeedHandler{feedService: feedService}
}

func (h *FeedHandler) GetFeed(c *fiber.Ctx) error {
    page := c.QueryInt("page", 1)
    limit := c.QueryInt("limit", 20)

    posts, err := h.feedService.GetFeed(page, limit)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "failed to get feed"})
    }

    return c.JSON(fiber.Map{"posts": posts, "page": page, "limit": limit})
}
