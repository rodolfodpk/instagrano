package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/rodolfodpk/instagrano/internal/domain"
	"github.com/rodolfodpk/instagrano/internal/events"
	"github.com/rodolfodpk/instagrano/internal/service"
	"go.uber.org/zap"
)

type PostHandler struct {
	postService    *service.PostService
	eventPublisher *events.Publisher
	logger         *zap.Logger
}

func NewPostHandler(postService *service.PostService, eventPublisher *events.Publisher, logger *zap.Logger) *PostHandler {
	return &PostHandler{
		postService:    postService,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// CreatePost godoc
// @Summary      Create a new post
// @Description  Upload image/video post with title and caption (file upload or URL)
// @Tags         posts
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        title       formData  string  true   "Post title"
// @Param        caption     formData  string  false  "Post caption"
// @Param        media_type  formData  string  false  "Media type (image or video) - required for file upload"
// @Param        media       formData  file    false  "Media file (alternative to media_url)"
// @Param        media_url   formData  string  false  "Media URL (alternative to file upload)"
// @Success      201  {object}  domain.Post
// @Failure      400  {object}  object{error=string}
// @Router       /posts [post]
func (h *PostHandler) CreatePost(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	// Parse multipart form
	_, err := c.MultipartForm()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid form data"})
	}

	title := c.FormValue("title")
	caption := c.FormValue("caption")
	mediaURL := c.FormValue("media_url") // NEW: URL input

	if title == "" {
		return c.Status(400).JSON(fiber.Map{"error": "title is required"})
	}

	// Check if URL is provided
	if mediaURL != "" {
		post, err := h.postService.CreatePostFromURL(userID, title, caption, mediaURL)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		// Publish new post event
		if err := h.eventPublisher.PublishNewPost(c.Context(), post.ID, userID, post); err != nil {
			h.logger.Error("failed to publish new post event",
				zap.Error(err),
				zap.Uint("post_id", post.ID),
				zap.Uint("user_id", userID))
			// Don't fail the request if event publishing fails
		}

		return c.Status(201).JSON(post)
	}

	// Otherwise, handle file upload (existing logic)
	mediaTypeStr := c.FormValue("media_type")
	mediaType := domain.MediaType(mediaTypeStr)
	if mediaType != domain.MediaTypeImage && mediaType != domain.MediaTypeVideo {
		mediaType = domain.MediaTypeImage // default
	}

	// Get uploaded file
	file, err := c.FormFile("media")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "either media file or media_url is required"})
	}

	// Open file
	fileReader, err := file.Open()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "failed to open file"})
	}
	defer fileReader.Close()

	post, err := h.postService.CreatePost(userID, title, caption, mediaType, fileReader, file.Filename)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Publish new post event
	if err := h.eventPublisher.PublishNewPost(c.Context(), post.ID, userID, post); err != nil {
		h.logger.Error("failed to publish new post event",
			zap.Error(err),
			zap.Uint("post_id", post.ID),
			zap.Uint("user_id", userID))
		// Don't fail the request if event publishing fails
	}

	return c.Status(201).JSON(post)
}

// GetPost godoc
// @Summary      Get post by ID
// @Description  Retrieve a single post with all details
// @Tags         posts
// @Produce      json
// @Param        id   path      int  true  "Post ID"
// @Success      200  {object}  domain.Post
// @Failure      400  {object}  object{error=string}
// @Failure      404  {object}  object{error=string}
// @Router       /posts/{id} [get]
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
