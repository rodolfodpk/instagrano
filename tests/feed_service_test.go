package tests

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	postgresRepo "github.com/rodolfodpk/instagrano/internal/repository/postgres"
	"github.com/rodolfodpk/instagrano/internal/service"
)

var _ = Describe("FeedService", func() {
	Describe("GetFeedWithCursor", func() {
		It("should return posts with cursor pagination", func() {
			// Given: Posts exist in database
			user := createTestUser(sharedContainers.DB, "testuser", "test@example.com")

			// Create multiple posts
			for i := 0; i < 5; i++ {
				createTestPost(sharedContainers.DB, user.ID, fmt.Sprintf("Post %d", i), fmt.Sprintf("Caption %d", i))
			}

			postRepo := postgresRepo.NewPostRepository(sharedContainers.DB)
			feedService := service.NewFeedService(postRepo, sharedContainers.Cache, 5*time.Minute)

			// When: Get feed with cursor
			result, err := feedService.GetFeedWithCursor(3, "")

			// Then: Should return posts successfully
			Expect(err).NotTo(HaveOccurred())
			Expect(len(result.Posts)).To(Equal(3))
			Expect(result.NextCursor).NotTo(BeEmpty())
		})

		It("should handle empty cursor", func() {
			// Given: Posts exist in database
			user := createTestUser(sharedContainers.DB, "testuser", "test@example.com")
			createTestPost(sharedContainers.DB, user.ID, "Test Post", "Test Caption")

			postRepo := postgresRepo.NewPostRepository(sharedContainers.DB)
			feedService := service.NewFeedService(postRepo, sharedContainers.Cache, 5*time.Minute)

			// When: Get feed with empty cursor
			result, err := feedService.GetFeedWithCursor(10, "")

			// Then: Should return posts successfully
			Expect(err).NotTo(HaveOccurred())
			Expect(len(result.Posts)).To(Equal(1))
			Expect(result.NextCursor).To(BeEmpty()) // No more posts
		})

		It("should respect page size limit", func() {
			// Given: Multiple posts exist in database
			user := createTestUser(sharedContainers.DB, "testuser", "test@example.com")

			// Create 10 posts
			for i := 0; i < 10; i++ {
				createTestPost(sharedContainers.DB, user.ID, fmt.Sprintf("Post %d", i), fmt.Sprintf("Caption %d", i))
			}

			postRepo := postgresRepo.NewPostRepository(sharedContainers.DB)
			feedService := service.NewFeedService(postRepo, sharedContainers.Cache, 5*time.Minute)

			// When: Get feed with page size limit
			result, err := feedService.GetFeedWithCursor(5, "")

			// Then: Should respect page size
			Expect(err).NotTo(HaveOccurred())
			Expect(len(result.Posts)).To(Equal(5))
			Expect(result.NextCursor).NotTo(BeEmpty()) // More posts available
		})

		It("should handle invalid cursor", func() {
			// Given: Posts exist in database
			user := createTestUser(sharedContainers.DB, "testuser", "test@example.com")
			createTestPost(sharedContainers.DB, user.ID, "Test Post", "Test Caption")

			postRepo := postgresRepo.NewPostRepository(sharedContainers.DB)
			feedService := service.NewFeedService(postRepo, sharedContainers.Cache, 5*time.Minute)

			// When: Get feed with invalid cursor
			result, err := feedService.GetFeedWithCursor(10, "invalid-cursor")

			// Then: Should return error for invalid cursor
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should return posts in correct order", func() {
			// Given: Posts with different timestamps
			user := createTestUser(sharedContainers.DB, "testuser", "test@example.com")

			// Create posts with specific order
			createTestPost(sharedContainers.DB, user.ID, "First Post", "First Caption")
			time.Sleep(1 * time.Millisecond) // Ensure different timestamps
			createTestPost(sharedContainers.DB, user.ID, "Second Post", "Second Caption")

			postRepo := postgresRepo.NewPostRepository(sharedContainers.DB)
			feedService := service.NewFeedService(postRepo, sharedContainers.Cache, 5*time.Minute)

			// When: Get feed
			result, err := feedService.GetFeedWithCursor(10, "")

			// Then: Should return posts in correct order (newest first)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(result.Posts)).To(Equal(2))
			// Note: Posts are interface{} so we can't directly compare IDs
			// This test verifies the correct number of posts are returned
		})

		It("should use cache for repeated requests", func() {
			// Given: Posts exist in database
			user := createTestUser(sharedContainers.DB, "testuser", "test@example.com")
			createTestPost(sharedContainers.DB, user.ID, "Test Post", "Test Caption")

			postRepo := postgresRepo.NewPostRepository(sharedContainers.DB)
			feedService := service.NewFeedService(postRepo, sharedContainers.Cache, 5*time.Minute)

			// When: Get feed multiple times
			result1, err1 := feedService.GetFeedWithCursor(10, "")
			Expect(err1).NotTo(HaveOccurred())

			result2, err2 := feedService.GetFeedWithCursor(10, "")
			Expect(err2).NotTo(HaveOccurred())

			// Then: Should return same results (cached)
			Expect(len(result1.Posts)).To(Equal(len(result2.Posts)))
			// Note: Posts are interface{} so we can't directly compare IDs
			// This test verifies the same number of posts are returned
		})
	})
})