package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"
)

func TestUserRegistrationAndLogin(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers start automatically!
	app, _, cleanup := setupTestApp(t)
	defer cleanup()

	// Given: Registration data
	regData := map[string]string{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "password123",
	}
	regBody, _ := json.Marshal(regData)

	// When: User registers
	regReq := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regResp, err := app.Test(regReq)

	// Then: Registration succeeds
	Expect(err).NotTo(HaveOccurred())
	Expect(regResp.StatusCode).To(Equal(200))

	// When: User logs in
	loginData := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	loginBody, _ := json.Marshal(loginData)
	loginReq := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp, err := app.Test(loginReq)

	// Then: Login succeeds and returns JWT
	Expect(err).NotTo(HaveOccurred())
	Expect(loginResp.StatusCode).To(Equal(200))

	var loginResult map[string]interface{}
	json.NewDecoder(loginResp.Body).Decode(&loginResult)
	Expect(loginResult).To(HaveKey("token"))
	Expect(loginResult["token"]).NotTo(BeEmpty())
}

func TestFeedEndpoint(t *testing.T) {
	RegisterTestingT(t)

	app, _, cleanup := setupTestApp(t)
	defer cleanup()

	// Given: User is registered and logged in
	token := registerAndLogin(t, app, "feeduser", "feed@example.com", "pass123")

	// When: Request feed
	feedReq := httptest.NewRequest("GET", "/api/feed", nil)
	feedReq.Header.Set("Authorization", "Bearer "+token)
	feedResp, err := app.Test(feedReq)

	// Then: Feed is returned
	Expect(err).NotTo(HaveOccurred())
	Expect(feedResp.StatusCode).To(Equal(200))

	var feedResult map[string]interface{}
	json.NewDecoder(feedResp.Body).Decode(&feedResult)
	Expect(feedResult).To(HaveKey("posts"))
}

func TestJWTTokenValidation(t *testing.T) {
	RegisterTestingT(t)

	app, _, cleanup := setupTestApp(t)
	defer cleanup()

	// Given: Invalid token
	invalidReq := httptest.NewRequest("GET", "/api/feed", nil)
	invalidReq.Header.Set("Authorization", "Bearer invalid-token")
	invalidResp, err := app.Test(invalidReq)

	// Then: Should return 401
	Expect(err).NotTo(HaveOccurred())
	Expect(invalidResp.StatusCode).To(Equal(401))

	// Given: Malformed Authorization header
	malformedReq := httptest.NewRequest("GET", "/api/feed", nil)
	malformedReq.Header.Set("Authorization", "InvalidFormat token")
	malformedResp, err := app.Test(malformedReq)

	// Then: Should return 401
	Expect(err).NotTo(HaveOccurred())
	Expect(malformedResp.StatusCode).To(Equal(401))

	// Given: Missing Authorization header
	missingReq := httptest.NewRequest("GET", "/api/feed", nil)
	missingResp, err := app.Test(missingReq)

	// Then: Should return 401
	Expect(err).NotTo(HaveOccurred())
	Expect(missingResp.StatusCode).To(Equal(401))
}

func TestPostCreation(t *testing.T) {
	RegisterTestingT(t)

	app, _, cleanup := setupTestApp(t)
	defer cleanup()

	// Given: User is registered and logged in
	token := registerAndLogin(t, app, "postuser", "post@example.com", "pass123")

	// Given: Post data as multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add form fields
	writer.WriteField("title", "Test Post")
	writer.WriteField("caption", "This is a test post")
	writer.WriteField("media_type", "image")

	// Add file
	fileWriter, err := writer.CreateFormFile("media", "test.jpg")
	Expect(err).NotTo(HaveOccurred())

	_, err = fileWriter.Write([]byte("fake image content"))
	Expect(err).NotTo(HaveOccurred())

	err = writer.Close()
	Expect(err).NotTo(HaveOccurred())

	// When: Create post
	postReq := httptest.NewRequest("POST", "/api/posts", &buf)
	postReq.Header.Set("Content-Type", writer.FormDataContentType())
	postReq.Header.Set("Authorization", "Bearer "+token)
	postResp, err := app.Test(postReq)

	// Then: Post creation succeeds
	Expect(err).NotTo(HaveOccurred())
	Expect(postResp.StatusCode).To(Equal(201))

	var postResult map[string]interface{}
	json.NewDecoder(postResp.Body).Decode(&postResult)
	Expect(postResult).To(HaveKey("id"))
	Expect(postResult["title"]).To(Equal("Test Post"))
	Expect(postResult["caption"]).To(Equal("This is a test post"))
}

