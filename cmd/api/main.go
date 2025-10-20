package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/rodolfodpk/instagrano/internal/config"
	"github.com/rodolfodpk/instagrano/internal/handler"
	"github.com/rodolfodpk/instagrano/internal/middleware"
	"github.com/rodolfodpk/instagrano/internal/repository/postgres"
	"github.com/rodolfodpk/instagrano/internal/repository/s3"
	"github.com/rodolfodpk/instagrano/internal/service"
)

func main() {
	cfg := config.Load()

	// Debug: Print JWT secret (first 10 chars for security)
	jwtSecretPreview := cfg.JWTSecret
	if len(jwtSecretPreview) > 10 {
		jwtSecretPreview = jwtSecretPreview[:10] + "..."
	}
	log.Printf("JWT Secret: %s", jwtSecretPreview)

	db, err := postgres.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	// Initialize S3 media storage
	mediaStorage, err := s3.NewMediaStorage(
		cfg.S3Endpoint,
		"us-east-1",
		cfg.S3Bucket,
	)
	if err != nil {
		log.Fatal("S3 connection failed:", err)
	}

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	postRepo := postgres.NewPostRepository(db)
	likeRepo := postgres.NewLikeRepository(db)
	commentRepo := postgres.NewCommentRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	postService := service.NewPostService(postRepo, mediaStorage)
	feedService := service.NewFeedService(postRepo)
	interactionService := service.NewInteractionService(likeRepo, commentRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	postHandler := handler.NewPostHandler(postService)
	feedHandler := handler.NewFeedHandler(feedService)
	interactionHandler := handler.NewInteractionHandler(interactionService)

	app := fiber.New()

	// Serve static files
	app.Static("/", "./web/public")

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":   "ok",
			"database": "connected",
		})
	})

	// Routes
	api := app.Group("/api")
	api.Post("/auth/register", authHandler.Register)
	api.Post("/auth/login", authHandler.Login)

	// Protected routes with JWT
	protected := api.Group("/", middleware.JWT(cfg.JWTSecret))
	protected.Get("/me", func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uint)
		return c.JSON(fiber.Map{"user_id": userID})
	})
	protected.Post("/posts", postHandler.CreatePost)
	protected.Get("/posts/:id", postHandler.GetPost)
	protected.Post("/posts/:id/like", interactionHandler.LikePost)
	protected.Post("/posts/:id/comment", interactionHandler.CommentPost)
	protected.Get("/feed", feedHandler.GetFeed)

	log.Printf("🚀 Server starting on port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
