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

    // Parse multipart form
    _, err := c.MultipartForm()
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "invalid form data"})
    }

    title := c.FormValue("title")
    caption := c.FormValue("caption")
    mediaTypeStr := c.FormValue("media_type")

    if title == "" {
        return c.Status(400).JSON(fiber.Map{"error": "title is required"})
    }

    mediaType := domain.MediaType(mediaTypeStr)
    if mediaType != domain.MediaTypeImage && mediaType != domain.MediaTypeVideo {
        mediaType = domain.MediaTypeImage // default
    }

    // Get uploaded file
    file, err := c.FormFile("media")
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "media file is required"})
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
