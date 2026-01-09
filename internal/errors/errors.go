// Package errors provides custom error types for TicketD.
// These error types allow for better error classification and handling
// throughout the application, enabling appropriate HTTP status codes
// and user-friendly error messages.
package errors

import (
	"errors"
	"fmt"
)

// Sentinel errors for common error conditions
var (
	// ErrNotFound indicates that the requested resource was not found.
	// This typically maps to HTTP 404 status code.
	ErrNotFound = errors.New("resource not found")

	// ErrInvalidInput indicates that the provided input is invalid.
	// This typically maps to HTTP 400 status code.
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized indicates that authentication is required but was not provided.
	// This typically maps to HTTP 401 status code.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates that the user is authenticated but lacks permission.
	// This typically maps to HTTP 403 status code.
	ErrForbidden = errors.New("forbidden")

	// ErrInternal indicates an unexpected internal server error.
	// This typically maps to HTTP 500 status code.
	ErrInternal = errors.New("internal server error")
)

// NotFoundError creates a new not found error with a descriptive message.
func NotFoundError(resource string, id interface{}) error {
	return fmt.Errorf("%s with id %v: %w", resource, id, ErrNotFound)
}

// InvalidInputError creates a new invalid input error with a descriptive message.
func InvalidInputError(field, reason string) error {
	return fmt.Errorf("invalid %s: %s: %w", field, reason, ErrInvalidInput)
}

// IsNotFound checks if an error is or wraps ErrNotFound.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsInvalidInput checks if an error is or wraps ErrInvalidInput.
func IsInvalidInput(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}

// IsUnauthorized checks if an error is or wraps ErrUnauthorized.
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsForbidden checks if an error is or wraps ErrForbidden.
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsInternal checks if an error is or wraps ErrInternal.
func IsInternal(err error) bool {
	return errors.Is(err, ErrInternal)
}

// Wrap wraps an error with additional context.
// It uses fmt.Errorf with %w to preserve the error chain for errors.Is/As.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Wrapf wraps an error with a formatted message.
// It uses fmt.Errorf with %w to preserve the error chain for errors.Is/As.
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	message := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %w", message, err)
}
