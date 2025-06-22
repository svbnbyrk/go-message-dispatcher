package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/httprate"
)

// RateLimitConfig contains configuration for rate limiting
type RateLimitConfig struct {
	Enabled           bool
	RequestsPerMinute int
	BurstSize         int
	CleanupInterval   time.Duration
}

// RateLimitMiddleware creates a rate limiting middleware based on IP address
func RateLimitMiddleware(config RateLimitConfig) func(http.Handler) http.Handler {
	if !config.Enabled {
		// Return no-op middleware if rate limiting is disabled
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	// Set defaults if not provided
	requestsPerMinute := config.RequestsPerMinute
	if requestsPerMinute <= 0 {
		requestsPerMinute = 60 // Default: 60 requests per minute
	}

	cleanupInterval := config.CleanupInterval
	if cleanupInterval <= 0 {
		cleanupInterval = 5 * time.Minute // Default: cleanup every 5 minutes
	}

	return httprate.LimitByIP(requestsPerMinute, time.Minute)
}
