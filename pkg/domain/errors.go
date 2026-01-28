package domain

import (
	"fmt"
)

// NotFoundError represents a resource not found
type NotFoundError struct {
	Resource string
	ID       interface{}
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with id '%v' not found", e.Resource, e.ID)
}

// NewNotFoundError creates a new NotFoundError
func NewNotFoundError(resource string, id interface{}) error {
	return &NotFoundError{Resource: resource, ID: id}
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
}

func (e *ConflictError) Error() string {
	return e.Message
}

// NewConflictError creates a new ConflictError
func NewConflictError(message string) error {
	return &ConflictError{Message: message}
}

// AuthenticationError represents invalid credentials
type AuthenticationError struct {
	Message string
}

func (e *AuthenticationError) Error() string {
	return e.Message
}

// NewAuthenticationError creates a new AuthenticationError
func NewAuthenticationError(message string) error {
	return &AuthenticationError{Message: message}
}
