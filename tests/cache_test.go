package tests

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/rodolfodpk/instagrano/internal/cache"
	"go.uber.org/zap"
)

func TestRedisCache_NewRedisCache(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers Redis
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	// Get Redis connection string
	redisAddr, err := containers.RedisContainer.ConnectionString(context.Background())
	Expect(err).NotTo(HaveOccurred())

	// Remove redis:// prefix if present
	if strings.HasPrefix(redisAddr, "redis://") {
		redisAddr = strings.TrimPrefix(redisAddr, "redis://")
	}

	// Given: Valid Redis configuration
	logger, _ := zap.NewProduction()

	// When: Create Redis cache
	redisCache, err := cache.NewRedisCache(redisAddr, "", 0, logger)

	// Then: Should create successfully
	Expect(err).NotTo(HaveOccurred())
	Expect(redisCache).NotTo(BeNil())
}

func TestRedisCache_NewRedisCacheInvalidAddress(t *testing.T) {
	RegisterTestingT(t)

	// Given: Invalid Redis address
	logger, _ := zap.NewProduction()

	// When: Try to create Redis cache with invalid address
	redisCache, err := cache.NewRedisCache("invalid:6379", "", 0, logger)

	// Then: Should return an error
	Expect(err).To(HaveOccurred())
	Expect(redisCache).To(BeNil())
	Expect(err.Error()).To(ContainSubstring("failed to connect to redis"))
}

func TestRedisCache_SetAndGet(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers Redis
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	redisAddr, err := containers.RedisContainer.ConnectionString(context.Background())
	Expect(err).NotTo(HaveOccurred())

	// Remove redis:// prefix if present
	if strings.HasPrefix(redisAddr, "redis://") {
		redisAddr = strings.TrimPrefix(redisAddr, "redis://")
	}
	logger, _ := zap.NewProduction()
	redisCache, err := cache.NewRedisCache(redisAddr, "", 0, logger)
	Expect(err).NotTo(HaveOccurred())

	ctx := context.Background()

	// Given: A key-value pair
	key := "test-key"
	value := []byte("test-value")
	ttl := 5 * time.Minute

	// When: Set value in cache
	err = redisCache.Set(ctx, key, value, ttl)

	// Then: Should set successfully
	Expect(err).NotTo(HaveOccurred())

	// When: Get value from cache
	retrievedValue, err := redisCache.Get(ctx, key)

	// Then: Should retrieve the same value
	Expect(err).NotTo(HaveOccurred())
	Expect(retrievedValue).To(Equal(value))
}

func TestRedisCache_GetNonExistent(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers Redis
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	redisAddr, err := containers.RedisContainer.ConnectionString(context.Background())
	Expect(err).NotTo(HaveOccurred())

	// Remove redis:// prefix if present
	if strings.HasPrefix(redisAddr, "redis://") {
		redisAddr = strings.TrimPrefix(redisAddr, "redis://")
	}
	logger, _ := zap.NewProduction()
	redisCache, err := cache.NewRedisCache(redisAddr, "", 0, logger)
	Expect(err).NotTo(HaveOccurred())

	ctx := context.Background()

	// When: Try to get non-existent key
	retrievedValue, err := redisCache.Get(ctx, "non-existent-key")

	// Then: Should return an error
	Expect(err).To(HaveOccurred())
	Expect(retrievedValue).To(BeNil())
	Expect(err.Error()).To(ContainSubstring("not found"))
}