func TestLikePost(t *testing.T) {
	RegisterTestingT(t)

	app, _, cleanup := setupTestApp(t)
	defer cleanup()

	// Given: User is registered and logged in
	token := registerAndLogin(t, app, "likeuser", "like@example.com", "pass123")

	// Given: A post exists (we'll create one)
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	writer.WriteField("title", "Post to Like")
	writer.WriteField("caption", "This post will be liked")
	writer.WriteField("media_type", "image")

	fileWriter, err := writer.CreateFormFile("media", "test.jpg")
	Expect(err).NotTo(HaveOccurred())

	_, err = fileWriter.Write([]byte("fake image content"))
	Expect(err).NotTo(HaveOccurred())

	err = writer.Close()
	Expect(err).NotTo(HaveOccurred())

	postReq := httptest.NewRequest("POST", "/api/posts", &buf)
	postReq.Header.Set("Content-Type", writer.FormDataContentType())
	postReq.Header.Set("Authorization", "Bearer "+token)
	postResp, err := app.Test(postReq)
	Expect(err).NotTo(HaveOccurred())
	Expect(postResp.StatusCode).To(Equal(201))

	var postResult map[string]interface{}
	json.NewDecoder(postResp.Body).Decode(&postResult)
	postID := fmt.Sprintf("%.0f", postResult["id"].(float64))

	// When: Like the post
	likeReq := httptest.NewRequest("POST", "/api/posts/"+postID+"/like", nil)
	likeReq.Header.Set("Authorization", "Bearer "+token)
	likeResp, err := app.Test(likeReq)

	// Then: Like succeeds
	Expect(err).NotTo(HaveOccurred())
	Expect(likeResp.StatusCode).To(Equal(200))

	var likeResult map[string]interface{}
	json.NewDecoder(likeResp.Body).Decode(&likeResult)
	Expect(likeResult).To(HaveKey("message"))
	Expect(likeResult["message"]).To(Equal("post liked"))
}

func TestCommentPost(t *testing.T) {
	RegisterTestingT(t)

	app, _, cleanup := setupTestApp(t)
	defer cleanup()

	// Given: User is registered and logged in
	token := registerAndLogin(t, app, "commentuser", "comment@example.com", "pass123")

	// Given: A post exists
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	writer.WriteField("title", "Post to Comment")
	writer.WriteField("caption", "This post will be commented on")
	writer.WriteField("media_type", "image")

	fileWriter, err := writer.CreateFormFile("media", "test.jpg")
	Expect(err).NotTo(HaveOccurred())

	_, err = fileWriter.Write([]byte("fake image content"))
	Expect(err).NotTo(HaveOccurred())

	err = writer.Close()
	Expect(err).NotTo(HaveOccurred())

	postReq := httptest.NewRequest("POST", "/api/posts", &buf)
	postReq.Header.Set("Content-Type", writer.FormDataContentType())
	postReq.Header.Set("Authorization", "Bearer "+token)
	postResp, err := app.Test(postReq)
	Expect(err).NotTo(HaveOccurred())
	Expect(postResp.StatusCode).To(Equal(201))

	var postResult map[string]interface{}
	json.NewDecoder(postResp.Body).Decode(&postResult)
	postID := fmt.Sprintf("%.0f", postResult["id"].(float64))

	// Given: Comment data
	commentData := map[string]string{
		"content": "This is a test comment",
	}
	commentBody, _ := json.Marshal(commentData)

	// When: Comment on the post
	commentReq := httptest.NewRequest("POST", "/api/posts/"+postID+"/comment", bytes.NewReader(commentBody))
	commentReq.Header.Set("Content-Type", "application/json")
	commentReq.Header.Set("Authorization", "Bearer "+token)
	commentResp, err := app.Test(commentReq)

	// Then: Comment succeeds
	Expect(err).NotTo(HaveOccurred())
	Expect(commentResp.StatusCode).To(Equal(200))

	var commentResult map[string]interface{}
	json.NewDecoder(commentResp.Body).Decode(&commentResult)
	Expect(commentResult).To(HaveKey("message"))
	Expect(commentResult["message"]).To(Equal("comment added"))
}

func TestPostCreationFromURL(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers start automatically!
	app, _, cleanup := setupTestApp(t)
	defer cleanup()

	// Register and login
	token := registerAndLogin(t, app, "urluser", "url@example.com", "password123")

	// Create post from URL using multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	writer.WriteField("title", "Post from URL")
	writer.WriteField("caption", "Downloaded from external URL")
	writer.WriteField("media_url", "https://via.placeholder.com/150.jpg")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/posts", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(201))

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	Expect(result["id"]).To(BeNumerically(">", 0))
	Expect(result["media_url"]).To(ContainSubstring("mock-s3"))
	Expect(result["title"]).To(Equal("Post from URL"))
	Expect(result["caption"]).To(Equal("Downloaded from external URL"))
}
