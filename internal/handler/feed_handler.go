package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/rodolfodpk/instagrano/internal/service"
	"go.uber.org/zap"
)

type FeedHandler struct {
	feedService *service.FeedService
	logger      *zap.Logger
}

func NewFeedHandler(feedService *service.FeedService) *FeedHandler {
	logger, _ := zap.NewProduction()
	return &FeedHandler{
		feedService: feedService,
		logger:      logger,
	}
}

func (h *FeedHandler) GetFeed(c *fiber.Ctx) error {
	// Check if cursor-based pagination is requested
	cursor := c.Query("cursor")
	if cursor != "" {
		return h.getFeedWithCursor(c, cursor)
	}

	// Fallback to page-based pagination for backward compatibility
	return h.getFeedWithPage(c)
}

func (h *FeedHandler) getFeedWithCursor(c *fiber.Ctx, cursor string) error {
	limitStr := c.Query("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	h.logger.Info("getting feed with cursor",
		zap.String("cursor", cursor),
		zap.Int("limit", limit),
	)

	result, err := h.feedService.GetFeedWithCursor(limit, cursor)
	if err != nil {
		h.logger.Error("failed to get feed with cursor", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{"error": "failed to get feed"})
	}

	h.logger.Info("feed retrieved successfully",
		zap.Int("posts_count", len(result.Posts)),
		zap.Bool("has_more", result.HasMore),
		zap.String("next_cursor", result.NextCursor),
	)

	return c.JSON(result)
}

func (h *FeedHandler) getFeedWithPage(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	h.logger.Info("getting feed with page",
		zap.Int("page", page),
		zap.Int("limit", limit),
	)

	posts, err := h.feedService.GetFeed(page, limit)
	if err != nil {
		h.logger.Error("failed to get feed with page", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{"error": "failed to get feed"})
	}

	h.logger.Info("feed retrieved successfully",
		zap.Int("posts_count", len(posts)),
		zap.Int("page", page),
	)

	return c.JSON(fiber.Map{"posts": posts, "page": page, "limit": limit})
}
