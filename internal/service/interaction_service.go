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
	postRepo    postgres.PostRepository
	cache       cache.Cache
}

func NewInteractionService(likeRepo postgres.LikeRepository, commentRepo postgres.CommentRepository, postRepo postgres.PostRepository, cache cache.Cache) *InteractionService {
	return &InteractionService{
		likeRepo:    likeRepo,
		commentRepo: commentRepo,
		postRepo:    postRepo,
		cache:       cache,
	}
}

func (s *InteractionService) LikePost(userID, postID uint) (int, int, error) {
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

	// Get updated counts
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		return 0, 0, err
	}

	return post.LikesCount, post.CommentsCount, nil
}

func (s *InteractionService) CommentPost(userID, postID uint, text string) (int, int, error) {
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

	// Get updated counts
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		return 0, 0, err
	}

	return post.LikesCount, post.CommentsCount, nil
}
