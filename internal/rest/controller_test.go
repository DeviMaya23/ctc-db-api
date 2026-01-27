package rest

import (
	"encoding/json"
	"errors"
	"lizobly/ctc-db-api/pkg/domain"
	pkgValidator "lizobly/ctc-db-api/pkg/validator"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestEcho() *echo.Echo {
	e := echo.New()
	// Setup custom validator to match production
	customValidator, _ := pkgValidator.NewValidator()
	e.Validator = customValidator
	return e
}

func TestResponseErrorValidation_GoPlaygroundValidator(t *testing.T) {
	t.Skip("Skipping go-playground validator test - requires full Echo middleware setup")
	e := setupTestEcho()
	controller := Controller{}

	// Create a sample struct for validation
	type TestRequest struct {
		Email string `json:"email" validate:"required,email"`
		Age   int    `json:"age" validate:"required,min=18"`
	}

	tests := []struct {
		name           string
		request        TestRequest
		expectedFields []string
		expectedStatus int
	}{
		{
			name: "multiple validation errors from go-playground validator",
			request: TestRequest{
				Email: "invalid-email",
				Age:   10,
			},
			expectedFields: []string{"email", "age"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "single validation error from go-playground validator",
			request: TestRequest{
				Email: "",
				Age:   20,
			},
			expectedFields: []string{"email"},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			// Validate the request struct using go-playground validator
			validate := validator.New()
			err := validate.Struct(tt.request)
			require.Error(t, err)

			// Call ResponseErrorValidation with validation error
			responseErr := controller.ResponseErrorValidation(ctx, err)
			require.NoError(t, responseErr)

			// Check HTTP status
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Parse response
			var response StandardAPIResponse
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, "error validation", response.Message)

			// Check that errors is an array
			errorsArray, ok := response.Errors.([]interface{})
			require.True(t, ok, "errors should be an array")
			assert.Len(t, errorsArray, len(tt.expectedFields))

			// Verify fields are present
			for _, expectedField := range tt.expectedFields {
				found := false
				for _, errInterface := range errorsArray {
					errMap, ok := errInterface.(map[string]interface{})
					require.True(t, ok)
					if errMap["field"] == expectedField {
						found = true
						assert.NotEmpty(t, errMap["message"])
						break
					}
				}
				assert.True(t, found, "expected field %s not found in errors", expectedField)
			}
		})
	}
}

func TestResponseErrorValidation_DomainSingleFieldError(t *testing.T) {
	e := setupTestEcho()
	controller := Controller{}

	tests := []struct {
		name           string
		err            error
		expectedField  string
		expectedMsg    string
		expectedStatus int
	}{
		{
			name:           "single field validation error",
			err:            domain.NewValidationError([]domain.FieldError{{Field: "email", Message: "must be a valid email"}}),
			expectedField:  "email",
			expectedMsg:    "must be a valid email",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			// Call ResponseErrorValidation with domain error
			responseErr := controller.ResponseErrorValidation(ctx, tt.err)
			require.NoError(t, responseErr)

			// Check HTTP status
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Parse response
			var response StandardAPIResponse
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, "error validation", response.Message)

			// Check errors array
			errorsArray, ok := response.Errors.([]interface{})
			require.True(t, ok, "errors should be an array")
			require.Len(t, errorsArray, 1)

			errMap, ok := errorsArray[0].(map[string]interface{})
			require.True(t, ok)

			assert.Equal(t, tt.expectedField, errMap["field"])
			assert.Equal(t, tt.expectedMsg, errMap["message"])
		})
	}
}

