package service

import (
    "github.com/rodolfodpk/instagrano/internal/domain"
    "github.com/rodolfodpk/instagrano/internal/repository/postgres"
)

type InteractionService struct {
    likeRepo postgres.LikeRepository
}

func NewInteractionService(likeRepo postgres.LikeRepository) *InteractionService {
    return &InteractionService{likeRepo: likeRepo}
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
