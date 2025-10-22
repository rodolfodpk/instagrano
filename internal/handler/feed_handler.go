package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rodolfodpk/instagrano/internal/config"
	"github.com/rodolfodpk/instagrano/internal/domain"
	"github.com/rodolfodpk/instagrano/internal/dto"
	"github.com/rodolfodpk/instagrano/internal/service"
	"go.uber.org/zap"
)

type FeedHandler struct {
	feedService *service.FeedService
	logger      *zap.Logger
	config      *config.Config
}

func NewFeedHandler(feedService *service.FeedService, cfg *config.Config) *FeedHandler {
	logger, _ := zap.NewProduction()
	return &FeedHandler{
		feedService: feedService,
		logger:      logger,
		config:      cfg,
	}
}

// GetFeed godoc
// @Summary      Get user feed
// @Description  Retrieve paginated feed using cursor-based pagination
// @Tags         feed
// @Produce      json
// @Security     BearerAuth
// @Param        cursor  query     string  false  "Pagination cursor"
// @Param        limit   query     int     false  "Number of posts (default 20, max 100)"
// @Success      200  {object}  pagination.FeedResult
// @Failure      500  {object}  object{error=string}
// @Router       /feed [get]
func (h *FeedHandler) GetFeed(c *fiber.Ctx) error {
	// Always use cursor-based pagination (more efficient)
	cursor := c.Query("cursor")
	return h.getFeedWithCursor(c, cursor)
}

func (h *FeedHandler) getFeedWithCursor(c *fiber.Ctx, cursor string) error {
	limitStr := c.Query("limit", strconv.Itoa(h.config.DefaultPageSize))
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > h.config.MaxPageSize {
		limit = h.config.DefaultPageSize
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

	// Convert to DTO
	response := dto.FeedResponse{
		Posts:      make([]*dto.PostResponse, len(result.Posts)),
		NextCursor: result.NextCursor,
		HasMore:    result.HasMore,
	}
	for i, post := range result.Posts {
		// Handle both cached (map[string]interface{}) and fresh (*domain.Post) data
		if domainPost, ok := post.(*domain.Post); ok {
			response.Posts[i] = dto.ToPostResponse(domainPost)
		} else if postMap, ok := post.(map[string]interface{}); ok {
			// Convert map back to domain.Post for DTO conversion
			domainPost := convertMapToPost(postMap)
			response.Posts[i] = dto.ToPostResponse(domainPost)
		} else {
			h.logger.Error("unexpected post type in feed result", zap.Any("post", post))
			return c.Status(500).JSON(fiber.Map{"error": "invalid post data"})
		}
	}
	return c.JSON(response)
}

// convertMapToPost converts a map[string]interface{} back to *domain.Post
func convertMapToPost(postMap map[string]interface{}) *domain.Post {
	post := &domain.Post{}

	if id, ok := postMap["id"].(float64); ok {
		post.ID = uint(id)
	}
	if userID, ok := postMap["user_id"].(float64); ok {
		post.UserID = uint(userID)
	}
	if username, ok := postMap["username"].(string); ok {
		post.Username = username
	}
	if title, ok := postMap["title"].(string); ok {
		post.Title = title
	}
	if caption, ok := postMap["caption"].(string); ok {
		post.Caption = caption
	}
	if mediaType, ok := postMap["media_type"].(string); ok {
		post.MediaType = domain.MediaType(mediaType)
	}
	if mediaURL, ok := postMap["media_url"].(string); ok {
		post.MediaURL = mediaURL
	}
	if likesCount, ok := postMap["likes_count"].(float64); ok {
		post.LikesCount = int(likesCount)
	}
	if commentsCount, ok := postMap["comments_count"].(float64); ok {
		post.CommentsCount = int(commentsCount)
	}
	if viewsCount, ok := postMap["views_count"].(float64); ok {
		post.ViewsCount = int(viewsCount)
	}
	if score, ok := postMap["score"].(float64); ok {
		post.Score = score
	}
	if createdAtStr, ok := postMap["created_at"].(string); ok {
		if createdAt, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
			post.CreatedAt = createdAt
		}
	}
	if updatedAtStr, ok := postMap["updated_at"].(string); ok {
		if updatedAt, err := time.Parse(time.RFC3339, updatedAtStr); err == nil {
			post.UpdatedAt = updatedAt
		}
	}

	return post
}
