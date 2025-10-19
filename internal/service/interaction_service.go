package service

import (
    "github.com/rodolfodpk/instagrano/internal/domain"
    "github.com/rodolfodpk/instagrano/internal/repository/postgres"
)

type InteractionService struct {
    likeRepo    postgres.LikeRepository
    commentRepo postgres.CommentRepository
}

func NewInteractionService(likeRepo postgres.LikeRepository, commentRepo postgres.CommentRepository) *InteractionService {
    return &InteractionService{
        likeRepo:    likeRepo,
        commentRepo: commentRepo,
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

    return s.likeRepo.IncrementPostLikeCount(postID)
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

    return s.commentRepo.IncrementPostCommentCount(postID)
}
