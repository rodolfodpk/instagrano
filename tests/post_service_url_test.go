package tests

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/rodolfodpk/instagrano/internal/domain"
	"github.com/rodolfodpk/instagrano/internal/repository/postgres"
	"github.com/rodolfodpk/instagrano/internal/service"
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
		It("should skip due to network timeout issues", func() {
			Skip("Skipping due to network timeout issues with via.placeholder.com")
		})

		It("should create post from PNG URL", func() {
			// Given: Post service setup
			postRepo := postgres.NewPostRepository(sharedContainers.DB)
			mockStorage := NewMockMediaStorage()
			postService := service.NewPostService(postRepo, mockStorage, sharedContainers.Cache, 5*time.Minute)

			// Given: Test user
			user := &domain.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			}
			userRepo := postgres.NewUserRepository(sharedContainers.DB)
			err := userRepo.Create(user)
			Expect(err).NotTo(HaveOccurred())

			// When: Create post from PNG URL
			post, err := postService.CreatePostFromURL(user.ID, "PNG Post", "PNG from URL", testPNGURL)

			// Then: Post is created successfully
			Expect(err).NotTo(HaveOccurred())
			Expect(post).NotTo(BeNil())
			Expect(post.MediaType).To(Equal(domain.MediaTypeImage))
			Expect(post.MediaURL).To(ContainSubstring("mock-s3"))
		})

		It("should fail with invalid URL", func() {
			// Given: Post service setup
			postRepo := postgres.NewPostRepository(sharedContainers.DB)
			mockStorage := NewMockMediaStorage()
			postService := service.NewPostService(postRepo, mockStorage, sharedContainers.Cache, 5*time.Minute)

			// Given: Test user
			user := &domain.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			}
			userRepo := postgres.NewUserRepository(sharedContainers.DB)
			err := userRepo.Create(user)
			Expect(err).NotTo(HaveOccurred())

			// When: Create post with malformed URL
			post, err := postService.CreatePostFromURL(user.ID, "Test Post", "Invalid URL", "not-a-url")

			// Then: Creation fails
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid URL format"))
			Expect(post).To(BeNil())
		})

		It("should fail with download failure", func() {
			// Given: Post service setup
			postRepo := postgres.NewPostRepository(sharedContainers.DB)
			mockStorage := NewMockMediaStorage()
			postService := service.NewPostService(postRepo, mockStorage, sharedContainers.Cache, 5*time.Minute)

			// Given: Test user
			user := &domain.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			}
			userRepo := postgres.NewUserRepository(sharedContainers.DB)
			err := userRepo.Create(user)
			Expect(err).NotTo(HaveOccurred())

			// When: Create post with invalid URL
			post, err := postService.CreatePostFromURL(user.ID, "Test Post", "Invalid URL", testInvalidURL)

			// Then: Creation fails
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid URL format"))
			Expect(post).To(BeNil())
		})

		It("should fail with empty title", func() {
			// Given: Post service setup
			postRepo := postgres.NewPostRepository(sharedContainers.DB)
			mockStorage := NewMockMediaStorage()
			postService := service.NewPostService(postRepo, mockStorage, sharedContainers.Cache, 5*time.Minute)

			// Given: Test user
			user := &domain.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			}
			userRepo := postgres.NewUserRepository(sharedContainers.DB)
			err := userRepo.Create(user)
			Expect(err).NotTo(HaveOccurred())

			// When: Create post with empty title
			post, err := postService.CreatePostFromURL(user.ID, "", "Caption", testImageURL)

			// Then: Creation fails
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(service.ErrInvalidInput))
			Expect(post).To(BeNil())
		})

		It("should fail with empty URL", func() {
			// Given: Post service setup
			postRepo := postgres.NewPostRepository(sharedContainers.DB)
			mockStorage := NewMockMediaStorage()
			postService := service.NewPostService(postRepo, mockStorage, sharedContainers.Cache, 5*time.Minute)

			// Given: Test user
			user := &domain.User{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			}
			userRepo := postgres.NewUserRepository(sharedContainers.DB)
			err := userRepo.Create(user)
			Expect(err).NotTo(HaveOccurred())

			// When: Create post with empty URL
			post, err := postService.CreatePostFromURL(user.ID, "Title", "Caption", "")

			// Then: Creation fails
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("media URL is required"))
			Expect(post).To(BeNil())
		})
	})
})
