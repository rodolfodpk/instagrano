package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rodolfodpk/instagrano/internal/logger"
	"go.uber.org/zap"
)

// RequestLogger creates a middleware for logging HTTP requests
func RequestLogger(log *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		
		// Generate request ID
		requestID := generateRequestID()
		c.Locals("requestID", requestID)
		
		// Create logger with request context
		reqLogger := log.WithRequestID(requestID)
		
		// Log request start
		reqLogger.Info("request started",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("ip", c.IP()),
			zap.String("user_agent", c.Get("User-Agent")),
		)
		
		// Process request
		err := c.Next()
		
		// Calculate duration
		duration := time.Since(start)
		
		// Get user ID if available
		var userID uint
		if uid := c.Locals("userID"); uid != nil {
			userID = uid.(uint)
		}
		
		// Log request completion
		fields := []zap.Field{
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("duration", duration),
			zap.Int("response_size", len(c.Response().Body())),
		}
		
		if userID > 0 {
			fields = append(fields, zap.Uint("user_id", userID))
		}
		
		if err != nil {
			fields = append(fields, zap.Error(err))
			reqLogger.Error("request completed with error", fields...)
		} else {
			reqLogger.Info("request completed", fields...)
		}
		
		return err
	}
}

// generateRequestID creates a unique request ID
func generateRequestID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
