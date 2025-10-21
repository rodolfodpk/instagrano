package tests

import (
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"time"
)

// MockMediaStorage implements s3.MediaStorage interface for testing
type MockMediaStorage struct {
	files map[string][]byte
}

// NewMockMediaStorage creates a new mock media storage
func NewMockMediaStorage() *MockMediaStorage {
	return &MockMediaStorage{
		files: make(map[string][]byte),
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

// UploadFromURL simulates downloading media from a URL without making real HTTP requests
func (m *MockMediaStorage) UploadFromURL(urlStr string) (string, string, error) {
	// Validate URL format
	parsedURL, err := url.Parse(urlStr)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return "", "", fmt.Errorf("invalid URL format: %s", urlStr)
	}

	// Generate mock key and content (no real HTTP request)
	key := fmt.Sprintf("mock-s3/url-%d", time.Now().Unix())
	mockContent := []byte("mock-image-data")
	m.files[key] = mockContent

	contentType := "image/jpeg"
	if filepath.Ext(urlStr) == ".png" {
		contentType = "image/png"
	}

	return key, contentType, nil
}
