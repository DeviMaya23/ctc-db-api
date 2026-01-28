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
