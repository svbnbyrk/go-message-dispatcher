package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	domainErrors "github.com/svbnbyrk/go-message-dispatcher/internal/core/domain/errors"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/ports/services"
)

// redisCacheService implements the CacheService interface
type redisCacheService struct {
	client *redis.Client
	config services.CacheConfig
}

// NewRedisService creates a new Redis cache service
func NewRedisService(config services.CacheConfig) services.CacheService {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	return &redisCacheService{
		client: rdb,
		config: config,
	}
}

// Set stores a value in cache with optional TTL
func (r *redisCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	var expiration time.Duration
	if ttl > 0 {
		expiration = ttl
	} else {
		expiration = r.config.TTL
	}

	err := r.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return domainErrors.NewBusinessError("failed to set cache key %s: %v", key, err)
	}

	return nil
}

// Get retrieves a value from cache
func (r *redisCacheService) Get(ctx context.Context, key string) (interface{}, error) {
	result := r.client.Get(ctx, key)
	if err := result.Err(); err != nil {
		if err == redis.Nil {
			return nil, domainErrors.NewNotFoundError("cache key %s not found", key)
		}
		return nil, domainErrors.NewBusinessError("failed to get cache key %s: %v", key, err)
	}

	return result.Val(), nil
}

// Delete removes a value from cache
func (r *redisCacheService) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return domainErrors.NewBusinessError("failed to delete cache key %s: %v", key, err)
	}

	return nil
}

// Exists checks if a key exists in cache
func (r *redisCacheService) Exists(ctx context.Context, key string) (bool, error) {
	result := r.client.Exists(ctx, key)
	if err := result.Err(); err != nil {
		return false, domainErrors.NewBusinessError("failed to check existence of cache key %s: %v", key, err)
	}

	return result.Val() > 0, nil
}

// SetJSON stores a JSON-serializable value in cache
func (r *redisCacheService) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return domainErrors.NewValidationError("failed to marshal value to JSON: %v", err)
	}

	return r.Set(ctx, key, jsonData, ttl)
}

// GetJSON retrieves and deserializes a JSON value from cache
func (r *redisCacheService) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := r.Get(ctx, key)
	if err != nil {
		return err
	}

	jsonStr, ok := data.(string)
	if !ok {
		return domainErrors.NewBusinessError("cached value is not a JSON string")
	}

	if err := json.Unmarshal([]byte(jsonStr), dest); err != nil {
		return domainErrors.NewBusinessError("failed to unmarshal JSON from cache: %v", err)
	}

	return nil
}

// IsHealthy checks if the cache service is healthy
func (r *redisCacheService) IsHealthy(ctx context.Context) error {
	result := r.client.Ping(ctx)
	if err := result.Err(); err != nil {
		return domainErrors.NewBusinessError("Redis health check failed: %v", err)
	}

	return nil
}
