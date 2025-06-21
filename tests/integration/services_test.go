package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/svbnbyrk/go-message-dispatcher/internal/adapters/secondary/services/webhook"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/ports/services"
)

func TestWebhookService_Integration(t *testing.T) {
	t.Run("successful webhook call", func(t *testing.T) {
		// Create test HTTP server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"messageId": "test-123", "message": "Message sent successfully"}`))
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
			To:      "+905551234567",
			Content: "Test message",
		}

		ctx := context.Background()
		response, err := webhookService.SendMessage(ctx, request)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if response == nil {
			t.Error("Expected response but got nil")
		}

		if response.MessageID != "test-123" {
			t.Errorf("Expected message ID 'test-123', got '%s'", response.MessageID)
		}

		if response.Message != "Message sent successfully" {
			t.Errorf("Expected message 'Message sent successfully', got '%s'", response.Message)
		}
	})
}
