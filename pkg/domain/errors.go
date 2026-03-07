package domain

import (
	"fmt"
)

// NotFoundError represents a resource not found
type NotFoundError struct {
	Resource string
	ID       interface{}
	cause    error
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with id '%v' not found", e.Resource, e.ID)
}

// PublicMessage returns a sanitized message without ID for client responses
func (e *NotFoundError) PublicMessage() string {
	return fmt.Sprintf("%s not found", e.Resource)
}

func (e *NotFoundError) Unwrap() error {
	return e.cause
}

// NewNotFoundError creates a new NotFoundError
func NewNotFoundError(resource string, id interface{}, cause error) error {
	return &NotFoundError{Resource: resource, ID: id, cause: cause}
}

// FieldError represents a single field validation error
type FieldError struct {
	Field   string
	Message string
}

// ValidationError represents a validation error
type ValidationError struct {
	Errors []FieldError
}

func (e *ValidationError) Error() string {
	if len(e.Errors) == 0 {
		return "validation error"
	}
	if len(e.Errors) == 1 {
		return fmt.Sprintf("validation error on field '%s': %s", e.Errors[0].Field, e.Errors[0].Message)
	}
	return fmt.Sprintf("validation error: %d fields failed", len(e.Errors))
}

// AddFieldError adds a field error to the ValidationError
func (e *ValidationError) AddFieldError(field, message string) {
	e.Errors = append(e.Errors, FieldError{
		Field:   field,
		Message: message,
	})
}

// NewValidationError creates a new ValidationError with field errors
func NewValidationError(errors []FieldError) error {
	return &ValidationError{Errors: errors}
}

// ConflictError represents a conflict error
type ConflictError struct {
	Message string
	cause   error
}

func (e *ConflictError) Error() string {
	return e.Message
}

func (e *ConflictError) Unwrap() error {
	return e.cause
}

// NewConflictError creates a new ConflictError
func NewConflictError(message string, cause error) error {
	return &ConflictError{Message: message, cause: cause}
}

// AuthenticationError represents invalid credentials
type AuthenticationError struct {
	Message string
	cause   error
}

func (e *AuthenticationError) Error() string {
	return e.Message
}

func (e *AuthenticationError) Unwrap() error {
	return e.cause
}

// NewAuthenticationError creates a new AuthenticationError
func NewAuthenticationError(message string, cause error) error {
	return &AuthenticationError{Message: message, cause: cause}
}

// TimeoutError represents a request timeout
type TimeoutError struct {
	Message string
	cause   error
}

func (e *TimeoutError) Error() string {
	return e.Message
}

func (e *TimeoutError) Unwrap() error {
	return e.cause
}

// NewTimeoutError creates a new TimeoutError
func NewTimeoutError(message string, cause error) error {
	return &TimeoutError{Message: message, cause: cause}
}
