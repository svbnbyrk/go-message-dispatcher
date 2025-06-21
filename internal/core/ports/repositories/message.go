package repositories

import (
	"context"

	"github.com/svbnbyrk/go-message-dispatcher/internal/core/domain/message"
)

// MessageRepository defines the interface for message data access
type MessageRepository interface {
	// Create creates a new message in the repository
	Create(ctx context.Context, msg *message.Message) error

	// GetByID retrieves a message by its ID
	GetByID(ctx context.Context, id message.MessageID) (*message.Message, error)

	// GetPendingMessages retrieves pending messages with limit
	GetPendingMessages(ctx context.Context, limit int) ([]*message.Message, error)

	// Update updates an existing message
	Update(ctx context.Context, msg *message.Message) error

	// GetSentMessages retrieves sent messages with pagination
	GetSentMessages(ctx context.Context, pagination Pagination) ([]*message.Message, error)

	// GetByStatus retrieves messages by status with pagination
	GetByStatus(ctx context.Context, status message.Status, pagination Pagination) ([]*message.Message, error)

	// GetByPhoneNumber retrieves messages by phone number with pagination
	GetByPhoneNumber(ctx context.Context, phoneNumber message.PhoneNumber, pagination Pagination) ([]*message.Message, error)

	// CountByStatus counts messages by status
	CountByStatus(ctx context.Context, status message.Status) (int64, error)

	// DeleteByID deletes a message by ID (for testing purposes)
	DeleteByID(ctx context.Context, id message.MessageID) error
}

// Pagination represents pagination parameters
type Pagination struct {
	Limit  int
	Offset int
}

// DefaultPagination returns default pagination settings
func DefaultPagination() Pagination {
	return Pagination{
		Limit:  50,
		Offset: 0,
	}
}

// NewPagination creates a new pagination with validation
func NewPagination(limit, offset int) Pagination {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return Pagination{
		Limit:  limit,
		Offset: offset,
	}
}
