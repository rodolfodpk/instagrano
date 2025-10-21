package tests

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/rodolfodpk/instagrano/internal/cache"
	"github.com/rodolfodpk/instagrano/internal/config"
	"github.com/rodolfodpk/instagrano/internal/domain"
	"github.com/rodolfodpk/instagrano/internal/events"
	"github.com/rodolfodpk/instagrano/internal/handler"
	"github.com/rodolfodpk/instagrano/internal/middleware"
	postgresRepo "github.com/rodolfodpk/instagrano/internal/repository/postgres"
	"github.com/rodolfodpk/instagrano/internal/service"
	"go.uber.org/zap"
)

// TestContainers holds container instances
type TestContainers struct {
	PostgresContainer *postgres.PostgresContainer
	RedisContainer    *redis.RedisContainer
	DB                *sql.DB
	Cache             cache.Cache
}

// Global shared containers for Ginkgo
var (
	sharedContainers *TestContainers
	ctx              context.Context
	cancel           context.CancelFunc
)

// Ginkgo test suite setup
var _ = BeforeSuite(func() {
	// Create context with timeout for test setup
	ctx, cancel = context.WithTimeout(context.Background(), 120*time.Second)

	// Start PostgreSQL container
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("instagrano"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	Expect(err).NotTo(HaveOccurred())

	// Get PostgreSQL connection string
	dbURL, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	Expect(err).NotTo(HaveOccurred())

	// Connect to PostgreSQL
	db, err := postgresRepo.Connect(dbURL)
	Expect(err).NotTo(HaveOccurred())

	// Run migrations
	runSharedMigrations(db)

	// Start Redis container
	redisContainer, err := redis.RunContainer(ctx,
		testcontainers.WithImage("redis:7-alpine"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(10*time.Second),
		),
	)
	Expect(err).NotTo(HaveOccurred())

	// Get Redis connection string
	redisAddr, err := redisContainer.ConnectionString(ctx)
	Expect(err).NotTo(HaveOccurred())

	// Remove redis:// prefix if present
	redisAddr = strings.TrimPrefix(redisAddr, "redis://")

	// Create Redis cache client
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	redisCache, err := cache.NewRedisCache(redisAddr, "", 0, logger)
	Expect(err).NotTo(HaveOccurred())

	sharedContainers = &TestContainers{
		PostgresContainer: pgContainer,
		RedisContainer:    redisContainer,
		DB:                db,
		Cache:             redisCache,
	}
})

var _ = BeforeEach(func() {
	// CRITICAL: Clean state between EVERY test
	truncateAllTables(sharedContainers.DB)
	sharedContainers.Cache.FlushAll(ctx)
})

var _ = AfterSuite(func() {
	if cancel != nil {
		cancel()
	}
	if sharedContainers != nil {
		if sharedContainers.Cache != nil {
			sharedContainers.Cache.Close()
		}
		if sharedContainers.DB != nil {
			sharedContainers.DB.Close()
		}
		if sharedContainers.PostgresContainer != nil {
			sharedContainers.PostgresContainer.Terminate(context.Background())
		}
		if sharedContainers.RedisContainer != nil {
			sharedContainers.RedisContainer.Terminate(context.Background())
		}
	}
})

// runSharedMigrations applies all SQL migrations (for shared containers)
func runSharedMigrations(db *sql.DB) {
	migrations := []string{
		"../migrations/001_create_users.up.sql",
		"../migrations/002_create_posts.up.sql",
		"../migrations/003_create_likes.up.sql",
		"../migrations/004_create_comments.up.sql",
		"../migrations/005_create_post_views.up.sql",
		"../migrations/006_optimize_indexes.up.sql",
	}

	for _, migration := range migrations {
		sql, err := os.ReadFile(migration)
		if err != nil {
			panic("failed to read migration " + migration + ": " + err.Error())
		}

		_, err = db.Exec(string(sql))
		if err != nil {
			panic("failed to run migration " + migration + ": " + err.Error())
		}
	}
}

// truncateAllTables cleans all tables between tests
func truncateAllTables(db *sql.DB) {
	ctx := context.Background()

	// Truncate tables in order to respect foreign key constraints
	tables := []string{
		"post_views", // Delete in order to respect foreign keys
		"comments",
		"likes",
		"posts",
		"users",
	}

	for _, table := range tables {
		_, err := db.ExecContext(ctx, "TRUNCATE TABLE "+table+" CASCADE")
		Expect(err).NotTo(HaveOccurred())
	}

	// Reset sequences
	sequences := []string{
		"users_id_seq",
		"posts_id_seq",
		"likes_id_seq",
		"comments_id_seq",
		"post_views_id_seq",
	}

	for _, seq := range sequences {
		_, err := db.ExecContext(ctx, "ALTER SEQUENCE "+seq+" RESTART WITH 1")
		Expect(err).NotTo(HaveOccurred())
	}
}

// setupTestContainers returns shared containers with truncate cleanup
func setupTestContainers(t *testing.T) (*TestContainers, func()) {
	RegisterTestingT(t)

	// Return shared containers with truncate cleanup
	cleanup := func() {
		truncateAllTables(sharedContainers.DB)

		// Flush Redis cache
		if sharedContainers.Cache != nil {
			ctx := context.Background()
			if err := sharedContainers.Cache.FlushAll(ctx); err != nil {
				// Log error but don't fail the test
				fmt.Printf("Warning: failed to flush Redis cache: %v\n", err)
			}
		}
	}

	return sharedContainers, cleanup
}

// runMigrations applies all SQL migrations
func runMigrations(t *testing.T, db *sql.DB) {
	RegisterTestingT(t)

	migrations := []string{
		"../migrations/001_create_users.up.sql",
		"../migrations/002_create_posts.up.sql",
		"../migrations/003_create_likes.up.sql",
		"../migrations/004_create_comments.up.sql",
		"../migrations/005_create_post_views.up.sql",
		"../migrations/006_optimize_indexes.up.sql",
	}

	for _, migration := range migrations {
		sql, err := os.ReadFile(migration)
		Expect(err).NotTo(HaveOccurred())

		_, err = db.Exec(string(sql))
		Expect(err).NotTo(HaveOccurred())
	}
}

// setupTestApp creates Fiber app with shared Testcontainers dependencies
func setupTestApp() (*fiber.App, *TestContainers, func()) {
	cleanup := func() {
		truncateAllTables(sharedContainers.DB)

		// Flush Redis cache
		if sharedContainers.Cache != nil {
			ctx := context.Background()
			if err := sharedContainers.Cache.FlushAll(ctx); err != nil {
				// Log error but don't fail the test
				fmt.Printf("Warning: failed to flush Redis cache: %v\n", err)
			}
		}
	}

	cfg := &config.Config{
		JWTSecret:  "test-secret",
		CacheTTL:   5 * time.Minute,
		S3Endpoint: "http://localhost:4566",
		S3Bucket:   "test-bucket",
	}

	// Initialize repositories
	userRepo := postgresRepo.NewUserRepository(sharedContainers.DB)
	postRepo := postgresRepo.NewPostRepository(sharedContainers.DB)
	likeRepo := postgresRepo.NewLikeRepository(sharedContainers.DB)
	commentRepo := postgresRepo.NewCommentRepository(sharedContainers.DB)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	feedService := service.NewFeedService(postRepo, sharedContainers.Cache, cfg.CacheTTL)
	interactionService := service.NewInteractionService(likeRepo, commentRepo, postRepo, sharedContainers.Cache)

	// Create mock S3 storage for testing
	mockStorage := NewMockMediaStorage()
	postService := service.NewPostService(postRepo, mockStorage, sharedContainers.Cache, cfg.CacheTTL)

	// Initialize event publisher
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	eventPublisher := events.NewPublisher(sharedContainers.Cache, logger)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	feedHandler := handler.NewFeedHandler(feedService, cfg)
	postHandler := handler.NewPostHandler(postService, eventPublisher, logger)
	interactionHandler := handler.NewInteractionHandler(interactionService, eventPublisher, logger)
	sseHandler := handler.NewSSEHandler(sharedContainers.Cache, logger, cfg.JWTSecret)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// Setup routes
	api := app.Group("/api")
	api.Post("/auth/register", authHandler.Register)
	api.Post("/auth/login", authHandler.Login)

	// SSE endpoint handles its own authentication via query parameter
	api.Get("/events/stream", sseHandler.Stream)

	// Serve static test image
	app.Static("/test/image", "./web/public/test-image.jpg")

	protected := api.Group("/", middleware.JWT(cfg.JWTSecret))
	protected.Get("/feed", feedHandler.GetFeed)
	protected.Post("/posts", postHandler.CreatePost)
	protected.Get("/posts/:id", postHandler.GetPost)
	protected.Post("/posts/:id/like", interactionHandler.LikePost)
	protected.Post("/posts/:id/comment", interactionHandler.CommentPost)

	return app, sharedContainers, cleanup
}

// Helper functions for creating test data
func createTestUser(db *sql.DB, username, email string) *domain.User {
	query := `INSERT INTO users (username, email, password, created_at, updated_at) 
			  VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id`

	hashedPassword := "$2a$10$testhash" // Mock bcrypt hash for testing
	var userID uint
	err := db.QueryRow(query, username, email, hashedPassword).Scan(&userID)
	Expect(err).NotTo(HaveOccurred())

	return &domain.User{
		ID:       userID,
		Username: username,
		Email:    email,
		Password: hashedPassword,
	}
}

func createTestPost(db *sql.DB, userID uint, title, caption string) *domain.Post {
	query := `INSERT INTO posts (user_id, title, caption, media_type, media_url, likes_count, comments_count, views_count, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()) RETURNING id`

	var postID uint
	err := db.QueryRow(query, userID, title, caption, "image", "/uploads/test.jpg", 0, 0, 0).Scan(&postID)
	Expect(err).NotTo(HaveOccurred())

	return &domain.Post{
		ID:            postID,
		UserID:        userID,
		Title:         title,
		Caption:       caption,
		MediaType:     "image",
		MediaURL:      "/uploads/test.jpg",
		LikesCount:    0,
		CommentsCount: 0,
		ViewsCount:    0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func createTestPostWithEngagement(t *testing.T, db *sql.DB, userID uint, title string, createdAt time.Time, likes, comments, views int) *domain.Post {
	RegisterTestingT(t)

	query := `INSERT INTO posts (user_id, title, caption, media_type, media_url, likes_count, comments_count, views_count, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW()) RETURNING id`

	var postID uint
	err := db.QueryRow(query, userID, title, "Test caption", "image", "/uploads/test.jpg", likes, comments, views, createdAt).Scan(&postID)
	Expect(err).NotTo(HaveOccurred())

	return &domain.Post{
		ID:            postID,
		UserID:        userID,
		Title:         title,
		Caption:       "Test caption",
		MediaType:     "image",
		MediaURL:      "/uploads/test.jpg",
		LikesCount:    likes,
		CommentsCount: comments,
		ViewsCount:    views,
		CreatedAt:     createdAt,
		UpdatedAt:     time.Now(),
	}
}

// registerAndLogin is a helper for integration tests
func registerAndLogin(app *fiber.App, username, email, password string) string {
	// Register user
	regData := map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	}

	regReq := httptest.NewRequest("POST", "/api/auth/register", strings.NewReader(marshalJSON(regData)))
	regReq.Header.Set("Content-Type", "application/json")
	regResp, err := app.Test(regReq)
	Expect(err).NotTo(HaveOccurred())
	Expect(regResp.StatusCode).To(Equal(200))

	// Login
	loginData := map[string]string{
		"email":    email,
		"password": password,
	}

	loginReq := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(marshalJSON(loginData)))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp, err := app.Test(loginReq)
	Expect(err).NotTo(HaveOccurred())
	Expect(loginResp.StatusCode).To(Equal(200))

	var loginResult map[string]interface{}
	json.NewDecoder(loginResp.Body).Decode(&loginResult)
	Expect(loginResult).To(HaveKey("token"))

	return loginResult["token"].(string)
}

// SSE Helper Functions

// SSEEvent represents a parsed SSE event
type SSEEvent struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// connectSSE creates an SSE connection and returns a channel for receiving events
func connectSSE(app *fiber.App, token string) (<-chan SSEEvent, func()) {
	eventCh := make(chan SSEEvent, 10)
	done := make(chan bool)

	// Create SSE request
	req := httptest.NewRequest("GET", "/api/events/stream?token="+token, nil)

	// Start SSE connection in goroutine
	go func() {
		defer close(eventCh)

		resp, err := app.Test(req, -1) // -1 means no timeout
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(200))

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			select {
			case <-done:
				return
			default:
				line := scanner.Text()
				if strings.HasPrefix(line, "event: ") {
					eventType := strings.TrimPrefix(line, "event: ")
					if scanner.Scan() {
						dataLine := scanner.Text()
						if strings.HasPrefix(dataLine, "data: ") {
							data := strings.TrimPrefix(dataLine, "data: ")

							var eventData json.RawMessage
							if err := json.Unmarshal([]byte(data), &eventData); err == nil {
								eventCh <- SSEEvent{
									Type: eventType,
									Data: eventData,
								}
							}
						}
					}
				}
			}
		}
	}()

	// Return cleanup function
	cleanup := func() {
		close(done)
	}

	return eventCh, cleanup
}

