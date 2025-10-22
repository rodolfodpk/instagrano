package tests

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	// Use local static test image endpoints
	testImageURL = "http://localhost/test/image"
	testPNGURL   = "http://localhost/test/image" // Same image for simplicity
	// URLs that should fail
	testNotFoundURL = "http://localhost/nonexistent"
	testInvalidURL  = "not-a-valid-url"
)

var _ = Describe("PostService URL Tests", func() {
	Describe("CreatePostFromURL", func() {
		It("should create post from image URL", func() {
			// Given: Test app setup
			app, _, cleanup := setupTestApp()
			defer cleanup()

			// Given: User is registered and logged in
			token := registerAndLogin(app, "urluser", "url@example.com", "pass123")

			// Given: Post data as multipart form
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			// Add form fields
			writer.WriteField("title", "Test Post from URL")
			writer.WriteField("caption", "This is a test post from URL")
			writer.WriteField("media_url", testImageURL)

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
			Expect(postResult["title"]).To(Equal("Test Post from URL"))
			Expect(postResult["caption"]).To(Equal("This is a test post from URL"))
		})

		It("should create post from PNG URL", func() {
			// Given: Test app setup
			app, _, cleanup := setupTestApp()
			defer cleanup()

			// Given: User is registered and logged in
			token := registerAndLogin(app, "pnguser", "png@example.com", "pass123")

			// Given: Post data as multipart form
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			// Add form fields
			writer.WriteField("title", "PNG Post")
			writer.WriteField("caption", "PNG image post")
			writer.WriteField("media_url", testPNGURL)

			err := writer.Close()
			Expect(err).NotTo(HaveOccurred())

			// When: Create post
			postReq := httptest.NewRequest("POST", "/api/posts", &buf)
			postReq.Header.Set("Content-Type", writer.FormDataContentType())
			postReq.Header.Set("Authorization", "Bearer "+token)
			postResp, err := app.Test(postReq, 3000)

			// Then: Post creation succeeds
			Expect(err).NotTo(HaveOccurred())
			Expect(postResp.StatusCode).To(Equal(201))

			var postResult map[string]interface{}
			json.NewDecoder(postResp.Body).Decode(&postResult)
			Expect(postResult).To(HaveKey("id"))
			Expect(postResult["title"]).To(Equal("PNG Post"))
		})

		It("should fail with invalid URL", func() {
			// Given: Test app setup
			app, _, cleanup := setupTestApp()
			defer cleanup()

			// Given: User is registered and logged in
			token := registerAndLogin(app, "invaliduser", "invalid@example.com", "pass123")

			// Given: Post data with invalid URL
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			writer.WriteField("title", "Invalid URL Post")
			writer.WriteField("media_url", testInvalidURL)

			err := writer.Close()
			Expect(err).NotTo(HaveOccurred())

			// When: Create post
			postReq := httptest.NewRequest("POST", "/api/posts", &buf)
			postReq.Header.Set("Content-Type", writer.FormDataContentType())
			postReq.Header.Set("Authorization", "Bearer "+token)
			postResp, err := app.Test(postReq, 3000)

			// Then: Post creation fails
			Expect(err).NotTo(HaveOccurred())
			Expect(postResp.StatusCode).To(Equal(400))
		})

		It("should fail with download failure", func() {
			// Given: Test app setup
			app, _, cleanup := setupTestApp()
			defer cleanup()

			// Given: User is registered and logged in
			token := registerAndLogin(app, "failuser", "fail@example.com", "pass123")

			// Given: Post data with non-existent URL that won't be rewritten by mock controller
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			writer.WriteField("title", "Failed Download Post")
			writer.WriteField("media_url", "http://nonexistent-domain-that-does-not-exist.com/image.jpg")

			err := writer.Close()
			Expect(err).NotTo(HaveOccurred())

			// When: Create post
			postReq := httptest.NewRequest("POST", "/api/posts", &buf)
			postReq.Header.Set("Content-Type", writer.FormDataContentType())
			postReq.Header.Set("Authorization", "Bearer "+token)
			postResp, err := app.Test(postReq, 3000)

			// Then: Post creation fails
			Expect(err).NotTo(HaveOccurred())
			Expect(postResp.StatusCode).To(Equal(400))
		})

		It("should fail with empty title", func() {
			// Given: Test app setup
			app, _, cleanup := setupTestApp()
			defer cleanup()

			// Given: User is registered and logged in
			token := registerAndLogin(app, "emptytitle", "empty@example.com", "pass123")

			// Given: Post data without title
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			writer.WriteField("caption", "Post without title")
			writer.WriteField("media_url", testImageURL)

			err := writer.Close()
			Expect(err).NotTo(HaveOccurred())

			// When: Create post
			postReq := httptest.NewRequest("POST", "/api/posts", &buf)
			postReq.Header.Set("Content-Type", writer.FormDataContentType())
			postReq.Header.Set("Authorization", "Bearer "+token)
			postResp, err := app.Test(postReq, 3000)

			// Then: Post creation fails
			Expect(err).NotTo(HaveOccurred())
			Expect(postResp.StatusCode).To(Equal(400))
		})

		It("should fail with empty URL", func() {
			// Given: Test app setup
			app, _, cleanup := setupTestApp()
			defer cleanup()

			// Given: User is registered and logged in
			token := registerAndLogin(app, "emptyurl", "emptyurl@example.com", "pass123")

			// Given: Post data without media_url
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			writer.WriteField("title", "Post without URL")
			writer.WriteField("caption", "This post has no media URL")

			err := writer.Close()
			Expect(err).NotTo(HaveOccurred())

			// When: Create post
			postReq := httptest.NewRequest("POST", "/api/posts", &buf)
			postReq.Header.Set("Content-Type", writer.FormDataContentType())
			postReq.Header.Set("Authorization", "Bearer "+token)
			postResp, err := app.Test(postReq, 3000)

			// Then: Post creation fails (no media provided)
			Expect(err).NotTo(HaveOccurred())
			Expect(postResp.StatusCode).To(Equal(400))
		})
	})
})
