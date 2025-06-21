package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/domain/message"
)

// Use Case Interfaces

// MessageManagementUseCase handles message creation and retrieval operations
type MessageManagementUseCase interface {
	// CreateMessage creates a new message for sending
	CreateMessage(ctx context.Context, cmd CreateMessageCommand) (*MessageResponse, error)

	// GetMessageByID retrieves a message by its ID
	GetMessageByID(ctx context.Context, id uuid.UUID) (*MessageResponse, error)

	// ListMessages retrieves messages with optional filtering and pagination
	ListMessages(ctx context.Context, query ListMessagesQuery) (*ListMessagesResponse, error)
}

// MessageProcessingUseCase handles message processing operations
type MessageProcessingUseCase interface {
	// ProcessPendingMessages processes a batch of pending messages
	ProcessPendingMessages(ctx context.Context, batchSize int) (*ProcessingResult, error)

	// GetProcessingStatus returns the current processing status
	GetProcessingStatus(ctx context.Context) (*ProcessingStatus, error)
}

// DTOs for use cases

// CreateMessageCommand represents the input for creating a message
type CreateMessageCommand struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
	Content     string `json:"content" validate:"required"`
}

// MessageResponse represents the output for message operations
type MessageResponse struct {
	ID          uuid.UUID  `json:"id"`
	PhoneNumber string     `json:"phone_number"`
	Content     string     `json:"content"`
	Status      string     `json:"status"`
	ExternalID  *string    `json:"external_id,omitempty"`
	RetryCount  int        `json:"retry_count"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	SentAt      *time.Time `json:"sent_at,omitempty"`
}

// ListMessagesQuery represents the input for listing messages
type ListMessagesQuery struct {
	Status *message.Status `json:"status,omitempty"`
	Limit  int             `json:"limit,omitempty"`
	Offset int             `json:"offset,omitempty"`
}

// ListMessagesResponse represents the output for listing messages
type ListMessagesResponse struct {
	Messages   []MessageResponse `json:"messages"`
	TotalCount int64             `json:"total_count"`
	HasMore    bool              `json:"has_more"`
}

// ProcessingResult represents the result of message processing
type ProcessingResult struct {
	ProcessedCount int     `json:"processed_count"`
	SuccessCount   int     `json:"success_count"`
	FailedCount    int     `json:"failed_count"`
	Errors         []error `json:"errors,omitempty"`
}

// ProcessingStatus represents the current state of message processing
type ProcessingStatus struct {
	IsProcessing     bool       `json:"is_processing"`
	LastProcessedAt  *time.Time `json:"last_processed_at,omitempty"`
	PendingCount     int64      `json:"pending_count"`
	ProcessedToday   int64      `json:"processed_today"`
	FailedToday      int64      `json:"failed_today"`
	NextProcessingAt *time.Time `json:"next_processing_at,omitempty"`
}
