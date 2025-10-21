package service

import (
	"fmt"
	"io"
	"net/url"
	"strings"

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

	return post, nil
}

func (s *PostService) GetPost(postID uint) (*domain.Post, error) {
	return s.postRepo.FindByID(postID)
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

	return post, nil
}
