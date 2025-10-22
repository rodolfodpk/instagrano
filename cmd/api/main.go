// @title           Instagrano API
// @version         1.0
// @description     A mini Instagram API with posts, feed, likes, and comments
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@instagrano.com

// @license.name  MIT
// @license.url   http://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token

package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"
	"github.com/rodolfodpk/instagrano/internal/cache"
	"github.com/rodolfodpk/instagrano/internal/config"
	"github.com/rodolfodpk/instagrano/internal/events"
	"github.com/rodolfodpk/instagrano/internal/handler"
	"github.com/rodolfodpk/instagrano/internal/logger"
	"github.com/rodolfodpk/instagrano/internal/middleware"
	"github.com/rodolfodpk/instagrano/internal/migration"
	"github.com/rodolfodpk/instagrano/internal/repository/postgres"
	"github.com/rodolfodpk/instagrano/internal/repository/s3"
	"github.com/rodolfodpk/instagrano/internal/service"
	"github.com/rodolfodpk/instagrano/internal/webclient"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	// Initialize structured logger
	appLogger := logger.New(cfg.GetZapLevel(), cfg.LogFormat)
	defer appLogger.Sync()

	// Debug: Print JWT secret (first 10 chars for security)
	jwtSecretPreview := cfg.JWTSecret
	if len(jwtSecretPreview) > 10 {
		jwtSecretPreview = jwtSecretPreview[:10] + "..."
	}
	appLogger.Info("application starting",
		zap.String("jwt_secret_preview", jwtSecretPreview),
		zap.String("log_level", cfg.LogLevel),
		zap.String("log_format", cfg.LogFormat),
	)

	// Run database migrations
	migrationRunner := migration.NewRunner(appLogger.Logger)

	// Extract database name from URL for creation
	dbName := "instagrano" // Default database name
	if err := migrationRunner.CreateDatabaseIfNotExists(cfg.DatabaseURL, dbName); err != nil {
		appLogger.Fatal("failed to create database", zap.Error(err))
	}

	// Connect to the specific database after creation
	db, err := postgres.Connect(cfg.DatabaseURL)
	if err != nil {
		appLogger.Fatal("database connection failed", zap.Error(err))
	}

	appLogger.Info("database connected successfully")

	// Run migrations
	if err := migrationRunner.RunMigrations(cfg.DatabaseURL, "./migrations"); err != nil {
		db.Close()
		appLogger.Fatal("failed to run migrations", zap.Error(err))
	}

	// Keep database connection open for the rest of the application
	defer db.Close()

	// Create webclient config
	webclientConfig := webclient.Config{
		UseMockController: cfg.WebclientUseMock,
		MockBaseURL:       cfg.WebclientMockBaseURL,
		RealURLTimeout:    cfg.WebclientTimeout,
	}

	// Initialize S3 media storage with webclient config
	mediaStorage, err := s3.NewMediaStorage(
		cfg.S3Endpoint,
		cfg.S3Region,
		cfg.S3Bucket,
		webclientConfig,
	)
	if err != nil {
		appLogger.Fatal("s3 connection failed", zap.Error(err))
	}

	appLogger.Info("s3 storage initialized",
		zap.String("endpoint", cfg.S3Endpoint),
		zap.String("bucket", cfg.S3Bucket),
		zap.Bool("webclient_use_mock", cfg.WebclientUseMock),
	)

	// Create S3 bucket if it doesn't exist
	if err := mediaStorage.CreateBucketIfNotExists(); err != nil {
		appLogger.Fatal("failed to create S3 bucket", zap.Error(err))
	}

	// Initialize Redis cache
	redisCache, err := cache.NewRedisCache(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB, appLogger.Logger)
	if err != nil {
		appLogger.Fatal("redis connection failed", zap.Error(err))
	}

	appLogger.Info("redis cache initialized",
		zap.String("addr", cfg.RedisAddr),
		zap.Duration("ttl", cfg.CacheTTL),
	)

	// Clear feed cache on startup
	ctx := context.Background()
	keys, err := redisCache.Keys(ctx, "feed:*")
	if err != nil {
		appLogger.Warn("failed to get feed cache keys", zap.Error(err))
	} else {
		for _, key := range keys {
			if err := redisCache.Delete(ctx, key); err != nil {
				appLogger.Warn("failed to delete cache key", zap.String("key", key), zap.Error(err))
			}
		}
		appLogger.Info("cleared feed cache on startup", zap.Int("keys_deleted", len(keys)))
	}

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	postRepo := postgres.NewPostRepository(db)
	likeRepo := postgres.NewLikeRepository(db)
	commentRepo := postgres.NewCommentRepository(db)
	viewRepo := postgres.NewPostViewRepository(db)

	// Initialize event publisher first
	eventPublisher := events.NewPublisher(redisCache, appLogger.Logger)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	postService := service.NewPostService(postRepo, mediaStorage, redisCache, cfg.CacheTTL)
	feedService := service.NewFeedService(postRepo, redisCache, cfg.CacheTTL)
	interactionService := service.NewInteractionService(likeRepo, commentRepo, postRepo, redisCache, eventPublisher, appLogger.Logger)
	viewService := service.NewPostViewService(viewRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	postHandler := handler.NewPostHandler(postService, eventPublisher, appLogger.Logger)
	feedHandler := handler.NewFeedHandler(feedService, cfg)
	interactionHandler := handler.NewInteractionHandler(interactionService, eventPublisher, appLogger.Logger)
	viewHandler := handler.NewPostViewHandler(viewService)
	testImageHandler := handler.NewTestImageHandler()
	wsHandler := handler.NewWSHandler(redisCache, appLogger.Logger, cfg.JWTSecret)

	app := fiber.New()

	// Add CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: false,
		ExposeHeaders:    "Content-Type",
	}))

	// Add request logging middleware
	app.Use(middleware.RequestLogger(appLogger))

	// HealthCheck godoc
	// @Summary      Health check
	// @Description  Check API and dependencies health
	// @Tags         system
	// @Produce      json
	// @Success      200  {object}  object{status=string,database=string,redis=string}
	// @Failure      503  {object}  object{status=string,database=string,redis=string}
	// @Router       /health [get]
	app.Get("/health", func(c *fiber.Ctx) error {
		// Check Redis
		if err := redisCache.Ping(c.Context()); err != nil {
			return c.Status(503).JSON(fiber.Map{
				"status":   "unhealthy",
				"database": "connected",
				"redis":    "disconnected",
			})
		}

		return c.JSON(fiber.Map{
			"status":   "ok",
			"database": "connected",
			"redis":    "connected",
		})
	})

	// Test image routes (no auth required - register EARLY)
	test := app.Group("/test")
	test.Get("/image", testImageHandler.ServeTestImage)
	test.Get("/image/png", testImageHandler.ServeTestPNG)

	appLogger.Info("registering test image routes",
		zap.String("path", "/test/image"),
		zap.String("path_png", "/test/image/png"),
	)

	// Serve static files (excluding root)
	app.Static("/static", "./web/public")

	// Serve index.html for root path
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("./web/public/index.html")
	})

	// Serve other HTML files directly
	app.Get("/feed.html", func(c *fiber.Ctx) error {
		return c.SendFile("./web/public/feed.html")
	})

	app.Get("/s3-browser.html", func(c *fiber.Ctx) error {
		return c.SendFile("./web/public/s3-browser.html")
	})

	app.Get("/sse-test.html", func(c *fiber.Ctx) error {
		return c.SendFile("./web/public/sse-test.html")
	})

	// Routes
	api := app.Group("/api")
	api.Post("/auth/register", authHandler.Register)
	api.Post("/auth/login", authHandler.Login)

	// WebSocket endpoint (unprotected - handles auth internally via query param)
	api.Get("/events/ws", websocket.New(wsHandler.HandleWebSocket))

	// Protected routes with JWT
	protected := api.Group("/", middleware.JWT(cfg.JWTSecret))
	protected.Get("/auth/me", authHandler.GetMe)
	protected.Post("/posts", postHandler.CreatePost)
	protected.Get("/posts/:id", postHandler.GetPost)
	protected.Delete("/posts/:id", postHandler.DeletePost)
	protected.Post("/posts/:id/like", interactionHandler.LikePost)
	protected.Post("/posts/:id/comment", interactionHandler.CommentPost)
	protected.Get("/posts/:id/comments", interactionHandler.GetComments)
	protected.Post("/posts/:id/view/start", viewHandler.StartView)
	protected.Post("/posts/:id/view/end", viewHandler.EndView)
	protected.Get("/feed", feedHandler.GetFeed)

	appLogger.Info("server starting",
		zap.String("port", cfg.Port),
		zap.String("environment", "development"),
	)

	log.Fatal(app.Listen(":" + cfg.Port))
}
