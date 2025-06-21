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

// Legacy error types for backward compatibility (deprecated)

// ValidationError represents validation errors in the domain
// Deprecated: Use shared errors.NewValidationError instead
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return "validation error: " + e.Message
}

// InvalidStatusTransitionError represents invalid status transition errors
// Deprecated: Use NewInvalidStatusTransitionError instead
type InvalidStatusTransitionError struct {
	From Status
	To   Status
}

func (e InvalidStatusTransitionError) Error() string {
	return "invalid status transition from " + string(e.From) + " to " + string(e.To)
}

// InvalidRetryError represents errors when trying to retry in invalid state
// Deprecated: Use NewInvalidRetryError instead
type InvalidRetryError struct {
	CurrentStatus Status
}

func (e InvalidRetryError) Error() string {
	return "cannot retry message in status: " + string(e.CurrentStatus)
}

// MaxRetriesExceededError represents errors when maximum retries are exceeded
// Deprecated: Use NewMaxRetriesExceededError instead
type MaxRetriesExceededError struct {
	CurrentRetries int
}

func (e MaxRetriesExceededError) Error() string {
	return "maximum retries exceeded"
}
