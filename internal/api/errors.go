package api

import (
	"errors"
	"fmt"
)

// Common errors
var (
	// ErrNotFound indicates the requested resource was not found
	ErrNotFound = errors.New("not found")

	// ErrInvalidRequest indicates the request parameters are invalid
	ErrInvalidRequest = errors.New("invalid request")

	// ErrServerError indicates a server-side error
	ErrServerError = errors.New("server error")

	// ErrTimeout indicates the request timed out
	ErrTimeout = errors.New("request timed out")

	// ErrNoResults indicates no results were found
	ErrNoResults = errors.New("no results found")
)

// APIError represents an error returned by the bahn.de API
type APIError struct {
	StatusCode int
	Status     string
	Endpoint   string
	Message    string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("API error %d (%s): %s", e.StatusCode, e.Endpoint, e.Message)
	}
	return fmt.Sprintf("API error %d: %s (endpoint: %s)", e.StatusCode, e.Status, e.Endpoint)
}

// Is implements errors.Is for APIError
func (e *APIError) Is(target error) bool {
	switch target {
	case ErrNotFound:
		return e.StatusCode == 404
	case ErrServerError:
		return e.StatusCode >= 500
	case ErrInvalidRequest:
		return e.StatusCode == 400
	}
	return false
}

// NewAPIError creates a new API error
func NewAPIError(statusCode int, status, endpoint string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Status:     status,
		Endpoint:   endpoint,
	}
}

// NewAPIErrorWithMessage creates a new API error with a custom message
func NewAPIErrorWithMessage(statusCode int, endpoint, message string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Endpoint:   endpoint,
		Message:    message,
	}
}

// ValidationError represents a validation error for request parameters
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// Common validation errors
func ErrMissingField(field string) error {
	return NewValidationError(field, "field is required")
}

func ErrInvalidFormat(field, expected string) error {
	return NewValidationError(field, fmt.Sprintf("invalid format, expected %s", expected))
}

func ErrInvalidValue(field string, value interface{}) error {
	return NewValidationError(field, fmt.Sprintf("invalid value: %v", value))
}
