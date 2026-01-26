package domain

import (
	"errors"
	"fmt"
)

// NotFoundError represents a resource not found error (404)
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

// IsNotFoundError checks if an error is a NotFoundError
func IsNotFoundError(err error) bool {
	return errors.As(err, new(*NotFoundError))
}

// ValidationError represents a validation error (400)
type ValidationError struct {
	Message string
	Field   string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// NewValidationError creates a new ValidationError
func NewValidationError(message string, field string) error {
	return &ValidationError{Message: message, Field: field}
}

// IsValidationError checks if an error is a ValidationError
func IsValidationError(err error) bool {
	return errors.As(err, new(*ValidationError))
}

// ConflictError represents a conflict error (409)
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

// IsConflictError checks if an error is a ConflictError
func IsConflictError(err error) bool {
	return errors.As(err, new(*ConflictError))
}

// AuthenticationError represents invalid credentials (401)
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

// IsAuthenticationError checks if an error is an AuthenticationError
func IsAuthenticationError(err error) bool {
	return errors.As(err, new(*AuthenticationError))
}

// InternalError represents an internal server error (500)
type InternalError struct {
	Message string
}

func (e *InternalError) Error() string {
	return e.Message
}

// NewInternalError creates a new InternalError
func NewInternalError(message string) error {
	return &InternalError{Message: message}
}

// IsInternalError checks if an error is an InternalError
func IsInternalError(err error) bool {
	return errors.As(err, new(*InternalError))
}
