package message

import (
	"time"

	"github.com/google/uuid"
)

// Status represents the current state of a message
type Status string

const (
	StatusPending Status = "PENDING"
	StatusSent    Status = "SENT"
	StatusFailed  Status = "FAILED"
)

// String returns the string representation of Status
func (s Status) String() string {
	return string(s)
}

// IsValid checks if the status is valid
func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusSent, StatusFailed:
		return true
	default:
		return false
	}
}

// MessageID represents a unique identifier for a message
type MessageID string

// NewMessageID generates a new unique message ID
func NewMessageID() MessageID {
	return MessageID(uuid.New().String())
}

// String returns the string representation of MessageID
func (id MessageID) String() string {
	return string(id)
}

// IsEmpty checks if the MessageID is empty
func (id MessageID) IsEmpty() bool {
	return string(id) == ""
}

// Message represents the core business entity for messages
type Message struct {
	ID          MessageID
	PhoneNumber PhoneNumber
	Content     Content
	Status      Status
	ExternalID  *string
	RetryCount  int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	SentAt      *time.Time
}

// NewMessage creates a new message with the provided phone number and content
func NewMessage(phoneNumber PhoneNumber, content Content) (*Message, error) {
	if err := phoneNumber.Validate(); err != nil {
		return nil, err
	}

	if err := content.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Message{
		ID:          NewMessageID(),
		PhoneNumber: phoneNumber,
		Content:     content,
		Status:      StatusPending,
		RetryCount:  0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Business methods

// MarkAsSent marks the message as successfully sent
func (m *Message) MarkAsSent(externalID string) error {
	if m.Status != StatusPending {
		return NewInvalidStatusTransitionError(m.Status, StatusSent)
	}

	now := time.Now()
	m.Status = StatusSent
	m.ExternalID = &externalID
	m.SentAt = &now
	m.UpdatedAt = now

	return nil
}

// MarkAsFailed marks the message as failed
func (m *Message) MarkAsFailed() error {
	if m.Status == StatusSent {
		return NewInvalidStatusTransitionError(m.Status, StatusFailed)
	}

	m.Status = StatusFailed
	m.UpdatedAt = time.Now()

	return nil
}

// IncrementRetry increments the retry count
func (m *Message) IncrementRetry() error {
	if m.Status != StatusPending && m.Status != StatusFailed {
		return NewInvalidRetryError(m.Status)
	}

	if m.RetryCount >= MaxRetryAttempts {
		return NewMaxRetriesExceededError(m.RetryCount)
	}

	m.RetryCount++
	m.UpdatedAt = time.Now()

	return nil
}

// CanRetry checks if the message can be retried
func (m *Message) CanRetry() bool {
	return (m.Status == StatusPending || m.Status == StatusFailed) && m.RetryCount < MaxRetryAttempts
}

// Constants for business rules
const (
	MaxRetryAttempts = 3
)
