package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"time"

	domainErrors "github.com/svbnbyrk/go-message-dispatcher/internal/core/domain/errors"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/ports/services"
	"github.com/svbnbyrk/go-message-dispatcher/internal/shared/logger"
	"go.uber.org/zap"
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

// SendMessage sends a message to the configured webhook endpoint with retry mechanism
func (s *httpWebhookService) SendMessage(ctx context.Context, request services.WebhookRequest) (*services.WebhookResponse, error) {
	if s.config.URL == "" {
		return nil, domainErrors.NewValidationError("webhook URL is not configured")
	}

	var lastErr error
	maxRetries := s.config.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3 // Default fallback
	}

	// Try initial request + retries
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Add delay for retry attempts (not for first attempt)
		if attempt > 0 {
			delay := s.calculateBackoffDelay(attempt)
			logger.WarnCtx(ctx, "Webhook request failed, retrying",
				zap.Int("attempt", attempt),
				zap.Int("max_retries", maxRetries),
				zap.Duration("delay", delay),
				zap.Error(lastErr),
			)

			select {
			case <-ctx.Done():
				return nil, domainErrors.NewBusinessError("webhook request cancelled during retry")
			case <-time.After(delay):
				// Continue with retry
			}
		}

		response, err := s.sendSingleRequest(ctx, request)
		if err == nil {
			// Success!
			if attempt > 0 {
				logger.InfoCtx(ctx, "Webhook request succeeded after retries",
					zap.Int("successful_attempt", attempt+1),
					zap.Int("total_attempts", attempt+1),
				)
			}
			return response, nil
		}

		lastErr = err

		// Check if we should retry this error
		if !s.shouldRetryError(err) {
			logger.WarnCtx(ctx, "Webhook request failed with non-retryable error",
				zap.Error(err),
				zap.Int("attempt", attempt+1),
			)
			break
		}

		// Don't retry if this is the last attempt
		if attempt == maxRetries {
			logger.ErrorCtx(ctx, "Webhook request failed after all retries",
				err,
				zap.Int("total_attempts", attempt+1),
				zap.Int("max_retries", maxRetries),
			)
			break
		}
	}

	return nil, domainErrors.NewBusinessError("webhook request failed after %d attempts: %v", maxRetries+1, lastErr)
}

// sendSingleRequest makes a single HTTP request to the webhook endpoint
func (s *httpWebhookService) sendSingleRequest(ctx context.Context, request services.WebhookRequest) (*services.WebhookResponse, error) {
	start := time.Now()

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
	httpReq.Header.Set("User-Agent", "Message-Dispatcher/1.0")
	if s.config.AuthToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+s.config.AuthToken)
	}

	// Send request
	httpResp, err := s.client.Do(httpReq)
	if err != nil {
		duration := time.Since(start)
		logger.WarnCtx(ctx, "Webhook HTTP request failed",
			zap.String("url", s.config.URL),
			zap.Error(err),
			zap.Int64("duration_ms", duration.Milliseconds()),
		)
		return nil, domainErrors.NewBusinessError("HTTP request failed: %v", err)
	}
	defer httpResp.Body.Close()

	duration := time.Since(start)

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		logger.WarnCtx(ctx, "Failed to read webhook response body",
			zap.String("url", s.config.URL),
			zap.Int("status_code", httpResp.StatusCode),
			zap.Error(err),
		)
		return nil, domainErrors.NewBusinessError("failed to read response body: %v", err)
	}

	// Log request completion
	logger.InfoCtx(ctx, "Webhook request completed",
		zap.String("url", s.config.URL),
		zap.Int("status_code", httpResp.StatusCode),
		zap.Int64("duration_ms", duration.Milliseconds()),
		zap.Int("response_size", len(respBody)),
	)

	// Check HTTP status
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, domainErrors.NewBusinessError("webhook returned HTTP %d: %s", httpResp.StatusCode, string(respBody))
	}

	// Parse response
	var webhookResp services.WebhookResponse
	if err := json.Unmarshal(respBody, &webhookResp); err != nil {
		return nil, domainErrors.NewBusinessError("failed to parse webhook response - Raw body: '%s', Error: %v", string(respBody), err)
	}

	return &webhookResp, nil
}

// shouldRetryError determines if an error should be retried based on HTTP status code
func (s *httpWebhookService) shouldRetryError(err error) bool {
	// Don't retry validation errors
	var validationErr domainErrors.ValidationError
	if errors.As(err, &validationErr) {
		return false
	}

	return true
}

// calculateBackoffDelay calculates the delay for retry attempts using exponential backoff
func (s *httpWebhookService) calculateBackoffDelay(attempt int) time.Duration {
	// Calculate exponential backoff: base * 2^(attempt-1)
	baseDelay := s.config.RetryBackoffBase
	if baseDelay <= 0 {
		baseDelay = time.Second // Default fallback
	}

	// Calculate exponential delay
	exponentialDelay := float64(baseDelay) * math.Pow(2, float64(attempt-1))

	// Apply maximum delay limit
	maxDelay := s.config.RetryBackoffMax
	if maxDelay <= 0 {
		maxDelay = 5 * time.Second // Default fallback
	}

	if exponentialDelay > float64(maxDelay) {
		exponentialDelay = float64(maxDelay)
	}

	return time.Duration(exponentialDelay)
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
