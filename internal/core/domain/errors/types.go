package errors

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Domain Error Interfaces - Business-specific error types
type ValidationError interface {
	error
}

type NotFoundError interface {
	error
}

type BusinessError interface {
	error
}

type RepositoryError interface {
	error
}

// Concrete implementations
type validationError struct {
	message string
	cause   error
}

func (e validationError) Error() string { return e.message }
func (e validationError) Unwrap() error { return e.cause }

type notFoundError struct {
	message string
	cause   error
}

func (e notFoundError) Error() string { return e.message }
func (e notFoundError) Unwrap() error { return e.cause }

type businessError struct {
	message string
	cause   error
}

func (e businessError) Error() string { return e.message }
func (e businessError) Unwrap() error { return e.cause }

type repositoryError struct {
	message string
	cause   error
}

func (e repositoryError) Error() string { return e.message }
func (e repositoryError) Unwrap() error { return e.cause }

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
