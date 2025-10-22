package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Integration Tests", func() {
	Describe("User Registration and Login", func() {
		It("should register and login successfully", func() {
			// Given: Test app setup
			app, _, cleanup := setupTestApp()
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
			regResp, err := app.Test(regReq, 2000) // 2 second timeout

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
			loginResp, err := app.Test(loginReq, 2000) // 2 second timeout

			// Then: Login succeeds and returns JWT
			Expect(err).NotTo(HaveOccurred())
			Expect(loginResp.StatusCode).To(Equal(200))

			var loginResult map[string]interface{}
			json.NewDecoder(loginResp.Body).Decode(&loginResult)
			Expect(loginResult).To(HaveKey("token"))
			Expect(loginResult["token"]).NotTo(BeEmpty())
		})
	})

	Describe("Feed Endpoint", func() {
		It("should return feed successfully", func() {
			// Given: Test app setup
			app, _, cleanup := setupTestApp()
			defer cleanup()

			// Given: User is registered and logged in
			token := registerAndLogin(app, "feeduser", "feed@example.com", "pass123")

			// When: Request feed
			feedReq := httptest.NewRequest("GET", "/api/feed", nil)
			feedReq.Header.Set("Authorization", "Bearer "+token)
			feedResp, err := app.Test(feedReq, 2000) // 2 second timeout

			// Then: Feed is returned
			Expect(err).NotTo(HaveOccurred())
			Expect(feedResp.StatusCode).To(Equal(200))

			var feedResult map[string]interface{}
			json.NewDecoder(feedResp.Body).Decode(&feedResult)
			Expect(feedResult).To(HaveKey("posts"))
		})
	})

	Describe("JWT Token Validation", func() {
		It("should reject invalid tokens", func() {
			// Given: Test app setup
			app, _, cleanup := setupTestApp()
			defer cleanup()

			// Given: Invalid token
			invalidReq := httptest.NewRequest("GET", "/api/feed", nil)
			invalidReq.Header.Set("Authorization", "Bearer invalid-token")
			invalidResp, err := app.Test(invalidReq, 2000) // 2 second timeout

			// Then: Should return 401
			Expect(err).NotTo(HaveOccurred())
			Expect(invalidResp.StatusCode).To(Equal(401))

			// Given: Malformed Authorization header
			malformedReq := httptest.NewRequest("GET", "/api/feed", nil)
			malformedReq.Header.Set("Authorization", "InvalidFormat token")
			malformedResp, err := app.Test(malformedReq, 2000) // 2 second timeout

			// Then: Should return 401
			Expect(err).NotTo(HaveOccurred())
			Expect(malformedResp.StatusCode).To(Equal(401))

			// Given: Missing Authorization header
			missingReq := httptest.NewRequest("GET", "/api/feed", nil)
			missingResp, err := app.Test(missingReq, 2000) // 2 second timeout

			// Then: Should return 401
			Expect(err).NotTo(HaveOccurred())
			Expect(missingResp.StatusCode).To(Equal(401))
		})
	})

	Describe("Post Creation", func() {
		It("should create post successfully", func() {
			// Given: Test app setup
			app, _, cleanup := setupTestApp()
			defer cleanup()

			// Given: User is registered and logged in
			token := registerAndLogin(app, "postuser", "post@example.com", "pass123")

			// Given: Post data as multipart form
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			// Add form fields
			writer.WriteField("title", "Test Post")
			writer.WriteField("caption", "This is a test post")
			writer.WriteField("media_url", "http://localhost/test/image")

			err := writer.Close()
			Expect(err).NotTo(HaveOccurred())

			// When: Create post
			postReq := httptest.NewRequest("POST", "/api/posts", &buf)
			postReq.Header.Set("Content-Type", writer.FormDataContentType())
			postReq.Header.Set("Authorization", "Bearer "+token)
			postResp, err := app.Test(postReq, 3000) // 3 second timeout for multipart

			// Then: Post creation succeeds
			Expect(err).NotTo(HaveOccurred())
			Expect(postResp.StatusCode).To(Equal(201))

			var postResult map[string]interface{}
			json.NewDecoder(postResp.Body).Decode(&postResult)
			Expect(postResult).To(HaveKey("id"))
			Expect(postResult["title"]).To(Equal("Test Post"))
			Expect(postResult["caption"]).To(Equal("This is a test post"))
		})
	})

	Describe("Post Interactions", func() {
		It("should like post successfully", func() {
			// Given: Test app setup
			app, _, cleanup := setupTestApp()
			defer cleanup()

			// Given: User is registered and logged in
			token := registerAndLogin(app, "likeuser", "like@example.com", "pass123")

			// Given: A post exists (we'll create one)
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			writer.WriteField("title", "Post to Like")
			writer.WriteField("caption", "This post will be liked")
			writer.WriteField("media_url", "http://localhost/test/image")

			err := writer.Close()
			Expect(err).NotTo(HaveOccurred())

			postReq := httptest.NewRequest("POST", "/api/posts", &buf)
			postReq.Header.Set("Content-Type", writer.FormDataContentType())
			postReq.Header.Set("Authorization", "Bearer "+token)
			postResp, err := app.Test(postReq, 3000) // 3 second timeout for multipart
			Expect(err).NotTo(HaveOccurred())
			Expect(postResp.StatusCode).To(Equal(201))

			var postResult map[string]interface{}
			json.NewDecoder(postResp.Body).Decode(&postResult)
			postID := fmt.Sprintf("%.0f", postResult["id"].(float64))

			// When: Like the post
			likeReq := httptest.NewRequest("POST", "/api/posts/"+postID+"/like", nil)
			likeReq.Header.Set("Authorization", "Bearer "+token)
			likeResp, err := app.Test(likeReq, 2000) // 2 second timeout

			// Then: Like succeeds
			Expect(err).NotTo(HaveOccurred())
			Expect(likeResp.StatusCode).To(Equal(200))

			var likeResult map[string]interface{}
			json.NewDecoder(likeResp.Body).Decode(&likeResult)
			Expect(likeResult).To(HaveKey("post_id"))
			Expect(likeResult).To(HaveKey("likes_count"))
			Expect(likeResult["post_id"]).To(Equal(float64(1)))
			Expect(likeResult["likes_count"]).To(Equal(float64(1)))
		})

		It("should comment on post successfully", func() {
			// Given: Test app setup
			app, _, cleanup := setupTestApp()
			defer cleanup()

			// Given: User is registered and logged in
			token := registerAndLogin(app, "commentuser", "comment@example.com", "pass123")

			// Given: A post exists
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			writer.WriteField("title", "Post to Comment")
			writer.WriteField("caption", "This post will be commented on")
			writer.WriteField("media_url", "http://localhost/test/image")

			err := writer.Close()
			Expect(err).NotTo(HaveOccurred())

			postReq := httptest.NewRequest("POST", "/api/posts", &buf)
			postReq.Header.Set("Content-Type", writer.FormDataContentType())
			postReq.Header.Set("Authorization", "Bearer "+token)
			postResp, err := app.Test(postReq, 3000) // 3 second timeout for multipart
			Expect(err).NotTo(HaveOccurred())
			Expect(postResp.StatusCode).To(Equal(201))

			var postResult map[string]interface{}
			json.NewDecoder(postResp.Body).Decode(&postResult)
			postID := fmt.Sprintf("%.0f", postResult["id"].(float64))

			// Given: Comment data
			commentData := map[string]string{
				"text": "This is a test comment",
			}
			commentBody, _ := json.Marshal(commentData)

			// When: Comment on the post
			commentReq := httptest.NewRequest("POST", "/api/posts/"+postID+"/comment", bytes.NewReader(commentBody))
			commentReq.Header.Set("Content-Type", "application/json")
			commentReq.Header.Set("Authorization", "Bearer "+token)
			commentResp, err := app.Test(commentReq, 2000) // 2 second timeout

			// Then: Comment succeeds
			Expect(err).NotTo(HaveOccurred())
			Expect(commentResp.StatusCode).To(Equal(200))

			var commentResult map[string]interface{}
			json.NewDecoder(commentResp.Body).Decode(&commentResult)
			Expect(commentResult).To(HaveKey("post_id"))
			Expect(commentResult).To(HaveKey("comments_count"))
			Expect(commentResult["post_id"]).To(Equal(float64(1)))
			Expect(commentResult["comments_count"]).To(Equal(float64(1)))
		})
	})
})
