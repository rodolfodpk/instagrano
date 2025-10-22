package service

import (
	"context"
	"fmt"

	"github.com/rodolfodpk/instagrano/internal/cache"
	"github.com/rodolfodpk/instagrano/internal/domain"
	"github.com/rodolfodpk/instagrano/internal/events"
	"github.com/rodolfodpk/instagrano/internal/repository/postgres"
	"go.uber.org/zap"
)

type InteractionService struct {
	likeRepo       postgres.LikeRepository
	commentRepo    postgres.CommentRepository
	postRepo       postgres.PostRepository
	cache          cache.Cache
	eventPublisher *events.Publisher
	logger         *zap.Logger
}

func NewInteractionService(likeRepo postgres.LikeRepository, commentRepo postgres.CommentRepository, postRepo postgres.PostRepository, cache cache.Cache, eventPublisher *events.Publisher, logger *zap.Logger) *InteractionService {
	return &InteractionService{
		likeRepo:       likeRepo,
		commentRepo:    commentRepo,
		postRepo:       postRepo,
		cache:          cache,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

func (s *InteractionService) LikePost(userID, postID uint) (int, int, error) {
	// Check if user already liked this post
	existingLike, err := s.likeRepo.FindByUserAndPost(userID, postID)
	if err != nil {
		return 0, 0, err
	}

	if existingLike != nil {
		// User already liked - unlike the post
		return s.unlikePost(userID, postID)
	} else {
		// User hasn't liked - like the post
		return s.likePost(userID, postID)
	}
}

func (s *InteractionService) likePost(userID, postID uint) (int, int, error) {
	like := &domain.Like{
		UserID: userID,
		PostID: postID,
	}

	if err := s.likeRepo.Create(like); err != nil {
		return 0, 0, err
	}

	if err := s.likeRepo.IncrementPostLikeCount(postID); err != nil {
		return 0, 0, err
	}

	// Invalidate post cache
	cacheKey := fmt.Sprintf("post:%d", postID)
	ctx := context.Background()
	s.cache.Delete(ctx, cacheKey)

	// Invalidate feed cache to ensure updated like counts appear
	s.invalidateFeedCache()

	// Get updated counts
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		return 0, 0, err
	}

	// Publish like event for real-time updates
	if err := s.eventPublisher.PublishPostLiked(ctx, postID, userID, post.LikesCount, post.CommentsCount); err != nil {
		s.logger.Error("failed to publish post liked event",
			zap.Error(err),
			zap.Uint("post_id", postID),
			zap.Uint("user_id", userID))
	}

	return post.LikesCount, post.CommentsCount, nil
}

func (s *InteractionService) unlikePost(userID, postID uint) (int, int, error) {
	if err := s.likeRepo.Delete(userID, postID); err != nil {
		return 0, 0, err
	}

	if err := s.likeRepo.DecrementPostLikeCount(postID); err != nil {
		return 0, 0, err
	}

	// Invalidate post cache
	cacheKey := fmt.Sprintf("post:%d", postID)
	ctx := context.Background()
	s.cache.Delete(ctx, cacheKey)

	// Invalidate feed cache to ensure updated like counts appear
	s.invalidateFeedCache()

	// Get updated counts
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		return 0, 0, err
	}

	// Publish unlike event for real-time updates (we can reuse the same event type)
	if err := s.eventPublisher.PublishPostLiked(ctx, postID, userID, post.LikesCount, post.CommentsCount); err != nil {
		s.logger.Error("failed to publish post unliked event",
			zap.Error(err),
			zap.Uint("post_id", postID),
			zap.Uint("user_id", userID))
	}

	return post.LikesCount, post.CommentsCount, nil
}

func (s *InteractionService) CommentPost(userID, postID uint, text string, username string) (int, int, error) {
	comment := &domain.Comment{
		UserID: userID,
		PostID: postID,
		Text:   text,
	}

	if err := s.commentRepo.Create(comment); err != nil {
		return 0, 0, err
	}

	if err := s.commentRepo.IncrementPostCommentCount(postID); err != nil {
		return 0, 0, err
	}

	// Invalidate post cache
	cacheKey := fmt.Sprintf("post:%d", postID)
	ctx := context.Background()
	s.cache.Delete(ctx, cacheKey)

	// Invalidate feed cache to ensure updated comment counts appear
	s.invalidateFeedCache()

	// Get updated counts
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		return 0, 0, err
	}

	// Publish comment event with full comment data for real-time updates
	eventComment := &events.Comment{
		ID:        comment.ID,
		Text:      comment.Text,
		Username:  username,
		UserID:    userID,
		CreatedAt: comment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if err := s.eventPublisher.PublishPostCommented(ctx, postID, userID, post.LikesCount, post.CommentsCount, eventComment); err != nil {
		s.logger.Error("failed to publish post commented event",
			zap.Error(err),
			zap.Uint("post_id", postID),
			zap.Uint("user_id", userID))
	}

	return post.LikesCount, post.CommentsCount, nil
}

// GetComments retrieves all comments for a post
func (s *InteractionService) GetComments(postID uint) ([]*domain.Comment, error) {
	return s.commentRepo.FindByPostID(postID)
}

// GetPost retrieves a post by ID
func (s *InteractionService) GetPost(postID uint) (*domain.Post, error) {
	return s.postRepo.FindByID(postID)
}

// invalidateFeedCache clears feed cache entries
func (s *InteractionService) invalidateFeedCache() {
	ctx := context.Background()

	// Clear the most common feed cache key (empty cursor, default limit)
	// This is a simple approach - in production you might want to clear all feed patterns
	cacheKey := "feed:cursor::limit:5" // Default feed cache key
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("failed to clear feed cache: %v\n", err)
	} else {
		fmt.Printf("cleared feed cache: %s\n", cacheKey)
	}
}
