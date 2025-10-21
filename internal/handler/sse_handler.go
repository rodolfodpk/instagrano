package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rodolfodpk/instagrano/internal/cache"
	"github.com/rodolfodpk/instagrano/internal/events"
	"go.uber.org/zap"
)

type SSEHandler struct {
	cache     cache.Cache
	logger    *zap.Logger
	jwtSecret string
}

func NewSSEHandler(cache cache.Cache, logger *zap.Logger, jwtSecret string) *SSEHandler {
	return &SSEHandler{
		cache:     cache,
		logger:    logger,
		jwtSecret: jwtSecret,
	}
}

// Stream godoc
// @Summary      Server-Sent Events stream
// @Description  Stream real-time events for authenticated user
// @Tags         events
// @Produce      text/event-stream
// @Param        token  query     string  true  "JWT token"
// @Success      200  {string}  string  "SSE stream"
// @Failure      401  {object}  object{error=string}
// @Router       /events/stream [get]
func (h *SSEHandler) Stream(c *fiber.Ctx) error {
	// Get token from query parameter (EventSource doesn't support custom headers)
	token := c.Query("token")
	if token == "" {
		return c.Status(401).JSON(fiber.Map{"error": "token required"})
	}

	// Validate JWT token
	userID, err := h.validateJWT(token)
	if err != nil {
		h.logger.Warn("SSE token validation failed", 
			zap.Error(err),
			zap.String("token", token[:min(len(token), 20)]+"..."))
		return c.Status(401).JSON(fiber.Map{"error": "invalid token"})
	}

	// Set SSE headers
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Access-Control-Allow-Origin", "*")
	c.Set("Access-Control-Allow-Headers", "Cache-Control")

	h.logger.Info("SSE connection established",
		zap.Uint("user_id", userID))

	// Create context that cancels when client disconnects
	ctx, cancel := context.WithCancel(c.Context())
	defer cancel()

	// Subscribe to Redis events channel
	eventCh, err := h.cache.Subscribe(ctx, events.EventChannel)
	if err != nil {
		h.logger.Error("failed to subscribe to events channel",
			zap.Error(err),
			zap.Uint("user_id", userID))
		return c.Status(500).JSON(fiber.Map{"error": "failed to establish event stream"})
	}

	// Send initial connection event
	h.sendSSEEvent(c, "connected", map[string]interface{}{
		"message": "Connected to real-time updates",
		"user_id": userID,
	})

	// Send heartbeat every 30 seconds
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	// Process events
	for {
		select {
		case eventJSON := <-eventCh:
			// Parse the event
			var event events.Event
			if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
				h.logger.Error("failed to parse event",
					zap.Error(err),
					zap.String("event_json", eventJSON))
				continue
			}

			// Filter out events triggered by the current user
			if event.TriggeredByUserID == userID {
				h.logger.Debug("filtered out self-triggered event",
					zap.String("event_type", string(event.Type)),
					zap.Uint("post_id", event.PostID),
					zap.Uint("user_id", userID))
				continue
			}

			// Send event to client
			h.sendSSEEvent(c, string(event.Type), event)

		case <-heartbeatTicker.C:
			// Send heartbeat to keep connection alive
			h.sendSSEEvent(c, "heartbeat", map[string]interface{}{
				"timestamp": time.Now().Unix(),
			})

		case <-ctx.Done():
			h.logger.Info("SSE connection closed",
				zap.Uint("user_id", userID))
			return nil
		}
	}
}

// sendSSEEvent sends an SSE-formatted event to the client
func (h *SSEHandler) sendSSEEvent(c *fiber.Ctx, eventType string, data interface{}) {
	// Marshal data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		h.logger.Error("failed to marshal SSE event data",
			zap.Error(err),
			zap.String("event_type", eventType))
		return
	}

	// Format as SSE event
	event := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, string(jsonData))

	// Send to client
	if _, err := c.Write([]byte(event)); err != nil {
		h.logger.Error("failed to write SSE event",
			zap.Error(err),
			zap.String("event_type", eventType))
		return
	}

	h.logger.Debug("SSE event sent",
		zap.String("event_type", eventType))
}

// validateJWT validates a JWT token and returns the user ID
func (h *SSEHandler) validateJWT(tokenString string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.jwtSecret), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if userIDFloat, ok := claims["user_id"].(float64); ok {
			return uint(userIDFloat), nil
		}
		return 0, fmt.Errorf("invalid user_id in token")
	}

	return 0, fmt.Errorf("invalid token")
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
