package services

import (
	"context"
	"time"
)

// WebhookRequest represents the data sent to webhook endpoint
type WebhookRequest struct {
	To      string `json:"to"`
	Content string `json:"content"`
}

// WebhookResponse represents the response from webhook endpoint
type WebhookResponse struct {
	MessageID string `json:"messageId"`
	Message   string `json:"message"`
}

// WebhookService handles sending messages to external webhook endpoints
type WebhookService interface {
	// SendMessage sends a message to the configured webhook endpoint
	SendMessage(ctx context.Context, request WebhookRequest) (*WebhookResponse, error)

	// IsHealthy checks if the webhook service is healthy and reachable
	IsHealthy(ctx context.Context) error
}

// WebhookConfig contains configuration for webhook service
type WebhookConfig struct {
	URL              string        `yaml:"url" env:"WEBHOOK_URL"`
	AuthToken        string        `yaml:"auth_token" env:"WEBHOOK_AUTH_TOKEN"`
	Timeout          time.Duration `yaml:"timeout" env:"WEBHOOK_TIMEOUT"`
	MaxRetries       int           `yaml:"max_retries" env:"WEBHOOK_MAX_RETRIES"`
	RetryBackoffBase time.Duration `yaml:"retry_backoff_base" env:"WEBHOOK_RETRY_BACKOFF_BASE"`
	RetryBackoffMax  time.Duration `yaml:"retry_backoff_max" env:"WEBHOOK_RETRY_BACKOFF_MAX"`
}
