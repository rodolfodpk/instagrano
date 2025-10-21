package tests

import (
	"fmt"
	"io"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/rodolfodpk/instagrano/internal/domain"
	postgresRepo "github.com/rodolfodpk/instagrano/internal/repository/postgres"
	"github.com/rodolfodpk/instagrano/internal/service"
)

var _ = Describe("PostService", func() {
	Describe("CreatePost", func() {
		It("should create post successfully", func() {
			// Given: Post service setup
			postRepo := postgresRepo.NewPostRepository(sharedContainers.DB)
			mockStorage := NewMockMediaStorage()
			postService := service.NewPostService(postRepo, mockStorage, sharedContainers.Cache, 5*time.Minute)

			// Given: User exists
			user := createTestUser(sharedContainers.DB, "postuser", "post@example.com")

			// Given: Post data
			title := "Test Post"
			caption := "This is a test post"
			mediaType := domain.MediaTypeImage
			fileContent := "fake image content"
			fileReader := strings.NewReader(fileContent)
			filename := "test.jpg"

			// When: Create post
			post, err := postService.CreatePost(user.ID, title, caption, mediaType, fileReader, filename)

			// Then: Post is created successfully
			Expect(err).NotTo(HaveOccurred())
			Expect(post).NotTo(BeNil())
			Expect(post.ID).To(BeNumerically(">", 0))
			Expect(post.UserID).To(Equal(user.ID))
			Expect(post.Title).To(Equal(title))
			Expect(post.Caption).To(Equal(caption))
			Expect(post.MediaType).To(Equal(mediaType))
			Expect(post.MediaURL).To(ContainSubstring("mock-s3.example.com"))
			Expect(post.MediaURL).To(ContainSubstring("test.jpg"))

			// Verify post was saved to database
			savedPost, err := postRepo.FindByID(post.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(savedPost.Title).To(Equal(title))
			Expect(savedPost.Caption).To(Equal(caption))
		})

		It("should create video post successfully", func() {
			// Given: Post service setup
			postRepo := postgresRepo.NewPostRepository(sharedContainers.DB)
			mockStorage := NewMockMediaStorage()
			postService := service.NewPostService(postRepo, mockStorage, sharedContainers.Cache, 5*time.Minute)

			user := createTestUser(sharedContainers.DB, "videouser", "video@example.com")

			// Given: Video post data
			title := "Test Video"
			caption := "This is a test video"
			mediaType := domain.MediaTypeVideo
			fileContent := "fake video content"
			fileReader := strings.NewReader(fileContent)
			filename := "test.mp4"

			// When: Create video post
			post, err := postService.CreatePost(user.ID, title, caption, mediaType, fileReader, filename)

			// Then: Video post is created successfully
			Expect(err).NotTo(HaveOccurred())
			Expect(post).NotTo(BeNil())
			Expect(post.MediaType).To(Equal(domain.MediaTypeVideo))
			Expect(post.MediaURL).To(ContainSubstring("test.mp4"))

			// Verify video post was saved to database
			savedPost, err := postRepo.FindByID(post.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(savedPost.MediaType).To(Equal(domain.MediaTypeVideo))
		})

		It("should fail with empty title", func() {
			// Given: Post service setup
			postRepo := postgresRepo.NewPostRepository(sharedContainers.DB)
			mockStorage := NewMockMediaStorage()
			postService := service.NewPostService(postRepo, mockStorage, sharedContainers.Cache, 5*time.Minute)

			user := createTestUser(sharedContainers.DB, "emptytitle", "empty@example.com")

			// Given: Post with empty title
			title := ""
			caption := "This post has no title"
			mediaType := domain.MediaTypeImage
			fileReader := strings.NewReader("fake content")
			filename := "test.jpg"

			// When: Create post with empty title
			post, err := postService.CreatePost(user.ID, title, caption, mediaType, fileReader, filename)

			// Then: Creation fails with validation error
			Expect(err).To(HaveOccurred())
			Expect(post).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("invalid input"))
		})

		It("should handle S3 upload failure", func() {
			// Given: Post service setup with failing storage
			postRepo := postgresRepo.NewPostRepository(sharedContainers.DB)
			failingStorage := NewFailingMockStorage()
			postService := service.NewPostService(postRepo, failingStorage, sharedContainers.Cache, 5*time.Minute)

			user := createTestUser(sharedContainers.DB, "failuser", "fail@example.com")

			// Given: Valid post data
			title := "Test Post"
			caption := "This will fail S3 upload"
			mediaType := domain.MediaTypeImage
			fileReader := strings.NewReader("fake content")
			filename := "test.jpg"

			// When: Create post (S3 upload will fail)
			post, err := postService.CreatePost(user.ID, title, caption, mediaType, fileReader, filename)

			// Then: Creation fails due to S3 error
			Expect(err).To(HaveOccurred())
			Expect(post).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to upload file to S3"))

			// Verify no post was saved to database
			posts, err := postRepo.GetFeed(10, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(posts).To(HaveLen(0))
		})

		It("should handle large files", func() {
			// Given: Post service setup
			postRepo := postgresRepo.NewPostRepository(sharedContainers.DB)
			mockStorage := NewMockMediaStorage()
			postService := service.NewPostService(postRepo, mockStorage, sharedContainers.Cache, 5*time.Minute)

			user := createTestUser(sharedContainers.DB, "largefile", "large@example.com")

			// Given: Large file content (simulate large file)
			largeContent := strings.Repeat("x", 1024*1024) // 1MB
			fileReader := strings.NewReader(largeContent)
			filename := "large-file.jpg"

			// When: Create post with large file
			post, err := postService.CreatePost(user.ID, "Large File Post", "Testing large file upload", domain.MediaTypeImage, fileReader, filename)

			// Then: Post is created successfully
			Expect(err).NotTo(HaveOccurred())
			Expect(post).NotTo(BeNil())
			Expect(post.Title).To(Equal("Large File Post"))

			// Verify the mock storage received the large content
			var foundKey string
			for key := range mockStorage.files {
				if strings.Contains(key, "large-file.jpg") {
					foundKey = key
					break
				}
			}
			Expect(foundKey).NotTo(BeEmpty())
			storedContent := mockStorage.files[foundKey]
			Expect(len(storedContent)).To(Equal(1024 * 1024))
		})

		It("should handle special characters", func() {
			// Given: Post service setup
			postRepo := postgresRepo.NewPostRepository(sharedContainers.DB)
			mockStorage := NewMockMediaStorage()
			postService := service.NewPostService(postRepo, mockStorage, sharedContainers.Cache, 5*time.Minute)

			user := createTestUser(sharedContainers.DB, "special", "special@example.com")

			// Given: Post with special characters
			title := "Test Post with émojis 🚀 and spëcial chars"
			caption := "Testing unicode: 你好世界 🌍"
			fileReader := strings.NewReader("fake content")
			filename := "test-émoji-🚀.jpg"

			// When: Create post with special characters
			post, err := postService.CreatePost(user.ID, title, caption, domain.MediaTypeImage, fileReader, filename)

			// Then: Post is created successfully
			Expect(err).NotTo(HaveOccurred())
			Expect(post).NotTo(BeNil())
			Expect(post.Title).To(Equal(title))
			Expect(post.Caption).To(Equal(caption))
			Expect(post.MediaURL).To(ContainSubstring("test-émoji-🚀.jpg"))
		})
	})

	Describe("GetPost", func() {
		It("should retrieve post successfully", func() {
			// Given: Post service setup
			postRepo := postgresRepo.NewPostRepository(sharedContainers.DB)
			mockStorage := NewMockMediaStorage()
			postService := service.NewPostService(postRepo, mockStorage, sharedContainers.Cache, 5*time.Minute)

			// Given: User and post exist
			user := createTestUser(sharedContainers.DB, "getuser", "get@example.com")
			createdPost := createTestPost(sharedContainers.DB, user.ID, "Get Test Post", "Testing post retrieval")

			// When: Get post by ID
			retrievedPost, err := postService.GetPost(createdPost.ID)

			// Then: Post is retrieved successfully
			Expect(err).NotTo(HaveOccurred())
			Expect(retrievedPost).NotTo(BeNil())
			Expect(retrievedPost.ID).To(Equal(createdPost.ID))
			Expect(retrievedPost.Title).To(Equal("Get Test Post"))
			Expect(retrievedPost.Caption).To(Equal("Testing post retrieval"))
			Expect(retrievedPost.UserID).To(Equal(user.ID))
		})

		It("should return error for non-existent post", func() {
			// Given: Post service setup
			postRepo := postgresRepo.NewPostRepository(sharedContainers.DB)
			mockStorage := NewMockMediaStorage()
			postService := service.NewPostService(postRepo, mockStorage, sharedContainers.Cache, 5*time.Minute)

			// When: Get non-existent post
			post, err := postService.GetPost(99999)

			// Then: Error is returned
			Expect(err).To(HaveOccurred())
			Expect(post).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})
	})
})

// Helper types for testing

// FailingMockStorage implements s3.MediaStorage but always fails
type FailingMockStorage struct{}

func NewFailingMockStorage() *FailingMockStorage {
	return &FailingMockStorage{}
}

func (m *FailingMockStorage) Upload(file io.Reader, filename, contentType string) (string, error) {
	return "", fmt.Errorf("mock S3 upload failure")
}

func (m *FailingMockStorage) GetURL(key string) string {
	return "http://failing-mock-s3.example.com/" + key
}

func (m *FailingMockStorage) UploadFromURL(url string) (string, string, error) {
	return "", "", fmt.Errorf("mock S3 upload from URL failure")
}
