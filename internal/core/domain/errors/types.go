package errors

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Domain Error Interfaces - Business-specific error types with marker methods
type ValidationError interface {
	error
	IsValidationError() // Marker method
}

type NotFoundError interface {
	error
	IsNotFoundError() // Marker method
}

type BusinessError interface {
	error
	IsBusinessError() // Marker method
}

type RepositoryError interface {
	error
	IsRepositoryError() // Marker method
}

// Concrete implementations
type validationError struct {
	message string
	cause   error
}

func (e validationError) Error() string      { return e.message }
func (e validationError) Unwrap() error      { return e.cause }
func (e validationError) IsValidationError() {} // Marker method implementation

type notFoundError struct {
	message string
	cause   error
}

func (e notFoundError) Error() string    { return e.message }
func (e notFoundError) Unwrap() error    { return e.cause }
func (e notFoundError) IsNotFoundError() {} // Marker method implementation

type businessError struct {
	message string
	cause   error
}

func (e businessError) Error() string    { return e.message }
func (e businessError) Unwrap() error    { return e.cause }
func (e businessError) IsBusinessError() {} // Marker method implementation

type repositoryError struct {
	message string
	cause   error
}

func (e repositoryError) Error() string      { return e.message }
func (e repositoryError) Unwrap() error      { return e.cause }
func (e repositoryError) IsRepositoryError() {} // Marker method implementation

// Factory Functions - sprintf style
func NewValidationError(format string, args ...interface{}) ValidationError {
	return validationError{
		message: fmt.Sprintf("validation error: "+format, args...),
	}
}

func NewValidationErrorWithCause(cause error, format string, args ...interface{}) ValidationError {
	return validationError{
		message: fmt.Sprintf("validation error: "+format, args...),
		cause:   cause,
	}
}

func NewNotFoundError(format string, args ...interface{}) NotFoundError {
	return notFoundError{
		message: fmt.Sprintf("not found: "+format, args...),
	}
}

func NewNotFoundErrorWithCause(cause error, format string, args ...interface{}) NotFoundError {
	return notFoundError{
		message: fmt.Sprintf("not found: "+format, args...),
		cause:   cause,
	}
}

func NewBusinessError(format string, args ...interface{}) BusinessError {
	return businessError{
		message: fmt.Sprintf("business error: "+format, args...),
	}
}

func NewBusinessErrorWithCause(cause error, format string, args ...interface{}) BusinessError {
	return businessError{
		message: fmt.Sprintf("business error: "+format, args...),
		cause:   cause,
	}
}

func NewRepositoryError(format string, args ...interface{}) RepositoryError {
	return repositoryError{
		message: fmt.Sprintf("repository error: "+format, args...),
	}
}

func NewRepositoryErrorWithCause(cause error, format string, args ...interface{}) RepositoryError {
	return repositoryError{
		message: fmt.Sprintf("repository error: "+format, args...),
		cause:   cause,
	}
}

// MapNotFoundError converts pgx.ErrNoRows to business not found error
func MapNotFoundError(err error, context string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return NewNotFoundError("%s not found", context)
	}
	return err // Return other errors as-is
}
