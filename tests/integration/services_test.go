package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/svbnbyrk/go-message-dispatcher/internal/adapters/secondary/services/cache"
	"github.com/svbnbyrk/go-message-dispatcher/internal/adapters/secondary/services/webhook"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/ports/services"
)

func TestWebhookService_Integration(t *testing.T) {
	t.Run("successful webhook call", func(t *testing.T) {
		// Create test HTTP server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true, "external_id": "test-123"}`))
		}))
		defer server.Close()

		// Create webhook service
		config := services.WebhookConfig{
			URL:              server.URL,
			AuthToken:        "test-token",
			Timeout:          5 * time.Second,
			MaxRetries:       3,
			RetryBackoffBase: 100 * time.Millisecond,
		}

		webhookService := webhook.NewWebhookService(config)

		// Test webhook call
		request := services.WebhookRequest{
			PhoneNumber: "+905551234567",
			Content:     "Test message",
			MessageID:   "msg-123",
		}

		ctx := context.Background()
		response, err := webhookService.SendMessage(ctx, request)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if response == nil {
			t.Error("Expected response but got nil")
		}

		if !response.Success {
			t.Error("Expected successful response")
		}

		if response.ExternalID != "test-123" {
			t.Errorf("Expected external ID 'test-123', got '%s'", response.ExternalID)
		}
	})
}

func TestCacheService_Integration(t *testing.T) {
	t.Run("memory cache operations", func(t *testing.T) {
		cacheService := cache.NewMemoryService()
		ctx := context.Background()

		// Test Set/Get
		key := "test-key"
		value := "test-value"
		err := cacheService.Set(ctx, key, value, 5*time.Second)
		if err != nil {
			t.Errorf("Unexpected error setting cache: %v", err)
		}

		retrievedValue, err := cacheService.Get(ctx, key)
		if err != nil {
			t.Errorf("Unexpected error getting cache: %v", err)
		}

		if retrievedValue != value {
			t.Errorf("Expected value '%s', got '%s'", value, retrievedValue)
		}

		// Test health check
		err = cacheService.IsHealthy(ctx)
		if err != nil {
			t.Errorf("Expected memory cache to be healthy, got error: %v", err)
		}
	})
}
