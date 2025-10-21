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
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/rodolfodpk/instagrano/internal/cache"
	"github.com/rodolfodpk/instagrano/internal/config"
	"github.com/rodolfodpk/instagrano/internal/events"
	"github.com/rodolfodpk/instagrano/internal/handler"
	"github.com/rodolfodpk/instagrano/internal/logger"
	"github.com/rodolfodpk/instagrano/internal/middleware"
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

	db, err := postgres.Connect(cfg.DatabaseURL)
	if err != nil {
		appLogger.Fatal("database connection failed", zap.Error(err))
	}
	defer db.Close()

	appLogger.Info("database connected successfully")

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

	// Initialize Redis cache
	redisCache, err := cache.NewRedisCache(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB, appLogger.Logger)
	if err != nil {
		appLogger.Fatal("redis connection failed", zap.Error(err))
	}

	appLogger.Info("redis cache initialized",
		zap.String("addr", cfg.RedisAddr),
		zap.Duration("ttl", cfg.CacheTTL),
	)

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	postRepo := postgres.NewPostRepository(db)
	likeRepo := postgres.NewLikeRepository(db)
	commentRepo := postgres.NewCommentRepository(db)
	viewRepo := postgres.NewPostViewRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	postService := service.NewPostService(postRepo, mediaStorage, redisCache, cfg.CacheTTL)
	feedService := service.NewFeedService(postRepo, redisCache, cfg.CacheTTL)
	interactionService := service.NewInteractionService(likeRepo, commentRepo, postRepo, redisCache)
	viewService := service.NewPostViewService(viewRepo)

	// Initialize event publisher
	eventPublisher := events.NewPublisher(redisCache, appLogger.Logger)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	postHandler := handler.NewPostHandler(postService, eventPublisher, appLogger.Logger)
	feedHandler := handler.NewFeedHandler(feedService, cfg)
	interactionHandler := handler.NewInteractionHandler(interactionService, eventPublisher, appLogger.Logger)
	viewHandler := handler.NewPostViewHandler(viewService)
	testImageHandler := handler.NewTestImageHandler()
	sseHandler := handler.NewSSEHandler(redisCache, appLogger.Logger, cfg.JWTSecret)

	app := fiber.New()

	// Add CORS middleware
	app.Use(cors.New())

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

	// Serve static files
	app.Static("/static", "./web/public")

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
	protected.Post("/posts/:id/view/start", viewHandler.StartView)
	protected.Post("/posts/:id/view/end", viewHandler.EndView)
	protected.Get("/feed", feedHandler.GetFeed)
	protected.Get("/events/stream", sseHandler.Stream)

	appLogger.Info("server starting",
		zap.String("port", cfg.Port),
		zap.String("environment", "development"),
	)

	log.Fatal(app.Listen(":" + cfg.Port))
}
