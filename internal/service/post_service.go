package service

import (
    "fmt"
    "io"
    "os"
    "path/filepath"
    "time"
    "github.com/rodolfodpk/instagrano/internal/domain"
    "github.com/rodolfodpk/instagrano/internal/repository/postgres"
)

type PostService struct {
    postRepo postgres.PostRepository
}

func NewPostService(postRepo postgres.PostRepository) *PostService {
    return &PostService{postRepo: postRepo}
}

func (s *PostService) CreatePost(userID uint, title, caption string, mediaType domain.MediaType, file io.Reader, filename string) (*domain.Post, error) {
    if title == "" {
        return nil, ErrInvalidInput
    }

    // Create uploads directory if it doesn't exist
    uploadDir := "web/public/uploads"
    os.MkdirAll(uploadDir, 0755)

    // Generate unique filename
    ext := filepath.Ext(filename)
    if ext == "" {
        ext = ".jpg" // default extension
    }
    uniqueFilename := fmt.Sprintf("%d-%d%s", userID, time.Now().Unix(), ext)
    filePath := filepath.Join(uploadDir, uniqueFilename)

    // Save file locally
    destFile, err := os.Create(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to create file: %w", err)
    }
    defer destFile.Close()

    _, err = io.Copy(destFile, file)
    if err != nil {
        return nil, fmt.Errorf("failed to save file: %w", err)
    }

    // Generate URL
    mediaURL := fmt.Sprintf("/uploads/%s", uniqueFilename)

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
