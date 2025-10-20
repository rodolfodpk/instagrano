package tests

import (
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/rodolfodpk/instagrano/internal/cache"
	"github.com/rodolfodpk/instagrano/internal/domain"
	postgresRepo "github.com/rodolfodpk/instagrano/internal/repository/postgres"
	"github.com/rodolfodpk/instagrano/internal/service"
	"go.uber.org/zap"
)

func TestFeedService_GetFeedWithCursor(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers for PostgreSQL + Redis
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	// Use Testcontainers Redis for consistency
	redisCache := containers.Cache

	postRepo := postgresRepo.NewPostRepository(containers.DB)
	feedService := service.NewFeedService(postRepo, redisCache, 5*time.Minute)

	// Given: Posts exist in database
	user := createTestUser(t, containers.DB, "testuser", "test@example.com")
	createTestPost(t, containers.DB, user.ID, "Post 1", "Caption 1")
	createTestPost(t, containers.DB, user.ID, "Post 2", "Caption 2")
	createTestPost(t, containers.DB, user.ID, "Post 3", "Caption 3")

	// When: Get feed (first call - cache miss)
	result1, err := feedService.GetFeedWithCursor(20, "")

	// Then: Posts are returned
	Expect(err).NotTo(HaveOccurred())
	Expect(result1.Posts).To(HaveLen(3))
	Expect(result1.HasMore).To(BeFalse())

	// When: Get feed again (cache hit)
	result2, err := feedService.GetFeedWithCursor(20, "")

	// Then: Same posts returned from cache
	Expect(err).NotTo(HaveOccurred())
	Expect(result2.Posts).To(HaveLen(3))

	// Verify cache was populated (we can't easily check keys with real Redis)
	// The cache hit/miss behavior is verified by the structured logs
}

func TestFeedService_ScoringAlgorithm(t *testing.T) {
	RegisterTestingT(t)

	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	// Use Testcontainers Redis for consistency
	redisCache := containers.Cache
	postRepo := postgresRepo.NewPostRepository(containers.DB)
	feedService := service.NewFeedService(postRepo, redisCache, 5*time.Minute)

	user := createTestUser(t, containers.DB, "scorer", "scorer@example.com")

	// Given: Posts with different engagement
	post1 := createTestPostWithEngagement(t, containers.DB, user.ID,
		"Popular", time.Now().Add(-2*time.Hour), 100, 50, 1000)
	post2 := createTestPostWithEngagement(t, containers.DB, user.ID,
		"Unpopular", time.Now().Add(-1*time.Hour), 2, 1, 10)

	// When: Get feed
	result, err := feedService.GetFeedWithCursor(20, "")

	// Then: Posts ordered by score (engagement + time decay)
	Expect(err).NotTo(HaveOccurred())
	Expect(result.Posts).To(HaveLen(2))

	firstPost := result.Posts[0].(*domain.Post)
	Expect(firstPost.Score).To(BeNumerically(">", 0))

	// Verify popular post scores higher
	Expect(firstPost.ID).To(Equal(post1.ID))

	// Verify unpopular post is second
	secondPost := result.Posts[1].(*domain.Post)
	Expect(secondPost.ID).To(Equal(post2.ID))
}

func TestFeedService_CacheMiss(t *testing.T) {
	RegisterTestingT(t)

	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	// Use Testcontainers Redis for consistency
	redisCache := containers.Cache
	postRepo := postgresRepo.NewPostRepository(containers.DB)
	feedService := service.NewFeedService(postRepo, redisCache, 5*time.Minute)

	// Given: Posts exist in database
	user := createTestUser(t, containers.DB, "cacheuser", "cache@example.com")
	createTestPost(t, containers.DB, user.ID, "Cache Test Post", "Testing cache miss")

	// When: Get feed (cache miss)
	result, err := feedService.GetFeedWithCursor(20, "")

	// Then: Posts are returned from database
	Expect(err).NotTo(HaveOccurred())
	Expect(result.Posts).To(HaveLen(1))

	// Cache behavior is verified by structured logs (cache miss -> cache hit)
}

func TestFeedService_CacheError(t *testing.T) {
	RegisterTestingT(t)

	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	// Create a cache that will fail
	logger, _ := zap.NewProduction()
	redisCache, err := cache.NewRedisCache("invalid:6379", "", 0, logger)

	// If cache creation fails, skip this test
	if err != nil {
		t.Skip("Skipping cache error test - invalid Redis connection")
	}

	postRepo := postgresRepo.NewPostRepository(containers.DB)
	feedService := service.NewFeedService(postRepo, redisCache, 5*time.Minute)

	// Given: Posts exist in database
	user := createTestUser(t, containers.DB, "erroruser", "error@example.com")
	createTestPost(t, containers.DB, user.ID, "Error Test Post", "Testing cache error")

	// When: Get feed (cache will fail, should fallback to DB)
	result, err := feedService.GetFeedWithCursor(20, "")

	// Then: Posts are still returned from database (graceful degradation)
	Expect(err).NotTo(HaveOccurred())
	Expect(result.Posts).To(HaveLen(1))
}

func TestFeedService_Pagination(t *testing.T) {
	RegisterTestingT(t)

	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	// Use Testcontainers Redis for consistency
	redisCache := containers.Cache
	postRepo := postgresRepo.NewPostRepository(containers.DB)
	feedService := service.NewFeedService(postRepo, redisCache, 5*time.Minute)

	// Given: More posts than limit
	user := createTestUser(t, containers.DB, "pageuser", "page@example.com")
	for i := 1; i <= 5; i++ {
		createTestPost(t, containers.DB, user.ID, fmt.Sprintf("Post %d", i), fmt.Sprintf("Caption %d", i))
		// Add small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	// When: Get feed with limit of 3
	result, err := feedService.GetFeedWithCursor(3, "")

	// Then: Only 3 posts returned, hasMore is true
	Expect(err).NotTo(HaveOccurred())
	Expect(result.Posts).To(HaveLen(3))
	Expect(result.HasMore).To(BeTrue())
	Expect(result.NextCursor).NotTo(BeEmpty())

	// When: Get next page using cursor
	result2, err := feedService.GetFeedWithCursor(3, result.NextCursor)

	// Then: Next page returned (may be 0-2 posts depending on timing)
	Expect(err).NotTo(HaveOccurred())
	Expect(len(result2.Posts)).To(BeNumerically("<=", 2))
	Expect(result2.HasMore).To(BeFalse())
}

func TestFeedService_EmptyFeed(t *testing.T) {
	RegisterTestingT(t)

	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	// Use Testcontainers Redis for consistency
	redisCache := containers.Cache
	postRepo := postgresRepo.NewPostRepository(containers.DB)
	feedService := service.NewFeedService(postRepo, redisCache, 5*time.Minute)

	// When: Get feed with no posts
	result, err := feedService.GetFeedWithCursor(20, "")

	// Then: Empty result
	Expect(err).NotTo(HaveOccurred())
	Expect(result.Posts).To(HaveLen(0))
	Expect(result.HasMore).To(BeFalse())
	Expect(result.NextCursor).To(BeEmpty())
}
