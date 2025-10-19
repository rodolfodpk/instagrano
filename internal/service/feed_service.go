package service

import (
    "sort"
    "github.com/rodolfodpk/instagrano/internal/domain"
    "github.com/rodolfodpk/instagrano/internal/repository/postgres"
)

type FeedService struct {
    postRepo postgres.PostRepository
}

func NewFeedService(postRepo postgres.PostRepository) *FeedService {
    return &FeedService{postRepo: postRepo}
}

func (s *FeedService) GetFeed(page, limit int) ([]*domain.Post, error) {
    offset := (page - 1) * limit
    posts, err := s.postRepo.GetFeed(limit, offset)
    if err != nil {
        return nil, err
    }

    for _, post := range posts {
        post.Score = post.CalculateScore()
    }

    sort.Slice(posts, func(i, j int) bool {
        return posts[i].Score > posts[j].Score
    })

    return posts, nil
}
