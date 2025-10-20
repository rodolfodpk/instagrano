package service

import (
	"sort"

	"github.com/rodolfodpk/instagrano/internal/domain"
	"github.com/rodolfodpk/instagrano/internal/pagination"
	"github.com/rodolfodpk/instagrano/internal/repository/postgres"
	"go.uber.org/zap"
)

type FeedService struct {
	postRepo postgres.PostRepository
	logger   *zap.Logger
}

func NewFeedService(postRepo postgres.PostRepository) *FeedService {
	logger, _ := zap.NewProduction()
	return &FeedService{
		postRepo: postRepo,
		logger:   logger,
	}
}

// GetFeedWithCursor implements cursor-based pagination
func (s *FeedService) GetFeedWithCursor(limit int, cursor string) (*pagination.FeedResult, error) {
	s.logger.Info("getting feed with cursor",
		zap.Int("limit", limit),
		zap.String("cursor", cursor),
	)

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

	s.logger.Info("feed retrieved successfully",
		zap.Int("posts_count", len(posts)),
		zap.Bool("has_more", hasMore),
		zap.String("next_cursor", nextCursor),
	)

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
