package tests

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/gofiber/fiber/v2"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/rodolfodpk/instagrano/internal/cache"
	"github.com/rodolfodpk/instagrano/internal/config"
	"github.com/rodolfodpk/instagrano/internal/domain"
	postgresRepo "github.com/rodolfodpk/instagrano/internal/repository/postgres"
	"github.com/rodolfodpk/instagrano/internal/service"
	"github.com/rodolfodpk/instagrano/internal/handler"
	"github.com/rodolfodpk/instagrano/internal/middleware"
	"go.uber.org/zap"
)

// TestContainers holds container instances
type TestContainers struct {
	PostgresContainer *postgres.PostgresContainer
	RedisContainer    *redis.RedisContainer
	DB                *sql.DB
	Cache             cache.Cache
}

// setupTestContainers starts PostgreSQL and Redis containers
func setupTestContainers(t *testing.T) (*TestContainers, func()) {
	RegisterTestingT(t)
	ctx := context.Background()

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
	runMigrations(t, db)

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
	if strings.HasPrefix(redisAddr, "redis://") {
		redisAddr = strings.TrimPrefix(redisAddr, "redis://")
	}

	// Create Redis cache client
	logger, _ := zap.NewProduction()
	redisCache, err := cache.NewRedisCache(redisAddr, "", 0, logger)
	Expect(err).NotTo(HaveOccurred())

	containers := &TestContainers{
		PostgresContainer: pgContainer,
		RedisContainer:    redisContainer,
		DB:                db,
		Cache:             redisCache,
	}

	cleanup := func() {
		db.Close()
		pgContainer.Terminate(ctx)
		redisContainer.Terminate(ctx)
	}

	return containers, cleanup
}

// runMigrations applies all SQL migrations
func runMigrations(t *testing.T, db *sql.DB) {
	RegisterTestingT(t)

	migrations := []string{
		"../migrations/001_create_users.up.sql",
		"../migrations/002_create_posts.up.sql",
		"../migrations/003_create_likes.up.sql",
		"../migrations/004_create_comments.up.sql",
	}

	for _, migration := range migrations {
		sql, err := os.ReadFile(migration)
		Expect(err).NotTo(HaveOccurred())

		_, err = db.Exec(string(sql))
		Expect(err).NotTo(HaveOccurred())
	}
}

// setupTestApp creates Fiber app with Testcontainers dependencies
func setupTestApp(t *testing.T) (*fiber.App, *TestContainers, func()) {
	RegisterTestingT(t)

	containers, cleanup := setupTestContainers(t)

	cfg := &config.Config{
		JWTSecret:   "test-secret",
		CacheTTL:    5 * time.Minute,
		S3Endpoint:  "http://localhost:4566",
		S3Bucket:    "test-bucket",
	}

	// Initialize repositories
	userRepo := postgresRepo.NewUserRepository(containers.DB)
	postRepo := postgresRepo.NewPostRepository(containers.DB)
	likeRepo := postgresRepo.NewLikeRepository(containers.DB)
	commentRepo := postgresRepo.NewCommentRepository(containers.DB)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	feedService := service.NewFeedService(postRepo, containers.Cache, cfg.CacheTTL)
	interactionService := service.NewInteractionService(likeRepo, commentRepo)
	
	// Create mock S3 storage for testing
	mockStorage := NewMockMediaStorage()
	postService := service.NewPostService(postRepo, mockStorage)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	feedHandler := handler.NewFeedHandler(feedService, cfg)
	postHandler := handler.NewPostHandler(postService)
	interactionHandler := handler.NewInteractionHandler(interactionService)

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

	protected := api.Group("/", middleware.JWT(cfg.JWTSecret))
	protected.Get("/feed", feedHandler.GetFeed)
	protected.Post("/posts", postHandler.CreatePost)
	protected.Get("/posts/:id", postHandler.GetPost)
	protected.Post("/posts/:id/like", interactionHandler.LikePost)
	protected.Post("/posts/:id/comment", interactionHandler.CommentPost)

	return app, containers, cleanup
}

// Helper functions for creating test data
func createTestUser(t *testing.T, db *sql.DB, username, email string) *domain.User {
	RegisterTestingT(t)

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

func createTestPost(t *testing.T, db *sql.DB, userID uint, title, caption string) *domain.Post {
	RegisterTestingT(t)

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
func registerAndLogin(t *testing.T, app *fiber.App, username, email, password string) string {
	RegisterTestingT(t)

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

// marshalJSON is a helper to marshal data to JSON
func marshalJSON(data interface{}) string {
	jsonData, err := json.Marshal(data)
	Expect(err).NotTo(HaveOccurred())
	return string(jsonData)
}
