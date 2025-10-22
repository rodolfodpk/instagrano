package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/rodolfodpk/instagrano/internal/cache"
	"github.com/rodolfodpk/instagrano/internal/domain"
	"github.com/rodolfodpk/instagrano/internal/repository/postgres"
	"github.com/rodolfodpk/instagrano/internal/repository/s3"
	"go.uber.org/zap"
)

type PostService struct {
	postRepo     postgres.PostRepository
	mediaStorage s3.MediaStorage
	cache        cache.Cache
	cacheTTL     time.Duration
	logger       *zap.Logger
}

func NewPostService(postRepo postgres.PostRepository, mediaStorage s3.MediaStorage, cache cache.Cache, cacheTTL time.Duration) *PostService {
	logger, _ := zap.NewProduction()
	return &PostService{
		postRepo:     postRepo,
		mediaStorage: mediaStorage,
		cache:        cache,
		cacheTTL:     cacheTTL,
		logger:       logger,
	}
}

func (s *PostService) CreatePost(userID uint, title, caption string, mediaType domain.MediaType, file io.Reader, filename string) (*domain.Post, error) {
	if title == "" {
		return nil, ErrInvalidInput
	}

	// Determine content type based on media type
	contentType := "image/jpeg"
	if mediaType == domain.MediaTypeVideo {
		contentType = "video/mp4"
	}

	// Upload file to S3
	key, err := s.mediaStorage.Upload(file, filename, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Generate URL from S3 key
	mediaURL := s.mediaStorage.GetURL(key)

	post := &domain.Post{
		UserID:    userID,
		Title:     title,
		Caption:   caption,
		MediaType: mediaType,
		MediaURL:  mediaURL,
	}

	if err := s.postRepo.Create(post); err != nil {
		return nil, err
	}

	// Invalidate feed cache to ensure new post appears
	s.invalidateFeedCache()

	return post, nil
}

func (s *PostService) GetPost(postID uint) (*domain.Post, error) {
	cacheKey := fmt.Sprintf("post:%d", postID)
	ctx := context.Background()

	s.logger.Info("getting post", zap.Uint("post_id", postID))

	// Try cache first
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		var post domain.Post
		if unmarshalErr := json.Unmarshal(cached, &post); unmarshalErr == nil {
			s.logger.Info("post cache hit", zap.Uint("post_id", postID))
			return &post, nil
		}
	}

	// Cache miss - fetch from database
	s.logger.Info("post cache miss", zap.Uint("post_id", postID))
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		return nil, err
	}

	// Store in cache (best effort)
	if data, marshalErr := json.Marshal(post); marshalErr == nil {
		if setErr := s.cache.Set(ctx, cacheKey, data, s.cacheTTL); setErr != nil {
			s.logger.Warn("failed to cache post", zap.Uint("post_id", postID), zap.Error(setErr))
		} else {
			s.logger.Info("cached post", zap.Uint("post_id", postID))
		}
	}

	return post, nil
}

// CreatePostFromURL creates a post by downloading media from a URL
func (s *PostService) CreatePostFromURL(userID uint, title, caption, mediaURL string) (*domain.Post, error) {
	if title == "" {
		return nil, ErrInvalidInput
	}

	if mediaURL == "" {
		return nil, fmt.Errorf("media URL is required")
	}

	// Validate URL format
	if _, err := url.Parse(mediaURL); err != nil {
		return nil, fmt.Errorf("invalid URL format: %w", err)
	}

	// Download from URL and upload to S3
	key, contentType, err := s.mediaStorage.UploadFromURL(mediaURL)
	if err != nil {
		return nil, fmt.Errorf("failed to process media URL: %w", err)
	}

	// Determine media type from content type
	mediaType := domain.MediaTypeImage
	if strings.HasPrefix(contentType, "video/") {
		mediaType = domain.MediaTypeVideo
	}

	// Generate S3 URL
	s3URL := s.mediaStorage.GetURL(key)

	post := &domain.Post{
		UserID:    userID,
		Title:     title,
		Caption:   caption,
		MediaType: mediaType,
		MediaURL:  s3URL,
	}

	if err := s.postRepo.Create(post); err != nil {
		return nil, err
	}

	// Invalidate feed cache to ensure new post appears
	s.invalidateFeedCache()

	return post, nil
}

// invalidateFeedCache clears feed cache entries
func (s *PostService) invalidateFeedCache() {
	ctx := context.Background()

	// Clear the most common feed cache key (empty cursor, default limit)
	// This is a simple approach - in production you might want to clear all feed patterns
	cacheKey := "feed:cursor::limit:5" // Default feed cache key
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		s.logger.Warn("failed to clear feed cache",
			zap.String("cache_key", cacheKey),
			zap.Error(err))
	} else {
		s.logger.Info("cleared feed cache", zap.String("cache_key", cacheKey))
	}
}

// DeletePost deletes a post by ID (only by the post author)
func (s *PostService) DeletePost(postID, userID uint) error {
	// First, get the post to check ownership
	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		return fmt.Errorf("post not found")
	}

	// Check if the user is the author of the post
	if post.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	// Delete the post (this will cascade delete likes and comments due to foreign key constraints)
	if err := s.postRepo.Delete(postID); err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	// Invalidate feed cache to ensure deleted post disappears
	s.invalidateFeedCache()

	s.logger.Info("post deleted successfully",
		zap.Uint("post_id", postID),
		zap.Uint("user_id", userID))

	return nil
}
