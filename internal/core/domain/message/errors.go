package message

import (
	"github.com/svbnbyrk/go-message-dispatcher/internal/core/domain/errors"
)

// Message-specific error helpers using domain error infrastructure

// NewValidationError creates a validation error for message domain
func NewValidationError(format string, args ...interface{}) error {
	return errors.NewValidationError(format, args...)
}

// NewInvalidStatusTransitionError creates an invalid status transition error
func NewInvalidStatusTransitionError(from, to Status) error {
	return errors.NewBusinessError("invalid status transition from %s to %s", from, to)
}

// NewInvalidRetryError creates an invalid retry error
func NewInvalidRetryError(status Status) error {
	return errors.NewBusinessError("cannot retry message in status: %s", status)
}

// NewMaxRetriesExceededError creates a max retries exceeded error
func NewMaxRetriesExceededError(currentRetries int) error {
	return errors.NewBusinessError("maximum retry attempts exceeded: %d", currentRetries)
}

// Business-specific error helpers for message validation
func NewPhoneNumberValidationError(phoneNumber string) error {
	return errors.NewValidationError("invalid phone number format: %s", phoneNumber)
}

func NewContentValidationError(reason string) error {
	return errors.NewValidationError("invalid message content: %s", reason)
}

func NewContentTooLongError(length, maxLength int) error {
	return errors.NewValidationError("message content too long: %d characters (max: %d)", length, maxLength)
}

func NewEmptyContentError() error {
	return errors.NewValidationError("message content cannot be empty")
}
