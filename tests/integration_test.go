package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	baseURL      = "http://localhost:3007"
	testPassword = "testpassword123"
)

// Generate unique test email for each test run
func getTestEmail() string {
	return fmt.Sprintf("test-%d@example.com", time.Now().Unix())
}

type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	} `json:"user"`
}

type PostResponse struct {
	ID            uint    `json:"id"`
	UserID        uint    `json:"user_id"`
	Title         string  `json:"title"`
	Caption       string  `json:"caption"`
	MediaType     string  `json:"media_type"`
	MediaURL      string  `json:"media_url"`
	LikesCount    int     `json:"likes_count"`
	CommentsCount int     `json:"comments_count"`
	ViewsCount    int     `json:"views_count"`
	Score         float64 `json:"score"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type FeedResponse struct {
	Posts []PostResponse `json:"posts"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

func TestJWTIntegration(t *testing.T) {
	// Check if server is running
	if !isServerRunning() {
		t.Skip("Server is not running. Please start the server with: JWT_SECRET='super-secret-key-for-testing' PORT=3007 go run cmd/api/main.go")
	}

	t.Run("User Registration and Login", func(t *testing.T) {
		testEmail := getTestEmail()

		// Register a new user
		registerData := map[string]string{
			"username": fmt.Sprintf("testuser-%d", time.Now().Unix()),
			"email":    testEmail,
			"password": testPassword,
		}

		registerResp, registerBody := makeRequestWithError(t, "POST", "/api/auth/register", registerData, "")
		if registerResp.StatusCode != 200 && registerResp.StatusCode != 201 {
			t.Logf("Registration failed with status %d: %s", registerResp.StatusCode, registerBody)
			// If user already exists, that's okay for this test
			if strings.Contains(registerBody, "duplicate key") {
				t.Log("User already exists, continuing with login test")
			} else {
				assert.True(t, registerResp.StatusCode == 200 || registerResp.StatusCode == 201, "User registration should succeed with status 200 or 201")
			}
		} else {
			assert.True(t, registerResp.StatusCode == 200 || registerResp.StatusCode == 201, "User registration should succeed with status 200 or 201")
		}

		// Login with the registered user
		loginData := map[string]string{
			"email":    testEmail,
			"password": testPassword,
		}

		loginResp, loginBody := makeRequestWithError(t, "POST", "/api/auth/login", loginData, "")
		if loginResp.StatusCode != 200 {
			t.Logf("Login failed with status %d: %s", loginResp.StatusCode, loginBody)
		}
		assert.Equal(t, 200, loginResp.StatusCode, "Login should succeed")

		var loginResponse LoginResponse
		err := json.NewDecoder(loginResp.Body).Decode(&loginResponse)
		require.NoError(t, err, "Should be able to decode login response")

		assert.NotEmpty(t, loginResponse.Token, "Token should not be empty")
		assert.Equal(t, testEmail, loginResponse.User.Email, "User email should match")

		// Test JWT authentication with /me endpoint
		meResp := makeRequest(t, "GET", "/api/me", nil, loginResponse.Token)
		assert.Equal(t, 200, meResp.StatusCode, "JWT authentication should work")

		var meResponse map[string]interface{}
		err = json.NewDecoder(meResp.Body).Decode(&meResponse)
		require.NoError(t, err, "Should be able to decode /me response")

		assert.Equal(t, float64(loginResponse.User.ID), meResponse["user_id"], "User ID should match")
	})

	t.Run("File Upload and Post Creation", func(t *testing.T) {
		testEmail := getTestEmail()

		// First login to get a token
		loginData := map[string]string{
			"email":    testEmail,
			"password": testPassword,
		}

		loginResp := makeRequest(t, "POST", "/api/auth/login", loginData, "")
		require.Equal(t, 200, loginResp.StatusCode, "Login should succeed")

		var loginResponse LoginResponse
		err := json.NewDecoder(loginResp.Body).Decode(&loginResponse)
		require.NoError(t, err, "Should be able to decode login response")

		// Create a test file
		testFileContent := "This is a test file for upload"
		testFileName := "test.txt"

		// Create multipart form data
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		// Add form fields
		writer.WriteField("title", "Integration Test Post")
		writer.WriteField("caption", "This is a test post created by integration test")
		writer.WriteField("media_type", "image")

		// Add file
		fileWriter, err := writer.CreateFormFile("media", testFileName)
		require.NoError(t, err, "Should be able to create form file")

		_, err = fileWriter.Write([]byte(testFileContent))
		require.NoError(t, err, "Should be able to write file content")

		err = writer.Close()
		require.NoError(t, err, "Should be able to close writer")

		// Make the request
		req, err := http.NewRequest("POST", baseURL+"/api/posts", &buf)
		require.NoError(t, err, "Should be able to create request")

		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+loginResponse.Token)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err, "Should be able to make request")
		defer resp.Body.Close()

		assert.Equal(t, 201, resp.StatusCode, "Post creation should succeed")

		var postResponse PostResponse
		err = json.NewDecoder(resp.Body).Decode(&postResponse)
		require.NoError(t, err, "Should be able to decode post response")

		assert.Equal(t, "Integration Test Post", postResponse.Title, "Post title should match")
		assert.Equal(t, "This is a test post created by integration test", postResponse.Caption, "Post caption should match")
		assert.Equal(t, "image", postResponse.MediaType, "Media type should match")
		assert.NotEmpty(t, postResponse.MediaURL, "Media URL should not be empty")
		assert.True(t, strings.HasPrefix(postResponse.MediaURL, "/uploads/"), "Media URL should start with /uploads/")
	})

	t.Run("Feed Access", func(t *testing.T) {
		testEmail := getTestEmail()

		// Login to get a token
		loginData := map[string]string{
			"email":    testEmail,
			"password": testPassword,
		}

		loginResp := makeRequest(t, "POST", "/api/auth/login", loginData, "")
		require.Equal(t, 200, loginResp.StatusCode, "Login should succeed")

		var loginResponse LoginResponse
		err := json.NewDecoder(loginResp.Body).Decode(&loginResponse)
		require.NoError(t, err, "Should be able to decode login response")

		// Access the feed
		feedResp := makeRequest(t, "GET", "/api/feed", nil, loginResponse.Token)
		assert.Equal(t, 200, feedResp.StatusCode, "Feed access should succeed")

		var feedResponse FeedResponse
		err = json.NewDecoder(feedResp.Body).Decode(&feedResponse)
		require.NoError(t, err, "Should be able to decode feed response")

		assert.GreaterOrEqual(t, len(feedResponse.Posts), 1, "Feed should contain at least one post")
		assert.Equal(t, 1, feedResponse.Page, "Page should be 1")
		assert.Equal(t, 20, feedResponse.Limit, "Limit should be 20")
	})

	t.Run("JWT Token Validation Edge Cases", func(t *testing.T) {
		// Test with invalid token
		invalidResp := makeRequest(t, "GET", "/api/me", nil, "invalid-token")
		assert.Equal(t, 401, invalidResp.StatusCode, "Invalid token should return 401")

		// Test with malformed Authorization header
		req, err := http.NewRequest("GET", baseURL+"/api/me", nil)
		require.NoError(t, err, "Should be able to create request")

		req.Header.Set("Authorization", "InvalidFormat token")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err, "Should be able to make request")
		defer resp.Body.Close()

		assert.Equal(t, 401, resp.StatusCode, "Malformed Authorization header should return 401")

		// Test with missing Authorization header
		req, err = http.NewRequest("GET", baseURL+"/api/me", nil)
		require.NoError(t, err, "Should be able to create request")

		resp, err = client.Do(req)
		require.NoError(t, err, "Should be able to make request")
		defer resp.Body.Close()

		assert.Equal(t, 401, resp.StatusCode, "Missing Authorization header should return 401")
	})
}

