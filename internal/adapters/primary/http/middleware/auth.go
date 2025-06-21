package middleware

import (
	"net/http"
	"strings"

	"github.com/svbnbyrk/go-message-dispatcher/internal/adapters/primary/http/dto"
)

// AuthConfig contains authentication configuration
type AuthConfig struct {
	APIKey string
}

// ClientCredentialsAuth creates a simple API key authentication middleware
func ClientCredentialsAuth(config AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for health check
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeJSONError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			// Expected format: "Bearer <api-key>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeJSONError(w, http.StatusUnauthorized, "invalid authorization format")
				return
			}

			apiKey := parts[1]
			if apiKey != config.APIKey {
				writeJSONError(w, http.StatusUnauthorized, "invalid API key")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResp := dto.ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}

	// Simple JSON writing without external dependencies
	w.Write([]byte(`{"error":"` + errorResp.Error + `","message":"` + errorResp.Message + `"}`))
}
