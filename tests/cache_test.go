package tests

import (
	"context"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/rodolfodpk/instagrano/internal/cache"
	"go.uber.org/zap"
)

var _ = Describe("RedisCache", func() {
	Describe("NewRedisCache", func() {
		It("should create Redis cache successfully", func() {
			// Given: Redis container is available
			redisAddr, err := sharedContainers.RedisContainer.ConnectionString(context.Background())
			Expect(err).NotTo(HaveOccurred())

			// Remove redis:// prefix if present
			if strings.HasPrefix(redisAddr, "redis://") {
				redisAddr = strings.TrimPrefix(redisAddr, "redis://")
			}

			// When: Create Redis cache
			logger, _ := zap.NewProduction()
			defer logger.Sync()
			redisCache, err := cache.NewRedisCache(redisAddr, "", 0, logger)

			// Then: Should create successfully
			Expect(err).NotTo(HaveOccurred())
			Expect(redisCache).NotTo(BeNil())
		})

		It("should handle invalid Redis address", func() {
			// Given: Invalid Redis address
			invalidAddr := "invalid:6379"

			// When: Create Redis cache with invalid address
			logger, _ := zap.NewProduction()
			defer logger.Sync()
			redisCache, err := cache.NewRedisCache(invalidAddr, "", 0, logger)

			// Then: Should return error
			Expect(err).To(HaveOccurred())
			Expect(redisCache).To(BeNil())
		})
	})

	Describe("Set and Get", func() {
		It("should set and get string values", func() {
			// Given: Redis cache
			cache := sharedContainers.Cache
			ctx := context.Background()

			// When: Set a value
			err := cache.Set(ctx, "test-key", []byte("test-value"), 5*time.Minute)

			// Then: Should set successfully
			Expect(err).NotTo(HaveOccurred())

			// When: Get the value
			value, err := cache.Get(ctx, "test-key")

			// Then: Should get the correct value
			Expect(err).NotTo(HaveOccurred())
			Expect(string(value)).To(Equal("test-value"))
		})

		It("should return error for non-existent key", func() {
			// Given: Redis cache
			cache := sharedContainers.Cache
			ctx := context.Background()

			// When: Get non-existent key
			value, err := cache.Get(ctx, "non-existent-key")

			// Then: Should return error
			Expect(err).To(HaveOccurred())
			Expect(value).To(BeEmpty())
		})

		It("should handle empty key", func() {
			// Given: Redis cache
			cache := sharedContainers.Cache
			ctx := context.Background()

			// When: Set empty key
			err := cache.Set(ctx, "", []byte("test-value"), 5*time.Minute)

			// Then: Should handle gracefully (Redis allows empty keys)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle empty value", func() {
			// Given: Redis cache
			cache := sharedContainers.Cache
			ctx := context.Background()

			// When: Set empty value
			err := cache.Set(ctx, "test-key", []byte(""), 5*time.Minute)

			// Then: Should set successfully
			Expect(err).NotTo(HaveOccurred())

			// When: Get the value
			value, err := cache.Get(ctx, "test-key")

			// Then: Should get empty value
			Expect(err).NotTo(HaveOccurred())
			Expect(string(value)).To(Equal(""))
		})
	})

	Describe("Delete", func() {
		It("should delete existing key", func() {
			// Given: Redis cache with existing key
			cache := sharedContainers.Cache
			ctx := context.Background()

			err := cache.Set(ctx, "delete-key", []byte("delete-value"), 5*time.Minute)
			Expect(err).NotTo(HaveOccurred())

			// When: Delete the key
			err = cache.Delete(ctx, "delete-key")

			// Then: Should delete successfully
			Expect(err).NotTo(HaveOccurred())

			// When: Try to get deleted key
			value, err := cache.Get(ctx, "delete-key")

			// Then: Should return error (key not found)
			Expect(err).To(HaveOccurred())
			Expect(value).To(BeEmpty())
		})

		It("should handle deletion of non-existent key", func() {
			// Given: Redis cache
			cache := sharedContainers.Cache
			ctx := context.Background()

			// When: Delete non-existent key
			err := cache.Delete(ctx, "non-existent-key")

			// Then: Should handle gracefully (Redis doesn't error on non-existent keys)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("FlushAll", func() {
		It("should flush all keys", func() {
			// Given: Redis cache with multiple keys
			cache := sharedContainers.Cache
			ctx := context.Background()

			// Set multiple keys
			err := cache.Set(ctx, "key1", []byte("value1"), 5*time.Minute)
			Expect(err).NotTo(HaveOccurred())
			err = cache.Set(ctx, "key2", []byte("value2"), 5*time.Minute)
			Expect(err).NotTo(HaveOccurred())

			// When: Flush all
			err = cache.FlushAll(ctx)

			// Then: Should flush successfully
			Expect(err).NotTo(HaveOccurred())

			// When: Try to get flushed keys
			value1, err1 := cache.Get(ctx, "key1")
			value2, err2 := cache.Get(ctx, "key2")

			// Then: Both should return errors (keys not found)
			Expect(err1).To(HaveOccurred())
			Expect(err2).To(HaveOccurred())
			Expect(value1).To(BeEmpty())
			Expect(value2).To(BeEmpty())
		})
	})

	Describe("TTL", func() {
		It("should respect TTL for keys", func() {
			// Given: Redis cache
			cache := sharedContainers.Cache
			ctx := context.Background()

			// When: Set key with short TTL
			err := cache.Set(ctx, "ttl-key", []byte("ttl-value"), 1*time.Second)
			Expect(err).NotTo(HaveOccurred())

			// Then: Should be available immediately
			value, err := cache.Get(ctx, "ttl-key")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(value)).To(Equal("ttl-value"))

			// When: Wait for TTL to expire
			time.Sleep(2 * time.Second)

			// Then: Should be expired
			value, err = cache.Get(ctx, "ttl-key")
			Expect(err).To(HaveOccurred())
			Expect(value).To(BeEmpty())
		})

		It("should handle zero TTL", func() {
			// Given: Redis cache
			cache := sharedContainers.Cache
			ctx := context.Background()

			// When: Set key with zero TTL
			err := cache.Set(ctx, "zero-ttl-key", []byte("zero-ttl-value"), 0)

			// Then: Should set successfully (Redis treats 0 as no expiration)
			Expect(err).NotTo(HaveOccurred())

			// When: Get the value
			value, err := cache.Get(ctx, "zero-ttl-key")

			// Then: Should get the value
			Expect(err).NotTo(HaveOccurred())
			Expect(string(value)).To(Equal("zero-ttl-value"))
		})
	})

	Describe("Publish and Subscribe", func() {
		It("should publish and subscribe to channels", func() {
			// Given: Redis cache
			cache := sharedContainers.Cache
			ctx := context.Background()

			// When: Subscribe to channel
			channel := "test-channel"
			eventCh, err := cache.Subscribe(ctx, channel)
			Expect(err).NotTo(HaveOccurred())

			// Give subscription time to establish
			time.Sleep(200 * time.Millisecond)

			// When: Publish message
			message := "test-message"
			err = cache.Publish(ctx, channel, message)
			Expect(err).NotTo(HaveOccurred())

			// Then: Should receive message
			select {
			case receivedMessage := <-eventCh:
				Expect(receivedMessage).To(Equal(message))
			case <-time.After(500 * time.Millisecond):
				Fail("timeout waiting for message")
			}
		})

		It("should handle multiple subscribers", func() {
			// Given: Redis cache
			cache := sharedContainers.Cache
			ctx := context.Background()

			// When: Create multiple subscribers
			channel := "multi-channel"
			eventCh1, err := cache.Subscribe(ctx, channel)
			Expect(err).NotTo(HaveOccurred())
			eventCh2, err := cache.Subscribe(ctx, channel)
			Expect(err).NotTo(HaveOccurred())

			// Give subscribers time to connect
			time.Sleep(200 * time.Millisecond)

			// When: Publish message
			message := "multi-message"
			err = cache.Publish(ctx, channel, message)
			Expect(err).NotTo(HaveOccurred())

			// Then: Both subscribers should receive message
			select {
			case receivedMessage1 := <-eventCh1:
				Expect(receivedMessage1).To(Equal(message))
			case <-time.After(500 * time.Millisecond):
				Fail("timeout waiting for message on subscriber 1")
			}

			select {
			case receivedMessage2 := <-eventCh2:
				Expect(receivedMessage2).To(Equal(message))
			case <-time.After(500 * time.Millisecond):
				Fail("timeout waiting for message on subscriber 2")
			}
		})

		It("should handle empty channel name", func() {
			// Given: Redis cache
			cache := sharedContainers.Cache
			ctx := context.Background()

			// When: Subscribe to empty channel
			eventCh, err := cache.Subscribe(ctx, "")

			// Then: Should handle gracefully (Redis allows empty channel names)
			Expect(err).NotTo(HaveOccurred())
			Expect(eventCh).NotTo(BeNil())
		})

		It("should handle empty message", func() {
			// Given: Redis cache
			cache := sharedContainers.Cache
			ctx := context.Background()

			// When: Publish empty message
			err := cache.Publish(ctx, "test-channel", "")

			// Then: Should handle gracefully (Redis allows empty messages)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Concurrent Operations", func() {
		It("should handle concurrent set operations", func() {
			// Given: Redis cache
			cache := sharedContainers.Cache
			ctx := context.Background()

			// When: Perform concurrent sets
			done := make(chan bool, 10)
			for i := 0; i < 10; i++ {
				go func(i int) {
					defer func() { done <- true }()
					key := fmt.Sprintf("concurrent-key-%d", i)
					value := fmt.Sprintf("concurrent-value-%d", i)
					err := cache.Set(ctx, key, []byte(value), 5*time.Minute)
					Expect(err).NotTo(HaveOccurred())
				}(i)
			}

			// Then: All operations should complete
			for i := 0; i < 10; i++ {
				select {
				case <-done:
					// Success
				case <-time.After(10 * time.Second):
					Fail("timeout waiting for concurrent operation")
				}
			}
		})

		It("should handle concurrent get operations", func() {
			// Given: Redis cache with existing keys
			cache := sharedContainers.Cache
			ctx := context.Background()

			// Set up test data
			for i := 0; i < 5; i++ {
				key := fmt.Sprintf("concurrent-get-key-%d", i)
				value := fmt.Sprintf("concurrent-get-value-%d", i)
				err := cache.Set(ctx, key, []byte(value), 5*time.Minute)
				Expect(err).NotTo(HaveOccurred())
			}

			// When: Perform concurrent gets
			done := make(chan bool, 5)
			for i := 0; i < 5; i++ {
				go func(i int) {
					defer func() { done <- true }()
					key := fmt.Sprintf("concurrent-get-key-%d", i)
					expectedValue := fmt.Sprintf("concurrent-get-value-%d", i)
					value, err := cache.Get(ctx, key)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(value)).To(Equal(expectedValue))
				}(i)
			}

			// Then: All operations should complete successfully
			for i := 0; i < 5; i++ {
				select {
				case <-done:
					// Success
				case <-time.After(10 * time.Second):
					Fail("timeout waiting for concurrent get operation")
				}
			}
		})
	})
})
