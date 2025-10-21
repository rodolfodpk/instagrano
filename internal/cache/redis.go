package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Cache interface for caching operations
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Ping(ctx context.Context) error
	FlushAll(ctx context.Context) error
	Publish(ctx context.Context, channel string, message string) error
	Subscribe(ctx context.Context, channel string) (<-chan string, error)
	Close() error
}

// RedisCache implements the Cache interface using Redis
type RedisCache struct {
	client *redis.Client
	logger *zap.Logger
}

// NewRedisCache creates a new Redis cache client
func NewRedisCache(addr, password string, db int, logger *zap.Logger) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	logger.Info("redis client created successfully",
		zap.String("addr", addr),
		zap.Int("db", db),
	)

	return &RedisCache{
		client: client,
		logger: logger,
	}, nil
}

// Get retrieves a value from cache
func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("cache miss: key not found")
	}
	if err != nil {
		r.logger.Error("redis get failed",
			zap.String("key", key),
			zap.Error(err),
		)
		return nil, err
	}
	return val, nil
}

// Set stores a value in cache with TTL
func (r *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	err := r.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		r.logger.Error("redis set failed",
			zap.String("key", key),
			zap.Duration("ttl", ttl),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// Delete removes a value from cache
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		r.logger.Error("redis delete failed",
			zap.String("key", key),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// Ping checks if Redis is reachable
func (r *RedisCache) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// FlushAll removes all keys from the current database
func (r *RedisCache) FlushAll(ctx context.Context) error {
	err := r.client.FlushDB(ctx).Err()
	if err != nil {
		r.logger.Error("redis flushdb failed", zap.Error(err))
		return err
	}
	return nil
}

// Publish sends a message to a Redis channel
func (r *RedisCache) Publish(ctx context.Context, channel string, message string) error {
	err := r.client.Publish(ctx, channel, message).Err()
	if err != nil {
		r.logger.Error("redis publish failed",
			zap.String("channel", channel),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// Subscribe subscribes to a Redis channel and returns a channel for receiving messages
func (r *RedisCache) Subscribe(ctx context.Context, channel string) (<-chan string, error) {
	pubsub := r.client.Subscribe(ctx, channel)

	// Get the channel for receiving messages
	ch := pubsub.Channel()

	// Convert redis.PubSubMessage to string channel
	stringCh := make(chan string)
	go func() {
		defer close(stringCh)
		defer pubsub.Close()
		for msg := range ch {
			select {
			case stringCh <- msg.Payload:
			case <-ctx.Done():
				return
			}
		}
	}()

	return stringCh, nil
}

// Close closes the Redis client connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}
