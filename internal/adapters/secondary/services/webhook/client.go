package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	domainErrors "github.com/svbnbyrk/go-message-dispatcher/internal/core/domain/errors"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/ports/services"
)

// httpWebhookService implements the WebhookService interface
type httpWebhookService struct {
	client *http.Client
	config services.WebhookConfig
}

// NewWebhookService creates a new webhook service with HTTP client
func NewWebhookService(config services.WebhookConfig) services.WebhookService {
	return &httpWebhookService{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		config: config,
	}
}

// SendMessage sends a message to the configured webhook endpoint
func (s *httpWebhookService) SendMessage(ctx context.Context, request services.WebhookRequest) (*services.WebhookResponse, error) {
	if s.config.URL == "" {
		return nil, domainErrors.NewValidationError("webhook URL is not configured")
	}

	var lastErr error

	// Retry logic with exponential backoff
	for attempt := 0; attempt <= s.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry with exponential backoff
			backoffDuration := s.config.RetryBackoffBase * time.Duration(1<<(attempt-1))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoffDuration):
			}
		}

		response, err := s.sendSingleRequest(ctx, request)
		if err == nil {
			return response, nil
		}

		lastErr = err

		// Don't retry on validation errors or context cancellation
		if isNonRetryableError(err) || ctx.Err() != nil {
			break
		}
	}

	return nil, domainErrors.NewBusinessErrorWithCause(lastErr, "webhook request failed after %d attempts", s.config.MaxRetries+1)
}

// sendSingleRequest makes a single HTTP request to the webhook endpoint
func (s *httpWebhookService) sendSingleRequest(ctx context.Context, request services.WebhookRequest) (*services.WebhookResponse, error) {
	// Prepare request body
	jsonBody, err := json.Marshal(request)
	if err != nil {
		return nil, domainErrors.NewValidationError("failed to marshal webhook request: %v", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.config.URL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, domainErrors.NewValidationError("failed to create HTTP request: %v", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	if s.config.AuthToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+s.config.AuthToken)
	}

	// Send request
	httpResp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, domainErrors.NewBusinessError("HTTP request failed: %v", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, domainErrors.NewBusinessError("failed to read response body: %v", err)
	}

	// Check HTTP status
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, domainErrors.NewBusinessError("webhook returned HTTP %d: %s", httpResp.StatusCode, string(respBody))
	}

	// Parse response
	var webhookResp services.WebhookResponse
	if err := json.Unmarshal(respBody, &webhookResp); err != nil {
		return nil, domainErrors.NewBusinessError("failed to parse webhook response: %v", err)
	}

	if !webhookResp.Success {
		return nil, domainErrors.NewBusinessError("webhook indicated failure: %s", webhookResp.ErrorMessage)
	}

	return &webhookResp, nil
}

// IsHealthy checks if the webhook service is healthy and reachable
func (s *httpWebhookService) IsHealthy(ctx context.Context) error {
	if s.config.URL == "" {
		return domainErrors.NewValidationError("webhook URL is not configured")
	}

	// Create a lightweight HEAD request to check availability
	req, err := http.NewRequestWithContext(ctx, "HEAD", s.config.URL, nil)
	if err != nil {
		return domainErrors.NewBusinessError("failed to create health check request: %v", err)
	}

	if s.config.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.config.AuthToken)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return domainErrors.NewBusinessError("webhook health check failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return domainErrors.NewBusinessError("webhook is unhealthy: HTTP %d", resp.StatusCode)
	}

	return nil
}

// isNonRetryableError determines if an error should not be retried
func isNonRetryableError(err error) bool {
	var validationErr domainErrors.ValidationError
	return errors.As(err, &validationErr)
}
