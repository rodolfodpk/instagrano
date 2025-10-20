package tests

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	. "github.com/onsi/gomega"

	"github.com/rodolfodpk/instagrano/internal/config"
	"github.com/rodolfodpk/instagrano/internal/handler"
	postgresRepo "github.com/rodolfodpk/instagrano/internal/repository/postgres"
	"github.com/rodolfodpk/instagrano/internal/service"
)

func TestPostHandler_GetPost(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers PostgreSQL + Mock S3
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	postRepo := postgresRepo.NewPostRepository(containers.DB)
	mockStorage := NewMockMediaStorage()
	postService := service.NewPostService(postRepo, mockStorage)
	postHandler := handler.NewPostHandler(postService)

	// Given: A post exists
	user := createTestUser(t, containers.DB, "getpostuser", "getpost@example.com")
	post := createTestPost(t, containers.DB, user.ID, "Test Post", "Test Caption")

	// Create Fiber app
	app := fiber.New()
	app.Get("/posts/:id", postHandler.GetPost)

	// When: Get post by ID
	req := httptest.NewRequest("GET", fmt.Sprintf("/posts/%d", post.ID), nil)
	resp, err := app.Test(req)

	// Then: Should return the post
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(200))

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	Expect(result["id"]).To(Equal(float64(post.ID)))
	Expect(result["title"]).To(Equal("Test Post"))
	Expect(result["caption"]).To(Equal("Test Caption"))
}

func TestPostHandler_GetPostNotFound(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers PostgreSQL + Mock S3
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	postRepo := postgresRepo.NewPostRepository(containers.DB)
	mockStorage := NewMockMediaStorage()
	postService := service.NewPostService(postRepo, mockStorage)
	postHandler := handler.NewPostHandler(postService)

	// Create Fiber app
	app := fiber.New()
	app.Get("/posts/:id", postHandler.GetPost)

	// When: Get non-existent post
	req := httptest.NewRequest("GET", "/posts/9999", nil)
	resp, err := app.Test(req)

	// Then: Should return 404
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(404))
}

func TestPostHandler_GetPostInvalidID(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers PostgreSQL + Mock S3
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	postRepo := postgresRepo.NewPostRepository(containers.DB)
	mockStorage := NewMockMediaStorage()
	postService := service.NewPostService(postRepo, mockStorage)
	postHandler := handler.NewPostHandler(postService)

	// Create Fiber app
	app := fiber.New()
	app.Get("/posts/:id", postHandler.GetPost)

	// When: Get post with invalid ID
	req := httptest.NewRequest("GET", "/posts/invalid", nil)
	resp, err := app.Test(req)

	// Then: Should return 400 for invalid ID
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(400))
}

func TestFeedHandler_GetFeedWithPage(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers PostgreSQL + Redis
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	postRepo := postgresRepo.NewPostRepository(containers.DB)
	feedService := service.NewFeedService(postRepo, containers.Cache, 5*time.Minute)
	cfg := &config.Config{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}
	feedHandler := handler.NewFeedHandler(feedService, cfg)

	// Given: Posts exist
	user := createTestUser(t, containers.DB, "pageuser", "page@example.com")
	createTestPost(t, containers.DB, user.ID, "Post 1", "Caption 1")
	createTestPost(t, containers.DB, user.ID, "Post 2", "Caption 2")

	// Create Fiber app
	app := fiber.New()
	app.Get("/feed", feedHandler.GetFeed)

	// When: Get feed with page parameter
	req := httptest.NewRequest("GET", "/feed?page=1&limit=10", nil)
	resp, err := app.Test(req)

	// Then: Should return posts
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(200))

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	Expect(result).To(HaveKey("posts"))
}

func TestFeedHandler_GetFeedWithInvalidPage(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers PostgreSQL + Redis
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	postRepo := postgresRepo.NewPostRepository(containers.DB)
	feedService := service.NewFeedService(postRepo, containers.Cache, 5*time.Minute)
	cfg := &config.Config{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}
	feedHandler := handler.NewFeedHandler(feedService, cfg)

	// Create Fiber app
	app := fiber.New()
	app.Get("/feed", feedHandler.GetFeed)

	// When: Get feed with invalid page parameter (should fallback to defaults)
	req := httptest.NewRequest("GET", "/feed?page=invalid", nil)
	resp, err := app.Test(req)

	// Then: Should return 200 with default values (current behavior)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(200))
}

