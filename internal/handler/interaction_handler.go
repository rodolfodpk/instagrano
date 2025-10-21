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

	err = h.interactionService.LikePost(userID, uint(postID))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "post liked"})
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
