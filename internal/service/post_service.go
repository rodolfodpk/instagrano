package service

import (
    "fmt"
    "github.com/rodolfodpk/instagrano/internal/domain"
    "github.com/rodolfodpk/instagrano/internal/repository/postgres"
)

type PostService struct {
    postRepo postgres.PostRepository
}

func NewPostService(postRepo postgres.PostRepository) *PostService {
    return &PostService{postRepo: postRepo}
}

func (s *PostService) CreatePost(userID uint, title, caption string, mediaType domain.MediaType) (*domain.Post, error) {
    if title == "" {
        return nil, ErrInvalidInput
    }

    // Temporary mock URL (we'll add S3 later)
    mockURL := fmt.Sprintf("https://mock-s3.example.com/%d-%s", userID, title)

    post := &domain.Post{
        UserID:    userID,
        Title:     title,
        Caption:   caption,
        MediaType: mediaType,
        MediaURL:  mockURL,
    }

    if err := s.postRepo.Create(post); err != nil {
        return nil, err
    }

    return post, nil
}

func (s *PostService) GetPost(postID uint) (*domain.Post, error) {
    return s.postRepo.FindByID(postID)
}