func makeRequest(t *testing.T, method, path string, data interface{}, token string) *http.Response {
	var body io.Reader

	if data != nil {
		jsonData, err := json.Marshal(data)
		require.NoError(t, err, "Should be able to marshal data")
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, baseURL+path, body)
	require.NoError(t, err, "Should be able to create request")

	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err, "Should be able to make request")

	return resp
}

func makeRequestWithError(t *testing.T, method, path string, data interface{}, token string) (*http.Response, string) {
	var body io.Reader

	if data != nil {
		jsonData, err := json.Marshal(data)
		require.NoError(t, err, "Should be able to marshal data")
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, baseURL+path, body)
	require.NoError(t, err, "Should be able to create request")

	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err, "Should be able to make request")

	// Read response body for error messages
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Should be able to read response body")

	// Create a new response with the body for further use
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return resp, string(bodyBytes)
}

func isServerRunning() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

func TestMain(m *testing.M) {
	// Check if server is running before running tests
	if !isServerRunning() {
		fmt.Println("⚠️  Server is not running!")
		fmt.Println("Please start the server with:")
		fmt.Println("JWT_SECRET='super-secret-key-for-testing' PORT=3007 go run cmd/api/main.go")
		fmt.Println("Then run the tests again.")
		os.Exit(1)
	}

	// Run tests
	code := m.Run()
	os.Exit(code)
}
