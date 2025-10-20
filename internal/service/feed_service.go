package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/rodolfodpk/instagrano/internal/cache"
	"github.com/rodolfodpk/instagrano/internal/domain"
	"github.com/rodolfodpk/instagrano/internal/pagination"
	"github.com/rodolfodpk/instagrano/internal/repository/postgres"
	"go.uber.org/zap"
)

type FeedService struct {
	postRepo postgres.PostRepository
	cache    cache.Cache
	cacheTTL time.Duration
	logger   *zap.Logger
}

func NewFeedService(postRepo postgres.PostRepository, cache cache.Cache, cacheTTL time.Duration) *FeedService {
	logger, _ := zap.NewProduction()
	return &FeedService{
		postRepo: postRepo,
		cache:    cache,
		cacheTTL: cacheTTL,
		logger:   logger,
	}
}

// GetFeedWithCursor implements cursor-based pagination with caching
func (s *FeedService) GetFeedWithCursor(limit int, cursor string) (*pagination.FeedResult, error) {
	start := time.Now()
	cacheKey := fmt.Sprintf("feed:cursor:%s:limit:%d", cursor, limit)
	ctx := context.Background()

	s.logger.Info("getting feed with cursor",
		zap.Int("limit", limit),
		zap.String("cursor", cursor),
		zap.String("cache_key", cacheKey),
	)

	// Try cache first
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		var result pagination.FeedResult
		if unmarshalErr := json.Unmarshal(cached, &result); unmarshalErr == nil {
			s.logger.Info("cache hit",
				zap.String("cache_key", cacheKey),
				zap.Duration("duration", time.Since(start)),
			)
			return &result, nil
		}
	}

	// Cache miss - fetch from database
	s.logger.Info("cache miss - fetching from database",
		zap.String("cache_key", cacheKey),
	)

	result, err := s.getFeedFromDatabase(limit, cursor)
	if err != nil {
		return nil, err
	}

	// Store in cache (best effort - don't fail if cache set fails)
	if data, marshalErr := json.Marshal(result); marshalErr == nil {
		if setErr := s.cache.Set(ctx, cacheKey, data, s.cacheTTL); setErr != nil {
			s.logger.Warn("failed to cache result",
				zap.String("cache_key", cacheKey),
				zap.Error(setErr),
			)
		} else {
			s.logger.Info("cached result",
				zap.String("cache_key", cacheKey),
				zap.Duration("ttl", s.cacheTTL),
			)
		}
	}

	s.logger.Info("feed retrieved successfully",
		zap.Int("posts_count", len(result.Posts)),
		zap.Bool("has_more", result.HasMore),
		zap.Duration("duration", time.Since(start)),
	)

	return result, nil
}

// getFeedFromDatabase fetches feed from the database (extracted for caching logic)
func (s *FeedService) getFeedFromDatabase(limit int, cursor string) (*pagination.FeedResult, error) {
	// Decode cursor if provided
	var cursorObj *pagination.Cursor
	var err error
	if cursor != "" {
		cursorObj, err = pagination.DecodeCursor(cursor)
		if err != nil {
			s.logger.Error("failed to decode cursor", zap.Error(err))
			return nil, err
		}
	}

	// Get posts with cursor-based pagination
	posts, err := s.postRepo.GetFeedWithCursor(limit+1, cursorObj) // +1 to check if there are more
	if err != nil {
		s.logger.Error("failed to get feed from repository", zap.Error(err))
		return nil, err
	}

	// Calculate scores and sort
	for _, post := range posts {
		post.Score = post.CalculateScore()
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Score > posts[j].Score
	})

	// Check if there are more posts
	hasMore := len(posts) > limit
	if hasMore {
		posts = posts[:limit] // Remove the extra post
	}

	// Generate next cursor
	var nextCursor string
	if hasMore && len(posts) > 0 {
		lastPost := posts[len(posts)-1]
		nextCursorObj := &pagination.Cursor{
			Timestamp: lastPost.CreatedAt,
			ID:        lastPost.ID,
		}
		nextCursor = nextCursorObj.Encode()
	}

	return &pagination.FeedResult{
		Posts:      convertPostsToInterface(posts),
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

// GetFeed maintains backward compatibility with page-based pagination
func (s *FeedService) GetFeed(page, limit int) ([]*domain.Post, error) {
	s.logger.Info("getting feed with page-based pagination",
		zap.Int("page", page),
		zap.Int("limit", limit),
	)

	offset := (page - 1) * limit
	posts, err := s.postRepo.GetFeed(limit, offset)
	if err != nil {
		s.logger.Error("failed to get feed from repository", zap.Error(err))
		return nil, err
	}

	for _, post := range posts {
		post.Score = post.CalculateScore()
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Score > posts[j].Score
	})

	s.logger.Info("feed retrieved successfully",
		zap.Int("posts_count", len(posts)),
	)

	return posts, nil
}

// convertPostsToInterface converts []*domain.Post to []interface{}
func convertPostsToInterface(posts []*domain.Post) []interface{} {
	result := make([]interface{}, len(posts))
	for i, post := range posts {
		result[i] = post
	}
	return result
}