func TestFeedHandler_GetFeedWithInvalidLimit(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers PostgreSQL + Redis
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	postRepo := postgresRepo.NewPostRepository(containers.DB)
	feedService := service.NewFeedService(postRepo, containers.Cache, 5*time.Minute)
	cfg := &config.Config{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}
	feedHandler := handler.NewFeedHandler(feedService, cfg)

	// Create Fiber app
	app := fiber.New()
	app.Get("/feed", feedHandler.GetFeed)

	// When: Get feed with invalid limit parameter
	req := httptest.NewRequest("GET", "/feed?limit=invalid", nil)
	resp, err := app.Test(req)

	// Then: Should return 200 with default values (current behavior)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(200))
}

func TestFeedHandler_GetFeedWithExcessiveLimit(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers PostgreSQL + Redis
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	postRepo := postgresRepo.NewPostRepository(containers.DB)
	feedService := service.NewFeedService(postRepo, containers.Cache, 5*time.Minute)
	cfg := &config.Config{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}
	feedHandler := handler.NewFeedHandler(feedService, cfg)

	// Create Fiber app
	app := fiber.New()
	app.Get("/feed", feedHandler.GetFeed)

	// When: Get feed with limit exceeding max
	req := httptest.NewRequest("GET", "/feed?limit=1000", nil)
	resp, err := app.Test(req)

	// Then: Should return 200 with default values (current behavior)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(200))
}

func TestFeedHandler_GetFeedWithZeroLimit(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers PostgreSQL + Redis
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	postRepo := postgresRepo.NewPostRepository(containers.DB)
	feedService := service.NewFeedService(postRepo, containers.Cache, 5*time.Minute)
	cfg := &config.Config{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}
	feedHandler := handler.NewFeedHandler(feedService, cfg)

	// Create Fiber app
	app := fiber.New()
	app.Get("/feed", feedHandler.GetFeed)

	// When: Get feed with zero limit
	req := httptest.NewRequest("GET", "/feed?limit=0", nil)
	resp, err := app.Test(req)

	// Then: Should return 200 with default values (current behavior)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(200))
}

func TestFeedHandler_GetFeedWithNegativePage(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers PostgreSQL + Redis
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	postRepo := postgresRepo.NewPostRepository(containers.DB)
	feedService := service.NewFeedService(postRepo, containers.Cache, 5*time.Minute)
	cfg := &config.Config{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}
	feedHandler := handler.NewFeedHandler(feedService, cfg)

	// Create Fiber app
	app := fiber.New()
	app.Get("/feed", feedHandler.GetFeed)

	// When: Get feed with negative page
	req := httptest.NewRequest("GET", "/feed?page=-1", nil)
	resp, err := app.Test(req)

	// Then: Should return 200 with default values (current behavior)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(200))
}

func TestFeedHandler_GetFeedWithNegativeLimit(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers PostgreSQL + Redis
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	postRepo := postgresRepo.NewPostRepository(containers.DB)
	feedService := service.NewFeedService(postRepo, containers.Cache, 5*time.Minute)
	cfg := &config.Config{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}
	feedHandler := handler.NewFeedHandler(feedService, cfg)

	// Create Fiber app
	app := fiber.New()
	app.Get("/feed", feedHandler.GetFeed)

	// When: Get feed with negative limit
	req := httptest.NewRequest("GET", "/feed?limit=-1", nil)
	resp, err := app.Test(req)

	// Then: Should return 200 with default values (current behavior)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(200))
}

func TestFeedHandler_GetFeedWithPageParameter(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers PostgreSQL + Redis
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	postRepo := postgresRepo.NewPostRepository(containers.DB)
	feedService := service.NewFeedService(postRepo, containers.Cache, 5*time.Minute)
	cfg := &config.Config{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}
	feedHandler := handler.NewFeedHandler(feedService, cfg)

	// Given: Posts exist
	user := createTestUser(t, containers.DB, "pageparamuser", "pageparam@example.com")
	createTestPost(t, containers.DB, user.ID, "Post 1", "Caption 1")
	createTestPost(t, containers.DB, user.ID, "Post 2", "Caption 2")

	// Create Fiber app
	app := fiber.New()
	app.Get("/feed", feedHandler.GetFeed)

	// When: Get feed with page parameter (should trigger page-based pagination)
	req := httptest.NewRequest("GET", "/feed?page=1", nil)
	resp, err := app.Test(req)

	// Then: Should return posts
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(200))

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	Expect(result).To(HaveKey("posts"))
}
