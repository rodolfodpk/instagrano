package handler

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// TestImageHandler serves test images for development and testing
type TestImageHandler struct{}

func NewTestImageHandler() *TestImageHandler {
	return &TestImageHandler{}
}

// ServeTestImage godoc
// @Summary      Serve a test image
// @Description  Returns a generated test image with specified dimensions
// @Tags         test
// @Produce      image/jpeg
// @Param        width   query  int  false  "Image width (default: 150)"
// @Param        height  query  int  false  "Image height (default: 150)"
// @Param        color   query  string  false  "Background color (red, green, blue, gray) (default: gray)"
// @Success      200  {file}  binary  "Generated test image"
// @Router       /test/image [get]
func (h *TestImageHandler) ServeTestImage(c *fiber.Ctx) error {
	// Parse query parameters
	width := 150
	height := 150
	bgColor := "gray"

	if w := c.Query("width"); w != "" {
		if parsed, err := strconv.Atoi(w); err == nil && parsed > 0 && parsed <= 1000 {
			width = parsed
		}
	}

	if h := c.Query("height"); h != "" {
		if parsed, err := strconv.Atoi(h); err == nil && parsed > 0 && parsed <= 1000 {
			height = parsed
		}
	}

	if color := c.Query("color"); color != "" {
		bgColor = color
	}

	// Create image
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	// Set background color
	var bg color.RGBA
	switch bgColor {
	case "red":
		bg = color.RGBA{255, 0, 0, 255}
	case "green":
		bg = color.RGBA{0, 255, 0, 255}
	case "blue":
		bg = color.RGBA{0, 0, 255, 255}
	default:
		bg = color.RGBA{128, 128, 128, 255} // gray
	}

	// Fill image with background color
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, bg)
		}
	}

	// Encode as JPEG
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate image"})
	}

	// Set headers and return image
	c.Set("Content-Type", "image/jpeg")
	c.Set("Content-Length", strconv.Itoa(buf.Len()))
	c.Set("Cache-Control", "no-cache")
	
	return c.Send(buf.Bytes())
}

// ServeTestPNG godoc
// @Summary      Serve a test PNG image
// @Description  Returns a generated test PNG image with specified dimensions
// @Tags         test
// @Produce      image/png
// @Param        width   query  int  false  "Image width (default: 150)"
// @Param        height  query  int  false  "Image height (default: 150)"
// @Success      200  {file}  binary  "Generated test PNG image"
// @Router       /test/image/png [get]
func (h *TestImageHandler) ServeTestPNG(c *fiber.Ctx) error {
	// Parse query parameters
	width := 150
	height := 150

	if w := c.Query("width"); w != "" {
		if parsed, err := strconv.Atoi(w); err == nil && parsed > 0 && parsed <= 1000 {
			width = parsed
		}
	}

	if h := c.Query("height"); h != "" {
		if parsed, err := strconv.Atoi(h); err == nil && parsed > 0 && parsed <= 1000 {
			height = parsed
		}
	}

	// Create image
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	// Set background color (blue for PNG)
	bg := color.RGBA{0, 100, 200, 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, bg)
		}
	}

	// Encode as PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to generate PNG image"})
	}

	// Set headers and return image
	c.Set("Content-Type", "image/png")
	c.Set("Content-Length", strconv.Itoa(buf.Len()))
	c.Set("Cache-Control", "no-cache")
	
	return c.Send(buf.Bytes())
}
