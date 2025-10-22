package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"

	"github.com/rodolfodpk/instagrano/internal/config"
	"github.com/rodolfodpk/instagrano/internal/events"
	"github.com/rodolfodpk/instagrano/internal/handler"
	"github.com/rodolfodpk/instagrano/internal/middleware"
	postgresRepo "github.com/rodolfodpk/instagrano/internal/repository/postgres"
	"github.com/rodolfodpk/instagrano/internal/repository/s3"
	"github.com/rodolfodpk/instagrano/internal/service"
	"github.com/rodolfodpk/instagrano/internal/webclient"
)

var _ = Describe("PostHandler", func() {
	Describe("GetPost", func() {
		It("should return post by ID", func() {
			// Given: A post exists in database
			user := createTestUser(sharedContainers.DB, "testuser", "test@example.com")
			post := createTestPost(sharedContainers.DB, user.ID, "Test Post", "Test Caption")

			postRepo := postgresRepo.NewPostRepository(sharedContainers.DB)
			mediaStorage := createTestS3Storage()
			postService := service.NewPostService(postRepo, mediaStorage, sharedContainers.Cache, 5*time.Minute)

			logger, _ := zap.NewProduction()
			defer logger.Sync()
			eventPublisher := events.NewPublisher(sharedContainers.Cache, logger)
			postHandler := handler.NewPostHandler(postService, eventPublisher, logger)

			// Create Fiber app
			app := fiber.New()
			app.Get("/posts/:id", postHandler.GetPost)

			// When: Request post by ID
			req := httptest.NewRequest("GET", fmt.Sprintf("/posts/%d", post.ID), nil)
			resp, err := app.Test(req)

			// Then: Should return post successfully
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))

			var response map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&response)
			Expect(response["id"]).To(Equal(float64(post.ID)))
			Expect(response["title"]).To(Equal("Test Post"))
		})

		It("should return 404 for non-existent post", func() {
			// Given: Post handler
			postRepo := postgresRepo.NewPostRepository(sharedContainers.DB)
			mediaStorage := createTestS3Storage()
			postService := service.NewPostService(postRepo, mediaStorage, sharedContainers.Cache, 5*time.Minute)

			logger, _ := zap.NewProduction()
			defer logger.Sync()
			eventPublisher := events.NewPublisher(sharedContainers.Cache, logger)
			postHandler := handler.NewPostHandler(postService, eventPublisher, logger)

			// Create Fiber app
			app := fiber.New()
			app.Get("/posts/:id", postHandler.GetPost)

			// When: Request non-existent post
			req := httptest.NewRequest("GET", "/posts/99999", nil)
			resp, err := app.Test(req)

			// Then: Should return 404
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(404))
		})

		It("should return 400 for invalid post ID", func() {
			// Given: Post handler
			postRepo := postgresRepo.NewPostRepository(sharedContainers.DB)
			mediaStorage := createTestS3Storage()
			postService := service.NewPostService(postRepo, mediaStorage, sharedContainers.Cache, 5*time.Minute)

			logger, _ := zap.NewProduction()
			defer logger.Sync()
			eventPublisher := events.NewPublisher(sharedContainers.Cache, logger)
			postHandler := handler.NewPostHandler(postService, eventPublisher, logger)

			// Create Fiber app
			app := fiber.New()
			app.Get("/posts/:id", postHandler.GetPost)

			// When: Request with invalid ID
			req := httptest.NewRequest("GET", "/posts/invalid", nil)
			resp, err := app.Test(req)

			// Then: Should return 400
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(400))
		})
	})

	Describe("CreatePost", func() {
		It("should create post successfully", func() {
			// Given: User exists and post handler
			user := createTestUser(sharedContainers.DB, "testuser", "test@example.com")
			token, err := createTestJWT(user.ID)
			Expect(err).NotTo(HaveOccurred())

			postRepo := postgresRepo.NewPostRepository(sharedContainers.DB)

			// Create S3 storage with mock controller enabled
			cfg := &config.Config{
				S3Endpoint: sharedContainers.S3Endpoint,
				S3Region:   "us-east-1",
				S3Bucket:   "test-bucket",
			}
			webclientConfig := webclient.Config{
				UseMockController: true,
				MockBaseURL:       "http://localhost:8080",
				RealURLTimeout:    cfg.WebclientTimeout,
			}
			mediaStorage, err := s3.NewMediaStorage(
				cfg.S3Endpoint,
				cfg.S3Region,
				cfg.S3Bucket,
				webclientConfig,
			)
			Expect(err).NotTo(HaveOccurred())

			// Create S3 bucket if it doesn't exist
			err = mediaStorage.CreateBucketIfNotExists()
			Expect(err).NotTo(HaveOccurred())

			postService := service.NewPostService(postRepo, mediaStorage, sharedContainers.Cache, 5*time.Minute)

			logger, _ := zap.NewProduction()
			defer logger.Sync()
			eventPublisher := events.NewPublisher(sharedContainers.Cache, logger)
			postHandler := handler.NewPostHandler(postService, eventPublisher, logger)

			// Create Fiber app with auth middleware
			app := fiber.New()
			appCfg := &config.Config{JWTSecret: "test-secret"}
			app.Use(middleware.JWT(appCfg.JWTSecret))
			app.Post("/posts", postHandler.CreatePost)

			// When: Create post
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			writer.WriteField("title", "New Post")
			writer.WriteField("caption", "New Caption")
			writer.WriteField("media_url", "https://via.placeholder.com/300x200/FF0000/FFFFFF?text=Test")

			writer.Close()

			req := httptest.NewRequest("POST", "/posts", &buf)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req.Header.Set("Authorization", "Bearer "+token)
			resp, err := app.Test(req)

			// Then: Should create post successfully
			Expect(err).NotTo(HaveOccurred())

			// Debug: Check response body if status is not 201
			if resp.StatusCode != 201 {
				var errorResponse map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&errorResponse)
				fmt.Printf("Error response: %+v\n", errorResponse)
			}

			Expect(resp.StatusCode).To(Equal(201))

			var response map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&response)
			Expect(response["title"]).To(Equal("New Post"))
			Expect(response["caption"]).To(Equal("New Caption"))
		})

		It("should return 401 without authentication", func() {
			// Given: Post handler
			postRepo := postgresRepo.NewPostRepository(sharedContainers.DB)
			mediaStorage := createTestS3Storage()
			postService := service.NewPostService(postRepo, mediaStorage, sharedContainers.Cache, 5*time.Minute)

			logger, _ := zap.NewProduction()
			defer logger.Sync()
			eventPublisher := events.NewPublisher(sharedContainers.Cache, logger)
			postHandler := handler.NewPostHandler(postService, eventPublisher, logger)

			// Create Fiber app with auth middleware
			app := fiber.New()
			appCfg := &config.Config{JWTSecret: "test-secret"}
			app.Use(middleware.JWT(appCfg.JWTSecret))
			app.Post("/posts", postHandler.CreatePost)

			// When: Create post without authentication
			postData := map[string]string{
				"title":     "New Post",
				"caption":   "New Caption",
				"media_url": "https://example.com/image.jpg",
			}
			req := httptest.NewRequest("POST", "/posts", strings.NewReader(marshalJSON(postData)))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)

			// Then: Should return 401
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(401))
		})

		It("should return 400 for invalid post data", func() {
			// Given: User exists and post handler
			user := createTestUser(sharedContainers.DB, "testuser", "test@example.com")
			token, err := createTestJWT(user.ID)
			Expect(err).NotTo(HaveOccurred())

			postRepo := postgresRepo.NewPostRepository(sharedContainers.DB)
			mediaStorage := createTestS3Storage()
			postService := service.NewPostService(postRepo, mediaStorage, sharedContainers.Cache, 5*time.Minute)

			logger, _ := zap.NewProduction()
			defer logger.Sync()
			eventPublisher := events.NewPublisher(sharedContainers.Cache, logger)
			postHandler := handler.NewPostHandler(postService, eventPublisher, logger)

			// Create Fiber app with auth middleware
			app := fiber.New()
			appCfg := &config.Config{JWTSecret: "test-secret"}
			app.Use(middleware.JWT(appCfg.JWTSecret))
			app.Post("/posts", postHandler.CreatePost)

			// When: Create post with invalid data
			postData := map[string]string{
				"title": "", // Empty title should be invalid
			}
			req := httptest.NewRequest("POST", "/posts", strings.NewReader(marshalJSON(postData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			resp, err := app.Test(req)

			// Then: Should return 400
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(400))
		})
	})
})
