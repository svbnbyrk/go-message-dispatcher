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
}

// CacheConfig contains configuration for cache service
type CacheConfig struct {
	Host     string        `yaml:"host" env:"REDIS_HOST"`
	Port     int           `yaml:"port" env:"REDIS_PORT"`
	Password string        `yaml:"password" env:"REDIS_PASSWORD"`
	DB       int           `yaml:"db" env:"REDIS_DB"`
	TTL      time.Duration `yaml:"ttl" env:"REDIS_TTL"`
}
