package services

import (
	"context"
	"time"
)

// CacheService handles generic caching operations
type CacheService interface {
	// Set stores a value in cache with optional TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Get retrieves a value from cache
	Get(ctx context.Context, key string) (interface{}, error)

	// Delete removes a value from cache
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in cache
	Exists(ctx context.Context, key string) (bool, error)

	// SetJSON stores a JSON-serializable value in cache
	SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// GetJSON retrieves and deserializes a JSON value from cache
	GetJSON(ctx context.Context, key string, dest interface{}) error

	// IsHealthy checks if the cache service is healthy
	IsHealthy(ctx context.Context) error

	// Sorted Set Operations for efficient message listing
	// ZAdd adds a member with score to sorted set
	ZAdd(ctx context.Context, key string, score float64, member string, ttl time.Duration) error

	// ZRevRange gets members from sorted set in descending order (latest first)
	ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error)

	// ZRem removes members from sorted set
	ZRem(ctx context.Context, key string, members ...string) error

	// ZCard gets the number of elements in sorted set
	ZCard(ctx context.Context, key string) (int64, error)

	// ZRemRangeByRank removes elements by rank (keep only top N)
	ZRemRangeByRank(ctx context.Context, key string, start, stop int64) error
}

// CacheConfig contains configuration for cache service
type CacheConfig struct {
	Host     string        `yaml:"host" env:"REDIS_HOST"`
	Port     int           `yaml:"port" env:"REDIS_PORT"`
	Password string        `yaml:"password" env:"REDIS_PASSWORD"`
	DB       int           `yaml:"db" env:"REDIS_DB"`
	TTL      time.Duration `yaml:"ttl" env:"REDIS_TTL"`
}

// SentMessageCacheData represents cached sent message data
type SentMessageCacheData struct {
	MessageID   string    `json:"message_id"`
	ExternalID  string    `json:"external_id"`
	PhoneNumber string    `json:"phone_number"`
	Content     string    `json:"content"`
	SentAt      time.Time `json:"sent_at"`
}