func TestRedisCache_SetMultipleKeys(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers Redis
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	redisAddr, err := containers.RedisContainer.ConnectionString(context.Background())
	Expect(err).NotTo(HaveOccurred())

	// Remove redis:// prefix if present
	if strings.HasPrefix(redisAddr, "redis://") {
		redisAddr = strings.TrimPrefix(redisAddr, "redis://")
	}
	logger, _ := zap.NewProduction()
	redisCache, err := cache.NewRedisCache(redisAddr, "", 0, logger)
	Expect(err).NotTo(HaveOccurred())

	ctx := context.Background()

	// Given: Multiple key-value pairs
	testData := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
		"key3": []byte("value3"),
	}
	ttl := 5 * time.Minute

	// When: Set multiple values
	for key, value := range testData {
		err = redisCache.Set(ctx, key, value, ttl)
		Expect(err).NotTo(HaveOccurred())
	}

	// Then: Should be able to retrieve all values
	for key, expectedValue := range testData {
		retrievedValue, err := redisCache.Get(ctx, key)
		Expect(err).NotTo(HaveOccurred())
		Expect(retrievedValue).To(Equal(expectedValue))
	}
}

func TestRedisCache_SetWithTTL(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers Redis
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	redisAddr, err := containers.RedisContainer.ConnectionString(context.Background())
	Expect(err).NotTo(HaveOccurred())

	// Remove redis:// prefix if present
	if strings.HasPrefix(redisAddr, "redis://") {
		redisAddr = strings.TrimPrefix(redisAddr, "redis://")
	}
	logger, _ := zap.NewProduction()
	redisCache, err := cache.NewRedisCache(redisAddr, "", 0, logger)
	Expect(err).NotTo(HaveOccurred())

	ctx := context.Background()

	// Given: A key-value pair with short TTL
	key := "ttl-test-key"
	value := []byte("ttl-test-value")
	ttl := 100 * time.Millisecond // Very short TTL

	// When: Set value with TTL
	err = redisCache.Set(ctx, key, value, ttl)
	Expect(err).NotTo(HaveOccurred())

	// Then: Should be able to retrieve immediately
	retrievedValue, err := redisCache.Get(ctx, key)
	Expect(err).NotTo(HaveOccurred())
	Expect(retrievedValue).To(Equal(value))

	// When: Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Then: Should not be able to retrieve (expired)
	retrievedValue, err = redisCache.Get(ctx, key)
	Expect(err).To(HaveOccurred())
	Expect(retrievedValue).To(BeNil())
}

func TestRedisCache_Ping(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers Redis
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	redisAddr, err := containers.RedisContainer.ConnectionString(context.Background())
	Expect(err).NotTo(HaveOccurred())

	// Remove redis:// prefix if present
	if strings.HasPrefix(redisAddr, "redis://") {
		redisAddr = strings.TrimPrefix(redisAddr, "redis://")
	}
	logger, _ := zap.NewProduction()
	redisCache, err := cache.NewRedisCache(redisAddr, "", 0, logger)
	Expect(err).NotTo(HaveOccurred())

	ctx := context.Background()

	// When: Ping Redis
	err = redisCache.Ping(ctx)

	// Then: Should succeed
	Expect(err).NotTo(HaveOccurred())
}

