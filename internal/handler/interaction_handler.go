package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/rodolfodpk/instagrano/internal/dto"
	"github.com/rodolfodpk/instagrano/internal/events"
	"github.com/rodolfodpk/instagrano/internal/service"
	"go.uber.org/zap"
)

type InteractionHandler struct {
	interactionService *service.InteractionService
	eventPublisher     *events.Publisher
	logger             *zap.Logger
}

func NewInteractionHandler(interactionService *service.InteractionService, eventPublisher *events.Publisher, logger *zap.Logger) *InteractionHandler {
	return &InteractionHandler{
		interactionService: interactionService,
		eventPublisher:     eventPublisher,
		logger:             logger,
	}
}

// LikePost godoc
// @Summary      Like a post
// @Description  Add a like to a post
// @Tags         interactions
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Post ID"
// @Success      200  {object}  object{message=string}
// @Failure      400  {object}  object{error=string}
// @Router       /posts/{id}/like [post]
func (h *InteractionHandler) LikePost(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	postID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid post id"})
	}

	likesCount, commentsCount, err := h.interactionService.LikePost(userID, uint(postID))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Publish post liked event
	if err := h.eventPublisher.PublishPostLiked(c.Context(), uint(postID), userID, likesCount, commentsCount); err != nil {
		h.logger.Error("failed to publish post liked event",
			zap.Error(err),
			zap.Uint("post_id", uint(postID)),
			zap.Uint("user_id", userID))
	}

	// Get updated post to return like count
	post, err := h.interactionService.GetPost(uint(postID))
	if err != nil {
		h.logger.Error("failed to get updated post", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{"error": "failed to get updated post"})
	}

	return c.JSON(dto.LikeResponse{
		PostID:     uint(postID),
		LikesCount: post.LikesCount,
	})
}

// CommentPost godoc
// @Summary      Comment on a post
// @Description  Add a comment to a post
// @Tags         interactions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int                          true  "Post ID"
// @Param        request  body      object{text=string}          true  "Comment text"
// @Success      200  {object}  object{message=string}
// @Failure      400  {object}  object{error=string}
// @Router       /posts/{id}/comment [post]
func (h *InteractionHandler) CommentPost(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	username := c.Locals("username").(string)
	postID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid post id"})
	}

	var req dto.CommentRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	// CommentPost now handles publishing the event internally
	_, _, err = h.interactionService.CommentPost(userID, uint(postID), req.Content, username)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Get updated post to return comment count
	post, err := h.interactionService.GetPost(uint(postID))
	if err != nil {
		h.logger.Error("failed to get updated post", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{"error": "failed to get updated post"})
	}

	return c.JSON(dto.CommentResponse{
		PostID:        uint(postID),
		CommentsCount: post.CommentsCount,
	})
}

// GetComments godoc
// @Summary      Get comments for a post
// @Description  Retrieve all comments for a specific post
// @Tags         interactions
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Post ID"
// @Success      200  {array}   domain.Comment
// @Failure      400  {object}  object{error=string}
// @Router       /posts/{id}/comments [get]
func (h *InteractionHandler) GetComments(c *fiber.Ctx) error {
	postID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid post id"})
	}

	comments, err := h.interactionService.GetComments(uint(postID))
	if err != nil {
		h.logger.Error("failed to get comments", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{"error": "failed to get comments"})
	}

	return c.JSON(comments)
}
