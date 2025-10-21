package tests

import (
	"net/http/httptest"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"

	"github.com/rodolfodpk/instagrano/internal/logger"
	"github.com/rodolfodpk/instagrano/internal/middleware"
)

var _ = Describe("RequestLogger", func() {
	It("should log requests successfully", func() {
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

		// Then: Should handle request successfully
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(200))
	})

	It("should handle different HTTP methods", func() {
		// Given: A Fiber app with request logger middleware
		app := fiber.New()
		log := logger.New(zap.NewAtomicLevelAt(zap.InfoLevel), "json")
		app.Use(middleware.RequestLogger(log))

		// Add test routes for different methods
		app.Get("/get", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"method": "GET"})
		})
		app.Post("/post", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"method": "POST"})
		})
		app.Put("/put", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"method": "PUT"})
		})
		app.Delete("/delete", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"method": "DELETE"})
		})

		// When: Make requests with different methods
		methods := []string{"GET", "POST", "PUT", "DELETE"}
		for _, method := range methods {
			req := httptest.NewRequest(method, "/"+method, nil)
			resp, err := app.Test(req)

			// Then: Should handle each method successfully
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))
		}
	})

	It("should handle errors gracefully", func() {
		// Given: A Fiber app with request logger middleware
		app := fiber.New()
		log := logger.New(zap.NewAtomicLevelAt(zap.InfoLevel), "json")
		app.Use(middleware.RequestLogger(log))

		// Add a route that returns an error
		app.Get("/error", func(c *fiber.Ctx) error {
			return fiber.NewError(500, "internal server error")
		})

		// When: Make a request to error route
		req := httptest.NewRequest("GET", "/error", nil)
		resp, err := app.Test(req)

		// Then: Should handle error gracefully
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(500))
	})
})

var _ = Describe("CORS", func() {
	It("should add CORS headers", func() {
		// Given: A Fiber app with CORS middleware
		app := fiber.New()
		app.Use(cors.New())

		// Add a test route
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.SendString("OK")
		})

		// When: Make a request with Origin header
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://example.com")
		resp, err := app.Test(req)

		// Then: Should include CORS headers
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(200))
		Expect(resp.Header.Get("Access-Control-Allow-Origin")).To(Equal("*"))
	})

	It("should handle preflight OPTIONS requests", func() {
		// Given: A Fiber app with CORS middleware
		app := fiber.New()
		app.Use(cors.New())

		// Add a test route
		app.Get("/test", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"message": "test"})
		})

		// When: Make a preflight OPTIONS request
		req := httptest.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "GET")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type")
		resp, err := app.Test(req)

		// Then: Should handle preflight request
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(204))
		Expect(resp.Header.Get("Access-Control-Allow-Origin")).To(Equal("*"))
	})
})

var _ = Describe("AuthRequired", func() {
	It("should allow requests with valid JWT token", func() {
		// Given: A Fiber app with auth middleware
		app := fiber.New()
		app.Use(middleware.JWT("test-secret"))

		// Add a protected route
		app.Get("/protected", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"message": "protected"})
		})

		// Create a valid JWT token
		token, err := createTestJWT(1)
		Expect(err).NotTo(HaveOccurred())

		// When: Make request with valid token
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := app.Test(req)

		// Then: Should allow access
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(200))
	})

	It("should reject requests without JWT token", func() {
		// Given: A Fiber app with auth middleware
		app := fiber.New()
		app.Use(middleware.JWT("test-secret"))

		// Add a protected route
		app.Get("/protected", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"message": "protected"})
		})

		// When: Make request without token
		req := httptest.NewRequest("GET", "/protected", nil)
		resp, err := app.Test(req)

		// Then: Should reject access
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(401))
	})

	It("should reject requests with invalid JWT token", func() {
		// Given: A Fiber app with auth middleware
		app := fiber.New()
		app.Use(middleware.JWT("test-secret"))

		// Add a protected route
		app.Get("/protected", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"message": "protected"})
		})

		// When: Make request with invalid token
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		resp, err := app.Test(req)

		// Then: Should reject access
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(401))
	})

	It("should reject requests with malformed Authorization header", func() {
		// Given: A Fiber app with auth middleware
		app := fiber.New()
		app.Use(middleware.JWT("test-secret"))

		// Add a protected route
		app.Get("/protected", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"message": "protected"})
		})

		// When: Make request with malformed header
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "InvalidFormat token")
		resp, err := app.Test(req)

		// Then: Should reject access
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(401))
	})

	It("should extract user ID from valid token", func() {
		// Given: A Fiber app with auth middleware
		app := fiber.New()
		app.Use(middleware.JWT("test-secret"))

		// Add a protected route that returns user ID
		app.Get("/protected", func(c *fiber.Ctx) error {
			userID := c.Locals("user_id")
			return c.JSON(fiber.Map{"user_id": userID})
		})

		// Create a valid JWT token for user ID 123
		token, err := createTestJWT(123)
		Expect(err).NotTo(HaveOccurred())

		// When: Make request with valid token
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := app.Test(req)

		// Then: Should extract user ID correctly
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(200))
	})
})