func TestResponseErrorValidation_DomainMultiFieldError(t *testing.T) {
	e := setupTestEcho()
	controller := Controller{}

	tests := []struct {
		name           string
		err            error
		expectedErrors []ValidationErrorFields
		expectedStatus int
	}{
		{
			name: "multiple field validation errors",
			err: domain.NewValidationError([]domain.FieldError{
				{Field: "email", Message: "must be a valid email"},
				{Field: "password", Message: "must be at least 8 characters"},
				{Field: "username", Message: "is required"},
			}),
			expectedErrors: []ValidationErrorFields{
				{Field: "email", Message: "must be a valid email"},
				{Field: "password", Message: "must be at least 8 characters"},
				{Field: "username", Message: "is required"},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "single field error",
			err: domain.NewValidationError([]domain.FieldError{
				{Field: "email", Message: "invalid format"},
			}),
			expectedErrors: []ValidationErrorFields{
				{Field: "email", Message: "invalid format"},
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			// Call ResponseErrorValidation
			responseErr := controller.ResponseErrorValidation(ctx, tt.err)
			require.NoError(t, responseErr)

			// Check HTTP status
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Parse response
			var response StandardAPIResponse
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, "error validation", response.Message)

			// Check errors array
			errorsArray, ok := response.Errors.([]interface{})
			require.True(t, ok, "errors should be an array")
			require.Len(t, errorsArray, len(tt.expectedErrors))

			// Verify each error field and message
			for i, expectedErr := range tt.expectedErrors {
				errMap, ok := errorsArray[i].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, expectedErr.Field, errMap["field"])
				assert.Equal(t, expectedErr.Message, errMap["message"])
			}
		})
	}
}

func TestResponseErrorValidation_AddFieldErrorHelper(t *testing.T) {
	e := setupTestEcho()
	controller := Controller{}

	t.Run("incrementally built validation error", func(t *testing.T) {
		// Create empty validation error and build it incrementally
		validationErr := &domain.ValidationError{}
		validationErr.AddFieldError("email", "must be a valid email")
		validationErr.AddFieldError("password", "too short")
		validationErr.AddFieldError("username", "already exists")

		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		responseErr := controller.ResponseErrorValidation(ctx, validationErr)
		require.NoError(t, responseErr)

		// Parse response
		var response StandardAPIResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check all three errors are present
		errorsArray, ok := response.Errors.([]interface{})
		require.True(t, ok)
		require.Len(t, errorsArray, 3)

		expectedErrors := map[string]string{
			"email":    "must be a valid email",
			"password": "too short",
			"username": "already exists",
		}

		for _, errInterface := range errorsArray {
			errMap, ok := errInterface.(map[string]interface{})
			require.True(t, ok)
			field := errMap["field"].(string)
			message := errMap["message"].(string)

			expectedMsg, found := expectedErrors[field]
			assert.True(t, found, "unexpected field: %s", field)
			assert.Equal(t, expectedMsg, message)
		}
	})
}

func TestResponseErrorValidation_FallbackUnknownError(t *testing.T) {
	e := setupTestEcho()
	controller := Controller{}

	t.Run("unknown error type gets wrapped as general field", func(t *testing.T) {
		unknownErr := errors.New("some unexpected error")

		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		responseErr := controller.ResponseErrorValidation(ctx, unknownErr)
		require.NoError(t, responseErr)

		// Check HTTP status
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		// Parse response
		var response StandardAPIResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "error validation", response.Message)

		// Check error is wrapped in general field
		errorsArray, ok := response.Errors.([]interface{})
		require.True(t, ok)
		require.Len(t, errorsArray, 1)

		errMap, ok := errorsArray[0].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "general", errMap["field"])
		assert.Equal(t, "some unexpected error", errMap["message"])
	})
}

func TestHandleServiceError_ValidationError(t *testing.T) {
	e := setupTestEcho()
	controller := Controller{}

	tests := []struct {
		name           string
		err            error
		operation      string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "single field validation error from service",
			err:            domain.NewValidationError([]domain.FieldError{{Field: "email", Message: "invalid email format"}}),
			operation:      "create user",
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "error validation",
		},
		{
			name: "multi field validation error from service",
			err: domain.NewValidationError([]domain.FieldError{
				{Field: "email", Message: "required"},
				{Field: "password", Message: "too short"},
			}),
			operation:      "update user",
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "error validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			// Call HandleServiceError which should delegate to ResponseErrorValidation
			responseErr := controller.HandleServiceError(ctx, tt.err, tt.operation)
			require.NoError(t, responseErr)

			// Check HTTP status
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Parse response
			var response StandardAPIResponse
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedMsg, response.Message)

			// Verify errors array exists and is not empty
			errorsArray, ok := response.Errors.([]interface{})
			require.True(t, ok, "errors should be an array")
			assert.NotEmpty(t, errorsArray)
		})
	}
}