// waitForSSEEvent waits for a specific event type with timeout
func waitForSSEEvent(eventCh <-chan SSEEvent, eventType string, timeout time.Duration) SSEEvent {
	select {
	case event := <-eventCh:
		Expect(event.Type).To(Equal(eventType), fmt.Sprintf("Expected event type %s, got %s", eventType, event.Type))
		return event
	case <-time.After(timeout):
		Fail(fmt.Sprintf("timeout waiting for SSE event type: %s", eventType))
		return SSEEvent{}
	}
}

// parseSSEEventData parses SSE event data into the expected Event struct
func parseSSEEventData(data json.RawMessage) events.Event {
	var event events.Event
	err := json.Unmarshal(data, &event)
	Expect(err).NotTo(HaveOccurred())
	return event
}

// createTestPostWithSSE creates a post and returns the post data for SSE testing
func createTestPostWithSSE(app *fiber.App, token, title, caption string) map[string]interface{} {
	// Create multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	writer.WriteField("title", title)
	writer.WriteField("caption", caption)
	writer.WriteField("media_url", "https://via.placeholder.com/300x200/FF0000/FFFFFF?text=Test")

	writer.Close()

	req := httptest.NewRequest("POST", "/api/posts", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(201))

	var postData map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&postData)
	return postData
}

// likeTestPostWithSSE likes a post and returns the response for SSE testing
func likeTestPostWithSSE(app *fiber.App, token string, postID uint) map[string]interface{} {
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/posts/%d/like", postID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(200))

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result
}

// commentTestPostWithSSE comments on a post and returns the response for SSE testing
func commentTestPostWithSSE(app *fiber.App, token string, postID uint, text string) map[string]interface{} {
	commentData := map[string]string{"text": text}
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/posts/%d/comment", postID), strings.NewReader(marshalJSON(commentData)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(200))

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result
}

// marshalJSON is a helper to marshal data to JSON
func marshalJSON(data interface{}) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal JSON: %v", err))
	}
	return string(jsonData)
}

func createTestJWT(userID uint) (string, error) {
	userRepo := postgresRepo.NewUserRepository(sharedContainers.DB)
	authService := service.NewAuthService(userRepo, "test-secret")
	return authService.GenerateJWT(userID)
}
