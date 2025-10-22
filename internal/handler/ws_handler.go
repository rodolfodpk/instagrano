package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rodolfodpk/instagrano/internal/cache"
	"github.com/rodolfodpk/instagrano/internal/events"
	"go.uber.org/zap"
)

type WSHandler struct {
	cache     cache.Cache
	logger    *zap.Logger
	jwtSecret string
}

func NewWSHandler(cache cache.Cache, logger *zap.Logger, jwtSecret string) *WSHandler {
	return &WSHandler{
		cache:     cache,
		logger:    logger,
		jwtSecret: jwtSecret,
	}
}

// HandleWebSocket handles WebSocket connections for real-time events
func (h *WSHandler) HandleWebSocket(c *websocket.Conn) {
	// Get token from query parameter
	token := c.Query("token")
	if token == "" {
		h.logger.Error("missing token in WebSocket request")
		c.WriteJSON(fiber.Map{"error": "token required"})
		c.Close()
		return
	}

	// Validate JWT token
	userID, err := h.validateJWT(token)
	if err != nil {
		h.logger.Error("invalid JWT token",
			zap.Error(err),
			zap.String("token", token[:min(len(token), 20)]+"..."))
		c.WriteJSON(fiber.Map{"error": "invalid token"})
		c.Close()
		return
	}

	h.logger.Info("WebSocket connection established", zap.Uint("user_id", userID))

	// Create context for this connection
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Subscribe to Redis events channel
	eventCh, err := h.cache.Subscribe(ctx, events.EventChannel)
	if err != nil {
		h.logger.Error("failed to subscribe to events channel",
			zap.Error(err),
			zap.Uint("user_id", userID))
		c.WriteJSON(fiber.Map{"error": "failed to subscribe to events"})
		c.Close()
		return
	}

	h.logger.Info("subscribed to Redis events channel",
		zap.String("channel", events.EventChannel),
		zap.Uint("user_id", userID))

	// Send initial connection message
	err = c.WriteJSON(map[string]interface{}{
		"type":    "connected",
		"message": "Connected to real-time updates",
		"user_id": userID,
	})
	if err != nil {
		h.logger.Error("failed to send connection message", zap.Error(err))
		return
	}

	// Create ticker for heartbeat
	heartbeatTicker := time.NewTicker(5 * time.Second)
	defer heartbeatTicker.Stop()

	// Handle incoming messages (ping/pong)
	go func() {
		for {
			if _, _, err := c.ReadMessage(); err != nil {
				h.logger.Info("WebSocket read error, closing connection",
					zap.Error(err),
					zap.Uint("user_id", userID))
				cancel()
				return
			}
		}
	}()

	// Main event loop
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

			// Log event received
			h.logger.Info("WebSocket event received",
				zap.String("event_type", string(event.Type)),
				zap.Uint("post_id", event.PostID),
				zap.Uint("triggered_by_user_id", event.TriggeredByUserID),
				zap.Uint("current_user_id", userID))

			// Send event to client
			if err := c.WriteJSON(event); err != nil {
				h.logger.Error("failed to send event to client",
					zap.Error(err),
					zap.Uint("user_id", userID))
				return
			}

			h.logger.Info("WebSocket event sent",
				zap.String("event_type", string(event.Type)),
				zap.Uint("post_id", event.PostID))

		case <-heartbeatTicker.C:
			// Send heartbeat
			err := c.WriteJSON(map[string]interface{}{
				"type":      "heartbeat",
				"timestamp": time.Now().Unix(),
			})
			if err != nil {
				h.logger.Error("failed to send heartbeat", zap.Error(err))
				return
			}

		case <-ctx.Done():
			h.logger.Info("WebSocket connection closed",
				zap.Uint("user_id", userID))
			return
		}
	}
}

// validateJWT validates a JWT token and returns the user ID
func (h *WSHandler) validateJWT(tokenString string) (uint, error) {
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
