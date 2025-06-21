package message

import "fmt"

// ValidationError represents validation errors in the domain
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s", e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(message string) error {
	return ValidationError{Message: message}
}

// InvalidStatusTransitionError represents invalid status transition errors
type InvalidStatusTransitionError struct {
	From Status
	To   Status
}

func (e InvalidStatusTransitionError) Error() string {
	return fmt.Sprintf("invalid status transition from %s to %s", e.From, e.To)
}

// NewInvalidStatusTransitionError creates a new invalid status transition error
func NewInvalidStatusTransitionError(from, to Status) error {
	return InvalidStatusTransitionError{From: from, To: to}
}

// InvalidRetryError represents errors when trying to retry in invalid state
type InvalidRetryError struct {
	CurrentStatus Status
}

func (e InvalidRetryError) Error() string {
	return fmt.Sprintf("cannot retry message in status: %s", e.CurrentStatus)
}

// NewInvalidRetryError creates a new invalid retry error
func NewInvalidRetryError(status Status) error {
	return InvalidRetryError{CurrentStatus: status}
}

// MaxRetriesExceededError represents errors when maximum retries are exceeded
type MaxRetriesExceededError struct {
	CurrentRetries int
}

func (e MaxRetriesExceededError) Error() string {
	return fmt.Sprintf("maximum retries exceeded: %d", e.CurrentRetries)
}

// NewMaxRetriesExceededError creates a new max retries exceeded error
func NewMaxRetriesExceededError(currentRetries int) error {
	return MaxRetriesExceededError{CurrentRetries: currentRetries}
}
