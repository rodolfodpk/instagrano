package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rodolfodpk/instagrano/internal/service"
)

type PostViewHandler struct {
	viewService *service.PostViewService
}

func NewPostViewHandler(viewService *service.PostViewService) *PostViewHandler {
	return &PostViewHandler{
		viewService: viewService,
	}
}

// StartView godoc
// @Summary      Start tracking view time
// @Description  Begin tracking how long a user views a post
// @Tags         views
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Post ID"
// @Success      200  {object}  domain.PostView
// @Failure      400  {object}  object{error=string}
// @Router       /posts/{id}/view/start [post]
func (h *PostViewHandler) StartView(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	postID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid post id"})
	}

	view, err := h.viewService.StartView(userID, uint(postID))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(view)
}

// EndView godoc
// @Summary      End tracking view time
// @Description  Stop tracking and record view duration
// @Tags         views
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int                          true  "Post ID"
// @Param        request  body      object{started_at=string}    true  "View start timestamp"
// @Success      200  {object}  object{message=string}
// @Failure      400  {object}  object{error=string}
// @Router       /posts/{id}/view/end [post]
func (h *PostViewHandler) EndView(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	postID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid post id"})
	}

	var req struct {
		StartedAt string `json:"started_at"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	startedAt, err := time.Parse(time.RFC3339, req.StartedAt)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid started_at format"})
	}

	endedAt := time.Now()

	if err := h.viewService.EndView(userID, uint(postID), startedAt, endedAt); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "view ended"})
}