func TestRedisCache_ConcurrentAccess(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers Redis
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	redisAddr, err := containers.RedisContainer.ConnectionString(context.Background())
	Expect(err).NotTo(HaveOccurred())

	// Remove redis:// prefix if present
	if strings.HasPrefix(redisAddr, "redis://") {
		redisAddr = strings.TrimPrefix(redisAddr, "redis://")
	}
	logger, _ := zap.NewProduction()
	redisCache, err := cache.NewRedisCache(redisAddr, "", 0, logger)
	Expect(err).NotTo(HaveOccurred())

	ctx := context.Background()

	// Given: Multiple goroutines accessing cache concurrently
	numGoroutines := 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			key := fmt.Sprintf("concurrent-key-%d", id)
			value := []byte(fmt.Sprintf("concurrent-value-%d", id))
			ttl := 5 * time.Minute

			// Set value
			err := redisCache.Set(ctx, key, value, ttl)
			Expect(err).NotTo(HaveOccurred())

			// Get value
			retrievedValue, err := redisCache.Get(ctx, key)
			Expect(err).NotTo(HaveOccurred())
			Expect(retrievedValue).To(Equal(value))
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestRedisCache_Delete(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers Redis
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	redisAddr, err := containers.RedisContainer.ConnectionString(context.Background())
	Expect(err).NotTo(HaveOccurred())

	// Remove redis:// prefix if present
	if strings.HasPrefix(redisAddr, "redis://") {
		redisAddr = strings.TrimPrefix(redisAddr, "redis://")
	}

	logger, _ := zap.NewProduction()
	redisCache, err := cache.NewRedisCache(redisAddr, "", 0, logger)
	Expect(err).NotTo(HaveOccurred())

	ctx := context.Background()

	// Given: A key-value pair exists in cache
	key := "delete-test-key"
	value := []byte("delete-test-value")
	ttl := 5 * time.Minute

	err = redisCache.Set(ctx, key, value, ttl)
	Expect(err).NotTo(HaveOccurred())

	// Verify it exists
	retrievedValue, err := redisCache.Get(ctx, key)
	Expect(err).NotTo(HaveOccurred())
	Expect(retrievedValue).To(Equal(value))

	// When: Delete the key
	err = redisCache.Delete(ctx, key)

	// Then: Should delete successfully
	Expect(err).NotTo(HaveOccurred())

	// When: Try to get the deleted key
	retrievedValue, err = redisCache.Get(ctx, key)

	// Then: Should return an error (key not found)
	Expect(err).To(HaveOccurred())
	Expect(retrievedValue).To(BeNil())
	Expect(err.Error()).To(ContainSubstring("not found"))
}

func TestRedisCache_DeleteNonExistent(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers Redis
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	redisAddr, err := containers.RedisContainer.ConnectionString(context.Background())
	Expect(err).NotTo(HaveOccurred())

	// Remove redis:// prefix if present
	if strings.HasPrefix(redisAddr, "redis://") {
		redisAddr = strings.TrimPrefix(redisAddr, "redis://")
	}

	logger, _ := zap.NewProduction()
	redisCache, err := cache.NewRedisCache(redisAddr, "", 0, logger)
	Expect(err).NotTo(HaveOccurred())

	ctx := context.Background()

	// When: Try to delete non-existent key
	err = redisCache.Delete(ctx, "non-existent-key")

	// Then: Should succeed (Redis DEL returns 0 for non-existent keys, but doesn't error)
	Expect(err).NotTo(HaveOccurred())
}

func TestRedisCache_LargeValue(t *testing.T) {
	RegisterTestingT(t)

	// Setup: Testcontainers Redis
	containers, cleanup := setupTestContainers(t)
	defer cleanup()

	redisAddr, err := containers.RedisContainer.ConnectionString(context.Background())
	Expect(err).NotTo(HaveOccurred())

	// Remove redis:// prefix if present
	if strings.HasPrefix(redisAddr, "redis://") {
		redisAddr = strings.TrimPrefix(redisAddr, "redis://")
	}
	logger, _ := zap.NewProduction()
	redisCache, err := cache.NewRedisCache(redisAddr, "", 0, logger)
	Expect(err).NotTo(HaveOccurred())

	ctx := context.Background()

	// Given: A large value (1MB)
	key := "large-value-key"
	largeValue := make([]byte, 1024*1024) // 1MB
	for i := range largeValue {
		largeValue[i] = byte(i % 256)
	}
	ttl := 5 * time.Minute

	// When: Set large value
	err = redisCache.Set(ctx, key, largeValue, ttl)

	// Then: Should set successfully
	Expect(err).NotTo(HaveOccurred())

	// When: Get large value
	retrievedValue, err := redisCache.Get(ctx, key)

	// Then: Should retrieve the same large value
	Expect(err).NotTo(HaveOccurred())
	Expect(retrievedValue).To(Equal(largeValue))
	Expect(len(retrievedValue)).To(Equal(1024 * 1024))
}
