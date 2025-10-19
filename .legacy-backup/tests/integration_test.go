package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInstagrano(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Instagrano Tests")
}

var _ = Describe("Unit Tests", func() {
	Describe("Basic functionality", func() {
		It("should pass basic test", func() {
			Expect(true).To(BeTrue())
		})

		It("should validate basic math", func() {
			Expect(2 + 2).To(Equal(4))
		})

		It("should validate string operations", func() {
			Expect("hello" + " world").To(Equal("hello world"))
		})
	})
})

var _ = Describe("Instagrano API Integration Tests", func() {
	var baseURL string
	var client *http.Client

	BeforeEach(func() {
		baseURL = "http://localhost:3000"
		client = &http.Client{Timeout: 10 * time.Second}
	})

	// Note: These tests require the server to be running manually
	// Run: go run ./cmd/app
	// Then run: go test ./tests/... -v -ginkgo.focus="Integration"

	Describe("Home endpoint", func() {
		It("should return welcome message", func() {
			resp, err := client.Get(baseURL + "/")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(200))

			var result map[string]string
			err = json.NewDecoder(resp.Body).Decode(&result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["message"]).To(Equal("🚀 Instagrano!"))
		})
	})

	Describe("Authentication", func() {
		var testEmail = "test@example.com"
		var testUsername = "testuser"
		var testPassword = "password123"

		It("should register a new user", func() {
			userData := map[string]string{
				"username": testUsername,
				"email":    testEmail,
				"password": testPassword,
			}

			jsonData, _ := json.Marshal(userData)
			resp, err := client.Post(baseURL+"/register", "application/json", bytes.NewBuffer(jsonData))
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(200))

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["user"]).NotTo(BeNil())
			Expect(result["token"]).To(Equal("instagrano-" + testUsername))
		})

		It("should reject duplicate user registration", func() {
			userData := map[string]string{
				"username": testUsername,
				"email":    testEmail,
				"password": testPassword,
			}

			jsonData, _ := json.Marshal(userData)
			resp, err := client.Post(baseURL+"/register", "application/json", bytes.NewBuffer(jsonData))
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(409))

			var result map[string]string
			err = json.NewDecoder(resp.Body).Decode(&result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["error"]).To(Equal("User already exists"))
		})

		It("should login with valid credentials", func() {
			loginData := map[string]string{
				"email":    testEmail,
				"password": testPassword,
			}

			jsonData, _ := json.Marshal(loginData)
			resp, err := client.Post(baseURL+"/login", "application/json", bytes.NewBuffer(jsonData))
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(200))

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["user"]).NotTo(BeNil())
			Expect(result["token"]).To(Equal("instagrano-" + testUsername))
		})

		It("should reject invalid login credentials", func() {
			loginData := map[string]string{
				"email":    "nonexistent@example.com",
				"password": "wrongpassword",
			}

			jsonData, _ := json.Marshal(loginData)
			resp, err := client.Post(baseURL+"/login", "application/json", bytes.NewBuffer(jsonData))
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(401))

			var result map[string]string
			err = json.NewDecoder(resp.Body).Decode(&result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["error"]).To(Equal("Invalid credentials"))
		})
	})

	Describe("Post Management", func() {
		It("should upload an image", func() {
			// Create a simple text file as a mock image
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			// Add file field
			fileWriter, err := writer.CreateFormFile("file", "test.jpg")
			Expect(err).NotTo(HaveOccurred())
			fileWriter.Write([]byte("fake image content"))

			// Add caption field
			writer.WriteField("caption", "Test image caption")
			writer.Close()

			resp, err := client.Post(baseURL+"/upload", writer.FormDataContentType(), &buf)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(200))

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["post_id"]).NotTo(BeNil())
			Expect(result["image_url"]).To(ContainSubstring("test.jpg"))
		})

		It("should reject upload without file", func() {
			resp, err := client.Post(baseURL+"/upload", "multipart/form-data", strings.NewReader(""))
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(400))

			var result map[string]string
			err = json.NewDecoder(resp.Body).Decode(&result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["error"]).To(Equal("No file uploaded"))
		})

		It("should get feed", func() {
			resp, err := client.Get(baseURL + "/feed")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(200))

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["posts"]).NotTo(BeNil())
		})
	})

	Describe("Social Features", func() {
		It("should like a post", func() {
			resp, err := client.Post(baseURL+"/posts/1/like", "application/json", strings.NewReader("{}"))
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(200))

			var result map[string]int
			err = json.NewDecoder(resp.Body).Decode(&result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["likes"]).To(BeNumerically(">=", 1))
		})

		It("should return 404 for non-existent post like", func() {
			resp, err := client.Post(baseURL+"/posts/999/like", "application/json", strings.NewReader("{}"))
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(404))

			var result map[string]string
			err = json.NewDecoder(resp.Body).Decode(&result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["error"]).To(Equal("Not found"))
		})

		It("should add a comment to a post", func() {
			commentData := map[string]string{
				"text": "This is a test comment",
			}

			jsonData, _ := json.Marshal(commentData)
			resp, err := client.Post(baseURL+"/posts/1/comments", "application/json", bytes.NewBuffer(jsonData))
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(200))

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["comments"]).NotTo(BeNil())
		})

		It("should reject empty comment", func() {
			commentData := map[string]string{
				"text": "",
			}

			jsonData, _ := json.Marshal(commentData)
			resp, err := client.Post(baseURL+"/posts/1/comments", "application/json", bytes.NewBuffer(jsonData))
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(400))

			var result map[string]string
			err = json.NewDecoder(resp.Body).Decode(&result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["error"]).To(Equal("Comment text is required"))
		})

		It("should reject invalid post ID for comment", func() {
			commentData := map[string]string{
				"text": "This is a test comment",
			}

			jsonData, _ := json.Marshal(commentData)
			resp, err := client.Post(baseURL+"/posts/invalid/comments", "application/json", bytes.NewBuffer(jsonData))
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(400))

			var result map[string]string
			err = json.NewDecoder(resp.Body).Decode(&result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["error"]).To(Equal("Invalid post ID"))
		})
	})

	Describe("API Documentation", func() {
		It("should serve swagger documentation", func() {
			resp, err := client.Get(baseURL + "/docs")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(200))
			Expect(resp.Header.Get("Content-Type")).To(Equal("text/html"))

			body, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(body)).To(ContainSubstring("Instagrano API Documentation"))
		})

		It("should serve swagger JSON", func() {
			resp, err := client.Get(baseURL + "/docs/doc.json")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(200))
			Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			Expect(err).NotTo(HaveOccurred())
			Expect(result["swagger"]).To(Equal("2.0"))
		})
	})
})
