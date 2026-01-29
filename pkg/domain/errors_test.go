package domain

import (
	"errors"
	"testing"
)

func TestNewValidationError(t *testing.T) {
	tests := []struct {
		name          string
		errors        []FieldError
		expectedError string
	}{
		{
			name: "single field error",
			errors: []FieldError{
				{Field: "email", Message: "must be a valid email"},
			},
			expectedError: "validation error on field 'email': must be a valid email",
		},
		{
			name: "multiple field errors",
			errors: []FieldError{
				{Field: "email", Message: "must be a valid email"},
				{Field: "password", Message: "must be at least 8 characters"},
			},
			expectedError: "validation error: 2 fields failed",
		},
		{
			name: "three field errors",
			errors: []FieldError{
				{Field: "email", Message: "must be a valid email"},
				{Field: "password", Message: "must be at least 8 characters"},
				{Field: "username", Message: "is required"},
			},
			expectedError: "validation error: 3 fields failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewValidationError(tt.errors)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Error() != tt.expectedError {
				t.Errorf("expected error message '%s', got '%s'", tt.expectedError, err.Error())
			}

			// Verify it's a ValidationError
			var ve *ValidationError
			if !errors.As(err, &ve) {
				t.Error("errors.As should return true for ValidationError")
			}

			// Verify errors are set correctly
			validationErr, ok := err.(*ValidationError)
			if !ok {
				t.Fatal("expected *ValidationError type")
			}
			if len(validationErr.Errors) != len(tt.errors) {
				t.Errorf("expected %d errors, got %d", len(tt.errors), len(validationErr.Errors))
			}
			for i, fieldErr := range validationErr.Errors {
				if fieldErr.Field != tt.errors[i].Field {
					t.Errorf("error %d: expected field '%s', got '%s'", i, tt.errors[i].Field, fieldErr.Field)
				}
				if fieldErr.Message != tt.errors[i].Message {
					t.Errorf("error %d: expected message '%s', got '%s'", i, tt.errors[i].Message, fieldErr.Message)
				}
			}
		})
	}
}

func TestValidationError_AddFieldError(t *testing.T) {
	t.Run("add field errors incrementally", func(t *testing.T) {
		err := &ValidationError{}

		// Start with no errors
		if len(err.Errors) != 0 {
			t.Errorf("expected 0 errors, got %d", len(err.Errors))
		}

		// Add first error
		err.AddFieldError("email", "must be a valid email")
		if len(err.Errors) != 1 {
			t.Errorf("expected 1 error, got %d", len(err.Errors))
		}
		if err.Errors[0].Field != "email" {
			t.Errorf("expected field 'email', got '%s'", err.Errors[0].Field)
		}
		if err.Errors[0].Message != "must be a valid email" {
			t.Errorf("expected message 'must be a valid email', got '%s'", err.Errors[0].Message)
		}

		// Add second error
		err.AddFieldError("password", "must be at least 8 characters")
		if len(err.Errors) != 2 {
			t.Errorf("expected 2 errors, got %d", len(err.Errors))
		}
		if err.Errors[1].Field != "password" {
			t.Errorf("expected field 'password', got '%s'", err.Errors[1].Field)
		}

		// Verify Error() message
		expectedError := "validation error: 2 fields failed"
		if err.Error() != expectedError {
			t.Errorf("expected error message '%s', got '%s'", expectedError, err.Error())
		}
	})
}

func TestIsValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "ValidationError single field",
			err:      NewValidationError([]FieldError{{Field: "field", Message: "test"}}),
			expected: true,
		},
		{
			name:     "ValidationError multi field",
			err:      NewValidationError([]FieldError{{Field: "test", Message: "msg"}}),
			expected: true,
		},
		{
			name:     "wrapped ValidationError",
			err:      errors.Join(NewValidationError([]FieldError{{Field: "field", Message: "test"}}), errors.New("another error")),
			expected: true,
		},
		{
			name:     "not a ValidationError",
			err:      errors.New("regular error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "NotFoundError",
			err:      NewNotFoundError("user", 123),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ve *ValidationError
			result := errors.As(tt.err, &ve)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestNewNotFoundError_Success tests NotFoundError creation with various ID types
func TestNewNotFoundError_Success(t *testing.T) {
	tests := []struct {
		name          string
		resource      string
		id            interface{}
		expectedError string
	}{
		{
			name:          "integer ID",
			resource:      "user",
			id:            123,
			expectedError: "user with id '123' not found",
		},
		{
			name:          "string ID",
			resource:      "post",
			id:            "abc-123-def",
			expectedError: "post with id 'abc-123-def' not found",
		},
		{
			name:          "UUID ID",
			resource:      "article",
			id:            "550e8400-e29b-41d4-a716-446655440000",
			expectedError: "article with id '550e8400-e29b-41d4-a716-446655440000' not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewNotFoundError(tt.resource, tt.id)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Error() != tt.expectedError {
				t.Errorf("expected error message '%s', got '%s'", tt.expectedError, err.Error())
			}

			// Verify it's a NotFoundError
			var nfe *NotFoundError
			if !errors.As(err, &nfe) {
				t.Error("errors.As should return true for NotFoundError")
			}

			notFoundErr, ok := err.(*NotFoundError)
			if !ok {
				t.Fatal("expected *NotFoundError type")
			}
			if notFoundErr.Resource != tt.resource {
				t.Errorf("expected resource '%s', got '%s'", tt.resource, notFoundErr.Resource)
			}
			if notFoundErr.ID != tt.id {
				t.Errorf("expected id '%v', got '%v'", tt.id, notFoundErr.ID)
			}
		})
	}
}

