package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rodolfodpk/instagrano/internal/cache"
	"go.uber.org/zap"
)

const (
	// EventChannel is the channel name for broadcasting events
	EventChannel = "instagrano:events"
)

// Publisher handles publishing events to Redis pub/sub
type Publisher struct {
	cache  cache.Cache
	logger *zap.Logger
}

// NewPublisher creates a new event publisher
func NewPublisher(cache cache.Cache, logger *zap.Logger) *Publisher {
	return &Publisher{
		cache:  cache,
		logger: logger,
	}
}

// Publish sends an event to the Redis pub/sub channel
func (p *Publisher) Publish(ctx context.Context, event Event) error {
	// Set timestamp if not already set
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().Unix()
	}

	// Marshal event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		p.logger.Error("failed to marshal event",
			zap.Error(err),
			zap.String("event_type", string(event.Type)),
			zap.Uint("post_id", event.PostID))
		return err
	}

	// Publish to Redis channel
	err = p.cache.Publish(ctx, EventChannel, string(eventJSON))
	if err != nil {
		p.logger.Error("failed to publish event",
			zap.Error(err),
			zap.String("event_type", string(event.Type)),
			zap.Uint("post_id", event.PostID))
		return err
	}

	p.logger.Info("event published successfully",
		zap.String("event_type", string(event.Type)),
		zap.Uint("post_id", event.PostID),
		zap.Uint("triggered_by_user_id", event.TriggeredByUserID))

	return nil
}

// PublishNewPost publishes a new post event
func (p *Publisher) PublishNewPost(ctx context.Context, postID uint, triggeredByUserID uint, post interface{}) error {
	event := Event{
		Type:              EventTypeNewPost,
		PostID:            postID,
		TriggeredByUserID: triggeredByUserID,
		Data:              NewPostData{Post: post},
	}
	return p.Publish(ctx, event)
}

// PublishPostLiked publishes a post liked event
func (p *Publisher) PublishPostLiked(ctx context.Context, postID uint, triggeredByUserID uint, likesCount, commentsCount int) error {
	event := Event{
		Type:              EventTypePostLiked,
		PostID:            postID,
		TriggeredByUserID: triggeredByUserID,
		Data:              PostInteractionData{LikesCount: likesCount, CommentsCount: commentsCount},
	}
	return p.Publish(ctx, event)
}

// PublishPostCommented publishes a post commented event with full comment data
func (p *Publisher) PublishPostCommented(ctx context.Context, postID uint, triggeredByUserID uint, likesCount, commentsCount int, comment *Comment) error {
	event := Event{
		Type:              EventTypePostCommented,
		PostID:            postID,
		TriggeredByUserID: triggeredByUserID,
		Data:              PostInteractionData{LikesCount: likesCount, CommentsCount: commentsCount, Comment: comment},
	}
	return p.Publish(ctx, event)
}

// PublishPostDeleted publishes a post deleted event
func (p *Publisher) PublishPostDeleted(ctx context.Context, postID uint, triggeredByUserID uint) error {
	event := Event{
		Type:              EventTypePostDeleted,
		PostID:            postID,
		TriggeredByUserID: triggeredByUserID,
		Data:              nil, // No additional data needed for deletion
	}
	return p.Publish(ctx, event)
}
