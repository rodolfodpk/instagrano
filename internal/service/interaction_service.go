package service

import (
	"context"
	"fmt"

	"github.com/rodolfodpk/instagrano/internal/cache"
	"github.com/rodolfodpk/instagrano/internal/domain"
	"github.com/rodolfodpk/instagrano/internal/repository/postgres"
)

type InteractionService struct {
	likeRepo    postgres.LikeRepository
	commentRepo postgres.CommentRepository
	cache       cache.Cache
}

func NewInteractionService(likeRepo postgres.LikeRepository, commentRepo postgres.CommentRepository, cache cache.Cache) *InteractionService {
	return &InteractionService{
		likeRepo:    likeRepo,
		commentRepo: commentRepo,
		cache:       cache,
	}
}

func (s *InteractionService) LikePost(userID, postID uint) error {
	like := &domain.Like{
		UserID: userID,
		PostID: postID,
	}

	if err := s.likeRepo.Create(like); err != nil {
		return err
	}

	if err := s.likeRepo.IncrementPostLikeCount(postID); err != nil {
		return err
	}

	// Invalidate post cache
	cacheKey := fmt.Sprintf("post:%d", postID)
	ctx := context.Background()
	s.cache.Delete(ctx, cacheKey)

	return nil
}

func (s *InteractionService) CommentPost(userID, postID uint, text string) error {
	comment := &domain.Comment{
		UserID: userID,
		PostID: postID,
		Text:   text,
	}

	if err := s.commentRepo.Create(comment); err != nil {
		return err
	}

	if err := s.commentRepo.IncrementPostCommentCount(postID); err != nil {
		return err
	}

	// Invalidate post cache
	cacheKey := fmt.Sprintf("post:%d", postID)
	ctx := context.Background()
	s.cache.Delete(ctx, cacheKey)

	return nil
}
