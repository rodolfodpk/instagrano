package service

import (
    "fmt"
    "io"
    "github.com/rodolfodpk/instagrano/internal/domain"
    "github.com/rodolfodpk/instagrano/internal/repository/postgres"
    "github.com/rodolfodpk/instagrano/internal/repository/s3"
)

type PostService struct {
    postRepo     postgres.PostRepository
    mediaStorage s3.MediaStorage
}

func NewPostService(postRepo postgres.PostRepository, mediaStorage s3.MediaStorage) *PostService {
    return &PostService{
        postRepo:     postRepo,
        mediaStorage: mediaStorage,
    }
}

func (s *PostService) CreatePost(userID uint, title, caption string, mediaType domain.MediaType, file io.Reader, filename string) (*domain.Post, error) {
    if title == "" {
        return nil, ErrInvalidInput
    }

    // Upload file to S3
    contentType := "image/jpeg"
    if mediaType == domain.MediaTypeVideo {
        contentType = "video/mp4"
    }

    key, err := s.mediaStorage.Upload(file, filename, contentType)
    if err != nil {
        return nil, fmt.Errorf("failed to upload media: %w", err)
    }

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

    return post, nil
}

func (s *PostService) GetPost(postID uint) (*domain.Post, error) {
    return s.postRepo.FindByID(postID)
}
