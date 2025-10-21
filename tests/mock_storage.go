package tests

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/rodolfodpk/instagrano/internal/webclient"
)

// MockMediaStorage implements s3.MediaStorage interface for testing
type MockMediaStorage struct {
	files      map[string][]byte
	httpClient webclient.HTTPClient
}

// NewMockMediaStorage creates a new mock media storage
func NewMockMediaStorage() *MockMediaStorage {
	// Use mock controller config for tests
	webclientConfig := webclient.Config{
		UseMockController: true,
		MockBaseURL:       "http://localhost:8080",
		RealURLTimeout:    5 * time.Second,
	}

	return &MockMediaStorage{
		files:      make(map[string][]byte),
		httpClient: webclient.NewDefaultHTTPClient(webclientConfig),
	}
}

// Upload simulates file upload to S3
func (m *MockMediaStorage) Upload(file io.Reader, filename, contentType string) (string, error) {
	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Generate a mock S3 key
	key := fmt.Sprintf("mock-s3/%s", filename)

	// Store in memory
	m.files[key] = content

	return key, nil
}

// GetURL returns a mock URL for the given key
func (m *MockMediaStorage) GetURL(key string) string {
	return fmt.Sprintf("http://mock-s3.example.com/%s", key)
}

// GetFile returns the stored file content (for testing)
func (m *MockMediaStorage) GetFile(key string) ([]byte, bool) {
	content, exists := m.files[key]
	return content, exists
}

// UploadFromURL downloads media from a URL and stores it in memory
func (m *MockMediaStorage) UploadFromURL(url string) (string, string, error) {
	// Download from URL using webclient
	result, err := m.httpClient.Download(context.Background(), url)
	if err != nil {
		return "", "", err
	}
	defer result.Content.(io.ReadCloser).Close()

	// Read content
	content, err := io.ReadAll(result.Content)
	if err != nil {
		return "", "", err
	}

	// Store in memory
	key := fmt.Sprintf("mock-s3/%s", filepath.Base(url))
	m.files[key] = content

	contentType := result.ContentType
	if contentType == "" {
		contentType = "image/jpeg"
	}

	return key, contentType, nil
}