// TestNewNotFoundError_ErrorMethod tests NotFoundError Error() method output
func TestNewNotFoundError_ErrorMethod(t *testing.T) {
	err := NewNotFoundError("traveller", 456)
	nfeErr := err.(*NotFoundError)

	if nfeErr.Resource != "traveller" {
		t.Errorf("expected resource 'traveller', got '%s'", nfeErr.Resource)
	}
	if nfeErr.ID != 456 {
		t.Errorf("expected id 456, got %v", nfeErr.ID)
	}

	expectedMsg := "traveller with id '456' not found"
	if err.Error() != expectedMsg {
		t.Errorf("expected '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestNewConflictError_Success tests ConflictError creation
func TestNewConflictError_Success(t *testing.T) {
	tests := []struct {
		name          string
		message       string
		expectedError string
	}{
		{
			name:          "duplicate email",
			message:       "email already exists",
			expectedError: "email already exists",
		},
		{
			name:          "username conflict",
			message:       "username is already taken",
			expectedError: "username is already taken",
		},
		{
			name:          "resource exists",
			message:       "resource with the same name already exists",
			expectedError: "resource with the same name already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewConflictError(tt.message)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Error() != tt.expectedError {
				t.Errorf("expected error message '%s', got '%s'", tt.expectedError, err.Error())
			}

			// Verify it's a ConflictError
			var ce *ConflictError
			if !errors.As(err, &ce) {
				t.Error("errors.As should return true for ConflictError")
			}

			conflictErr, ok := err.(*ConflictError)
			if !ok {
				t.Fatal("expected *ConflictError type")
			}
			if conflictErr.Message != tt.message {
				t.Errorf("expected message '%s', got '%s'", tt.message, conflictErr.Message)
			}
		})
	}
}

// TestNewConflictError_ErrorMethod tests ConflictError Error() method
func TestNewConflictError_ErrorMethod(t *testing.T) {
	message := "duplicate key value violates unique constraint"
	err := NewConflictError(message)

	if err.Error() != message {
		t.Errorf("expected '%s', got '%s'", message, err.Error())
	}
}

// TestNewAuthenticationError_Success tests AuthenticationError creation
func TestNewAuthenticationError_Success(t *testing.T) {
	tests := []struct {
		name          string
		message       string
		expectedError string
	}{
		{
			name:          "token expired",
			message:       "authentication token has expired",
			expectedError: "authentication token has expired",
		},
		{
			name:          "unauthorized access",
			message:       "unauthorized access to resource",
			expectedError: "unauthorized access to resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAuthenticationError(tt.message)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Error() != tt.expectedError {
				t.Errorf("expected error message '%s', got '%s'", tt.expectedError, err.Error())
			}

			// Verify it's an AuthenticationError
			var ae *AuthenticationError
			if !errors.As(err, &ae) {
				t.Error("errors.As should return true for AuthenticationError")
			}

			authErr, ok := err.(*AuthenticationError)
			if !ok {
				t.Fatal("expected *AuthenticationError type")
			}
			if authErr.Message != tt.message {
				t.Errorf("expected message '%s', got '%s'", tt.message, authErr.Message)
			}
		})
	}
}

// TestNewAuthenticationError_ErrorMethod tests AuthenticationError Error() method
func TestNewAuthenticationError_ErrorMethod(t *testing.T) {
	message := "missing or invalid authorization header"
	err := NewAuthenticationError(message)

	if err.Error() != message {
		t.Errorf("expected '%s', got '%s'", message, err.Error())
	}
}

// TestErrorTypes_Differentiation tests that different error types are distinct
func TestErrorTypes_Differentiation(t *testing.T) {
	notFoundErr := NewNotFoundError("user", 123)
	conflictErr := NewConflictError("duplicate entry")
	authErr := NewAuthenticationError("invalid credentials")
	validationErr := NewValidationError([]FieldError{{Field: "email", Message: "required"}})

	// Test that each error is its correct type
	var nfe *NotFoundError
	var ce *ConflictError
	var ae *AuthenticationError
	var ve *ValidationError

	if !errors.As(notFoundErr, &nfe) {
		t.Error("notFoundErr should be NotFoundError")
	}
	if !errors.As(conflictErr, &ce) {
		t.Error("conflictErr should be ConflictError")
	}
	if !errors.As(authErr, &ae) {
		t.Error("authErr should be AuthenticationError")
	}
	if !errors.As(validationErr, &ve) {
		t.Error("validationErr should be ValidationError")
	}

	// Test that errors are not confused with each other
	if errors.As(notFoundErr, &ce) {
		t.Error("notFoundErr should not be ConflictError")
	}
	if errors.As(conflictErr, &ae) {
		t.Error("conflictErr should not be AuthenticationError")
	}
	if errors.As(authErr, &ve) {
		t.Error("authErr should not be ValidationError")
	}
}
