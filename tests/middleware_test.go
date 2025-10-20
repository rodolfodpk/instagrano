package tests

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"

	"github.com/rodolfodpk/instagrano/internal/logger"
	"github.com/rodolfodpk/instagrano/internal/middleware"
)

func TestRequestLogger(t *testing.T) {
	RegisterTestingT(t)

	// Given: A Fiber app with request logger middleware
	app := fiber.New()
	log := logger.New(zap.NewAtomicLevelAt(zap.InfoLevel), "json")
	app.Use(middleware.RequestLogger(log))

	// Add a test route
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "test"})
	})

	// When: Make a request
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	// Then: Request should be processed successfully
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(200))
}

func TestRequestLogger_WithUserAgent(t *testing.T) {
	RegisterTestingT(t)

	// Given: A Fiber app with request logger middleware
	app := fiber.New()
	log := logger.New(zap.NewAtomicLevelAt(zap.InfoLevel), "json")
	app.Use(middleware.RequestLogger(log))

	// Add a test route
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "test"})
	})

	// When: Make a request with User-Agent header
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "TestAgent/1.0")
	resp, err := app.Test(req)

	// Then: Request should be processed successfully
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(200))
}

func TestRequestLogger_WithXForwardedFor(t *testing.T) {
	RegisterTestingT(t)

	// Given: A Fiber app with request logger middleware
	app := fiber.New()
	log := logger.New(zap.NewAtomicLevelAt(zap.InfoLevel), "json")
	app.Use(middleware.RequestLogger(log))

	// Add a test route
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "test"})
	})

	// When: Make a request with X-Forwarded-For header
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	resp, err := app.Test(req)

	// Then: Request should be processed successfully
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(200))
}

func TestRequestLogger_PostRequest(t *testing.T) {
	RegisterTestingT(t)

	// Given: A Fiber app with request logger middleware
	app := fiber.New()
	log := logger.New(zap.NewAtomicLevelAt(zap.InfoLevel), "json")
	app.Use(middleware.RequestLogger(log))

	// Add a test route
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "test"})
	})

	// When: Make a POST request
	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	// Then: Request should be processed successfully
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(200))
}

func TestRequestLogger_ErrorResponse(t *testing.T) {
	RegisterTestingT(t)

	// Given: A Fiber app with request logger middleware
	app := fiber.New()
	log := logger.New(zap.NewAtomicLevelAt(zap.InfoLevel), "json")
	app.Use(middleware.RequestLogger(log))

	// Add a test route that returns an error
	app.Get("/error", func(c *fiber.Ctx) error {
		return c.Status(500).JSON(fiber.Map{"error": "test error"})
	})

	// When: Make a request to error route
	req := httptest.NewRequest("GET", "/error", nil)
	resp, err := app.Test(req)

	// Then: Request should be processed successfully (error is handled by route)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(500))
}

func TestRequestLogger_NotFound(t *testing.T) {
	RegisterTestingT(t)

	// Given: A Fiber app with request logger middleware
	app := fiber.New()
	log := logger.New(zap.NewAtomicLevelAt(zap.InfoLevel), "json")
	app.Use(middleware.RequestLogger(log))

	// When: Make a request to non-existent route
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	resp, err := app.Test(req)

	// Then: Request should return 404
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(404))
}

func TestGenerateRequestID(t *testing.T) {
	RegisterTestingT(t)

	// When: Generate multiple request IDs (we can't test the private function directly,
	// but we can test that the middleware works correctly)
	app := fiber.New()
	log := logger.New(zap.NewAtomicLevelAt(zap.InfoLevel), "json")
	app.Use(middleware.RequestLogger(log))

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "test"})
	})

	// Make multiple requests to test request ID generation
	req1 := httptest.NewRequest("GET", "/test", nil)
	resp1, err := app.Test(req1)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp1.StatusCode).To(Equal(200))

	req2 := httptest.NewRequest("GET", "/test", nil)
	resp2, err := app.Test(req2)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp2.StatusCode).To(Equal(200))

	// Both requests should succeed (request IDs are generated internally)
}
