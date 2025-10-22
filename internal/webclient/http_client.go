package webclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Config holds configuration for the HTTP client
type Config struct {
	UseMockController bool
	MockBaseURL       string
	RealURLTimeout    time.Duration
}

// HTTPClient interface for downloading content from URLs
type HTTPClient interface {
	Download(ctx context.Context, url string) (*DownloadResult, error)
}

// DownloadResult contains the downloaded content and metadata
type DownloadResult struct {
	Content     io.ReadCloser
	ContentType string
	Size        int64
	StatusCode  int
}

// DefaultHTTPClient implements HTTPClient using Go's standard net/http
type DefaultHTTPClient struct {
	client *http.Client
	config Config
}

// NewDefaultHTTPClient creates a new HTTP client with configuration
func NewDefaultHTTPClient(config Config) *DefaultHTTPClient {
	client := &DefaultHTTPClient{
		client: &http.Client{
			Timeout: config.RealURLTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
				MaxIdleConnsPerHost: 2,
			},
		},
		config: config,
	}

	// If using mock controller, replace with mock client for testing
	if config.UseMockController {
		return &DefaultHTTPClient{
			client: &http.Client{
				Transport: &mockTransport{},
			},
			config: config,
		}
	}

	return client
}

// Download downloads content from the given URL
func (c *DefaultHTTPClient) Download(ctx context.Context, url string) (*DownloadResult, error) {
	// If using mock controller, rewrite URLs to point to our test endpoints
	if c.config.UseMockController {
		url = c.rewriteToMockURL(url)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set reasonable headers
	req.Header.Set("User-Agent", "Instagrano/1.0")
	req.Header.Set("Accept", "image/*, video/*, application/octet-stream")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download from URL: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return &DownloadResult{
		Content:     resp.Body,
		ContentType: contentType,
		Size:        resp.ContentLength,
		StatusCode:  resp.StatusCode,
	}, nil
}

// rewriteToMockURL maps external URLs to our static test image
func (c *DefaultHTTPClient) rewriteToMockURL(originalURL string) string {
	// Only map known test URLs to our static image for better performance
	if strings.Contains(originalURL, "placeholder.com") ||
		strings.Contains(originalURL, "httpbin.org") ||
		strings.Contains(originalURL, "example.com") ||
		strings.Contains(originalURL, "localhost") {
		return c.config.MockBaseURL + "/test/image"
	}

	// For unmapped URLs, return the original URL (don't rewrite)
	return originalURL
}

// MockHTTPClient for testing
type MockHTTPClient struct {
	responses map[string]*DownloadResult
}

// NewMockHTTPClient creates a mock HTTP client for testing
func NewMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{
		responses: make(map[string]*DownloadResult),
	}
}

// SetResponse sets a mock response for a given URL
func (m *MockHTTPClient) SetResponse(url string, result *DownloadResult) {
	m.responses[url] = result
}

// Download returns the mock response for the given URL
func (m *MockHTTPClient) Download(ctx context.Context, url string) (*DownloadResult, error) {
	if result, exists := m.responses[url]; exists {
		return result, nil
	}
	return nil, fmt.Errorf("no mock response set for URL: %s", url)
}

// mockTransport implements http.RoundTripper for testing
type mockTransport struct{}

// RoundTrip returns a mock response for test URLs
func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Handle error cases
	if strings.Contains(req.URL.String(), "nonexistent-domain-that-does-not-exist.com") {
		return &http.Response{
			StatusCode: 404,
			Status:     "404 Not Found",
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("Not Found")),
			Request:    req,
		}, nil
	}

	if strings.Contains(req.URL.String(), "not-a-valid-url") || req.URL.Scheme == "" {
		return nil, fmt.Errorf("unsupported protocol scheme \"\"")
	}

	// Create a mock response with fake image data for successful cases
	fakeImageData := []byte("fake-image-data-for-testing")

	resp := &http.Response{
		StatusCode:    200,
		Status:        "200 OK",
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        make(http.Header),
		Body:          io.NopCloser(strings.NewReader(string(fakeImageData))),
		ContentLength: int64(len(fakeImageData)),
		Request:       req,
	}

	// Set appropriate content type based on URL
	if strings.Contains(req.URL.Path, "image") {
		resp.Header.Set("Content-Type", "image/jpeg")
	} else {
		resp.Header.Set("Content-Type", "application/octet-stream")
	}

	return resp, nil
}
