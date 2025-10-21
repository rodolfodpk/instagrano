package tests

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/rodolfodpk/instagrano/internal/domain"
	"github.com/rodolfodpk/instagrano/internal/repository/postgres"
	"github.com/rodolfodpk/instagrano/internal/service"
)

const (
	// Use external URLs that will be mapped to mock controller
	testImageURL    = "https://via.placeholder.com/150.jpg"
	testPNGURL      = "https://httpbin.org/image/png"
	// URLs that should fail (not mapped by webclient)
	testNotFoundURL = "https://nonexistent-domain-12345.com/notfound.jpg"
	testTimeoutURL  = "https://httpbin.org/delay/10"
)

func TestPostService_CreatePostFromURL(t *testing.T) {
	RegisterTestingT(t)

	// Setup test containers
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	// Setup mock storage
	mockStorage := NewMockMediaStorage()

	// Setup post repository
	postRepo := postgres.NewPostRepository(containers.DB)

	// Setup post service
	postService := service.NewPostService(postRepo, mockStorage, containers.Cache, 5*time.Minute)

	// Create test user
	user := &domain.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	userRepo := postgres.NewUserRepository(containers.DB)
	err := userRepo.Create(user)
	Expect(err).NotTo(HaveOccurred())

	// Test valid image URL
	post, err := postService.CreatePostFromURL(user.ID, "Test Post", "From URL", testImageURL)
	Expect(err).NotTo(HaveOccurred())
	Expect(post).NotTo(BeNil())
	Expect(post.Title).To(Equal("Test Post"))
	Expect(post.Caption).To(Equal("From URL"))
	Expect(post.MediaType).To(Equal(domain.MediaTypeImage))
	Expect(post.MediaURL).To(ContainSubstring("mock-s3"))
	Expect(post.UserID).To(Equal(user.ID))

	// Verify file was stored in mock
	key := "mock-s3/150.jpg"  // filepath.Base(testImageURL)
	content, exists := mockStorage.GetFile(key)
	Expect(exists).To(BeTrue())
	Expect(len(content)).To(BeNumerically(">", 0))
}

func TestPostService_CreatePostFromURLVideo(t *testing.T) {
	RegisterTestingT(t)

	// Setup test containers
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	// Setup mock storage
	mockStorage := NewMockMediaStorage()

	// Setup post repository
	postRepo := postgres.NewPostRepository(containers.DB)

	// Setup post service
	postService := service.NewPostService(postRepo, mockStorage, containers.Cache, 5*time.Minute)

	// Create test user
	user := &domain.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	userRepo := postgres.NewUserRepository(containers.DB)
	err := userRepo.Create(user)
	Expect(err).NotTo(HaveOccurred())

	// Test PNG URL (should still be treated as image)
	post, err := postService.CreatePostFromURL(user.ID, "PNG Post", "PNG from URL", testPNGURL)
	Expect(err).NotTo(HaveOccurred())
	Expect(post).NotTo(BeNil())
	Expect(post.MediaType).To(Equal(domain.MediaTypeImage))
	Expect(post.MediaURL).To(ContainSubstring("mock-s3"))
}

func TestPostService_CreatePostFromURLInvalidURL(t *testing.T) {
	RegisterTestingT(t)

	// Setup test containers
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	// Setup mock storage
	mockStorage := NewMockMediaStorage()

	// Setup post repository
	postRepo := postgres.NewPostRepository(containers.DB)

	// Setup post service
	postService := service.NewPostService(postRepo, mockStorage, containers.Cache, 5*time.Minute)

	// Create test user
	user := &domain.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	userRepo := postgres.NewUserRepository(containers.DB)
	err := userRepo.Create(user)
	Expect(err).NotTo(HaveOccurred())

	// Test malformed URL
	post, err := postService.CreatePostFromURL(user.ID, "Test Post", "Invalid URL", "not-a-url")
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("unsupported protocol scheme"))
	Expect(post).To(BeNil())
}

func TestPostService_CreatePostFromURLDownloadFailure(t *testing.T) {
	RegisterTestingT(t)

	// Setup test containers
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	// Setup mock storage
	mockStorage := NewMockMediaStorage()

	// Setup post repository
	postRepo := postgres.NewPostRepository(containers.DB)

	// Setup post service
	postService := service.NewPostService(postRepo, mockStorage, containers.Cache, 5*time.Minute)

	// Create test user
	user := &domain.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	userRepo := postgres.NewUserRepository(containers.DB)
	err := userRepo.Create(user)
	Expect(err).NotTo(HaveOccurred())

	// Test 404 URL
	post, err := postService.CreatePostFromURL(user.ID, "Test Post", "404 URL", testNotFoundURL)
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("no such host"))
	Expect(post).To(BeNil())
}

func TestPostService_CreatePostFromURLEmptyTitle(t *testing.T) {
	RegisterTestingT(t)

	// Setup test containers
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	// Setup mock storage
	mockStorage := NewMockMediaStorage()

	// Setup post repository
	postRepo := postgres.NewPostRepository(containers.DB)

	// Setup post service
	postService := service.NewPostService(postRepo, mockStorage, containers.Cache, 5*time.Minute)

	// Create test user
	user := &domain.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	userRepo := postgres.NewUserRepository(containers.DB)
	err := userRepo.Create(user)
	Expect(err).NotTo(HaveOccurred())

	// Test empty title
	post, err := postService.CreatePostFromURL(user.ID, "", "Caption", testImageURL)
	Expect(err).To(HaveOccurred())
	Expect(err).To(Equal(service.ErrInvalidInput))
	Expect(post).To(BeNil())
}

func TestPostService_CreatePostFromURLEmptyURL(t *testing.T) {
	RegisterTestingT(t)

	// Setup test containers
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	// Setup mock storage
	mockStorage := NewMockMediaStorage()

	// Setup post repository
	postRepo := postgres.NewPostRepository(containers.DB)

	// Setup post service
	postService := service.NewPostService(postRepo, mockStorage, containers.Cache, 5*time.Minute)

	// Create test user
	user := &domain.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	userRepo := postgres.NewUserRepository(containers.DB)
	err := userRepo.Create(user)
	Expect(err).NotTo(HaveOccurred())

	// Test empty URL
	post, err := postService.CreatePostFromURL(user.ID, "Title", "Caption", "")
	Expect(err).To(HaveOccurred())
	Expect(err.Error()).To(ContainSubstring("media URL is required"))
	Expect(post).To(BeNil())
}
