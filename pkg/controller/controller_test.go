package controller

import (
	"encoding/json"
	"errors"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/logging"
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

func setupTestLogger() *logging.Logger {
	logger, _ := logging.NewDevelopmentLogger()
	return logger
}

func TestResponseErrorValidation_GoPlaygroundValidator(t *testing.T) {
	t.Skip("Skipping go-playground validator test - requires full Echo middleware setup")
	e := setupTestEcho()

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
			responseErr := ResponseErrorValidation(ctx, err)
			require.NoError(t, responseErr)

			// Check HTTP status
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Parse response
			var response ErrorResponse
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, "validation failed", response.Message)

			// Check that errors is an array
			assert.Len(t, response.Errors, len(tt.expectedFields))

			// Verify fields are present
			for _, expectedField := range tt.expectedFields {
				found := false
				for _, fieldErr := range response.Errors {
					if fieldErr.Field == expectedField {
						found = true
						assert.NotEmpty(t, fieldErr.Message)
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
			responseErr := ResponseErrorValidation(ctx, tt.err)
			require.NoError(t, responseErr)

			// Check HTTP status
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Parse response
			var response ErrorResponse
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, "validation failed", response.Message)

			// Check errors array
			require.Len(t, response.Errors, 1)

			assert.Equal(t, tt.expectedField, response.Errors[0].Field)
			assert.Equal(t, tt.expectedMsg, response.Errors[0].Message)
		})
	}
}

func TestResponseErrorValidation_DomainMultiFieldError(t *testing.T) {
	e := setupTestEcho()

	tests := []struct {
		name           string
		err            error
		expectedErrors []FieldError
		expectedStatus int
	}{
		{
			name: "multiple field validation errors",
			err: domain.NewValidationError([]domain.FieldError{
				{Field: "email", Message: "must be a valid email"},
				{Field: "password", Message: "must be at least 8 characters"},
				{Field: "username", Message: "is required"},
			}),
			expectedErrors: []FieldError{
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
			expectedErrors: []FieldError{
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
			responseErr := ResponseErrorValidation(ctx, tt.err)
			require.NoError(t, responseErr)

			// Check HTTP status
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Parse response
			var response ErrorResponse
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, "validation failed", response.Message)

			// Check errors array
			require.Len(t, response.Errors, len(tt.expectedErrors))

			// Verify each error field and message
			for i, expectedErr := range tt.expectedErrors {
				assert.Equal(t, expectedErr.Field, response.Errors[i].Field)
				assert.Equal(t, expectedErr.Message, response.Errors[i].Message)
			}
		})
	}
}

func TestResponseErrorValidation_AddFieldErrorHelper(t *testing.T) {
	e := setupTestEcho()

	t.Run("incrementally built validation error", func(t *testing.T) {
		// Create empty validation error and build it incrementally
		validationErr := &domain.ValidationError{}
		validationErr.AddFieldError("email", "must be a valid email")
		validationErr.AddFieldError("password", "too short")
		validationErr.AddFieldError("username", "already exists")

		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		responseErr := ResponseErrorValidation(ctx, validationErr)
		require.NoError(t, responseErr)

		// Parse response
		var response ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check all three errors are present
		require.Len(t, response.Errors, 3)

		expectedErrors := map[string]string{
			"email":    "must be a valid email",
			"password": "too short",
			"username": "already exists",
		}

		for _, fieldErr := range response.Errors {
			expectedMsg, found := expectedErrors[fieldErr.Field]
			assert.True(t, found, "unexpected field: %s", fieldErr.Field)
			assert.Equal(t, expectedMsg, fieldErr.Message)
		}
	})
}

func TestResponseErrorValidation_FallbackUnknownError(t *testing.T) {
	e := setupTestEcho()

	t.Run("unknown error type gets wrapped as general field", func(t *testing.T) {
		unknownErr := errors.New("some unexpected error")

		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		responseErr := ResponseErrorValidation(ctx, unknownErr)
		require.NoError(t, responseErr)

		// Check HTTP status
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		// Parse response
		var response ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "validation failed", response.Message)

		// Check error is wrapped in general field
		require.Len(t, response.Errors, 1)

		assert.Equal(t, "general", response.Errors[0].Field)
		assert.Equal(t, "some unexpected error", response.Errors[0].Message)
	})
}

func TestHandleServiceError_ValidationError(t *testing.T) {
	e := setupTestEcho()

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
			expectedMsg:    "validation failed",
		},
		{
			name: "multi field validation error from service",
			err: domain.NewValidationError([]domain.FieldError{
				{Field: "email", Message: "required"},
				{Field: "password", Message: "too short"},
			}),
			operation:      "update user",
			expectedStatus: http.StatusBadRequest,
			expectedMsg:    "validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			// Call HandleServiceError which should delegate to ResponseErrorValidation
			responseErr := HandleServiceError(ctx, tt.err, tt.operation, setupTestLogger())
			require.NoError(t, responseErr)

			// Check HTTP status
			assert.Equal(t, tt.expectedStatus, rec.Code)

			// Parse response
			var response ErrorResponse
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedMsg, response.Message)

			// Verify errors array exists and is not empty
			assert.NotEmpty(t, response.Errors)
		})
	}
}

// TestOk_SuccessResponse tests the Ok() response helper
func TestOk_SuccessResponse(t *testing.T) {
	e := setupTestEcho()

	type ResponseData struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	tests := []struct {
		name     string
		data     ResponseData
		wantCode int
	}{
		{
			name:     "returns single item",
			data:     ResponseData{ID: 1, Name: "John", Email: "john@example.com"},
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			err := Ok(ctx, tt.data)
			require.NoError(t, err)

			assert.Equal(t, tt.wantCode, rec.Code)

			var response DataResponse[ResponseData]
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.data.ID, response.Data.ID)
			assert.Equal(t, tt.data.Name, response.Data.Name)
		})
	}
}

// TestCreated_CreatedResponse tests the Created() response helper
func TestCreated_CreatedResponse(t *testing.T) {
	e := setupTestEcho()

	type CreateData struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	tests := []struct {
		name       string
		data       CreateData
		location   string
		wantCode   int
		wantHeader string
	}{
		{
			name:       "returns created resource with location header",
			data:       CreateData{ID: 1, Name: "NewItem"},
			location:   "/items/1",
			wantCode:   http.StatusCreated,
			wantHeader: "/items/1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			err := Created(ctx, tt.data, tt.location)
			require.NoError(t, err)

			assert.Equal(t, tt.wantCode, rec.Code)
			assert.Equal(t, tt.wantHeader, rec.Header().Get("Location"))

			var response DataResponse[CreateData]
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.data.ID, response.Data.ID)
		})
	}
}

// TestNoContent_NoContentResponse tests the NoContent() response helper
func TestNoContent_NoContentResponse(t *testing.T) {
	e := setupTestEcho()

	t.Run("returns no content status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/test", nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)

		err := NoContent(ctx)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Empty(t, rec.Body.String())
	})
}

// TestNotFound_NotFoundResponse tests the NotFound() response helper
func TestNotFound_NotFoundResponse(t *testing.T) {
	e := setupTestEcho()

	tests := []struct {
		name     string
		message  string
		wantCode int
	}{
		{
			name:     "returns not found with message",
			message:  "resource not found",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "returns not found with custom message",
			message:  "user with id 123 not found",
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			err := NotFound(ctx, tt.message)
			require.NoError(t, err)

			assert.Equal(t, tt.wantCode, rec.Code)

			var response ErrorResponse
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.message, response.Message)
		})
	}
}

// TestInternalError_InternalServerErrorResponse tests the InternalError() response helper
func TestInternalError_InternalServerErrorResponse(t *testing.T) {
	e := setupTestEcho()

	tests := []struct {
		name     string
		message  string
		wantCode int
	}{
		{
			name:     "returns internal server error",
			message:  "internal server error",
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "returns with custom error message",
			message:  "database connection failed",
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			err := InternalError(ctx, tt.message)
			require.NoError(t, err)

			assert.Equal(t, tt.wantCode, rec.Code)

			var response ErrorResponse
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.message, response.Message)
		})
	}
}

// TestResponseError_CustomHTTPStatus tests the ResponseError() response helper
func TestResponseError_CustomHTTPStatus(t *testing.T) {
	e := setupTestEcho()

	tests := []struct {
		name       string
		httpStatus int
		message    string
	}{
		{
			name:       "returns custom HTTP status",
			httpStatus: http.StatusForbidden,
			message:    "access forbidden",
		},
		{
			name:       "returns service unavailable",
			httpStatus: http.StatusServiceUnavailable,
			message:    "service temporarily unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			err := ResponseError(ctx, tt.httpStatus, tt.message)
			require.NoError(t, err)

			assert.Equal(t, tt.httpStatus, rec.Code)

			var response ErrorResponse
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.message, response.Message)
		})
	}
}

// TestHandleServiceError_ErrorTypes tests HandleServiceError with various error types
func TestHandleServiceError_ErrorTypes(t *testing.T) {
	e := setupTestEcho()

	tests := []struct {
		name              string
		err               error
		operation         string
		expectedStatus    int
		expectedMsg       string
		assertMsgContains bool // if true, use Contains instead of Equal
	}{
		{
			name:              "handles NotFoundError from service",
			err:               domain.NewNotFoundError("user", 123, nil),
			operation:         "get user",
			expectedStatus:    http.StatusNotFound,
			expectedMsg:       "not found",
			assertMsgContains: true,
		},
		{
			name:           "handles ConflictError from service",
			err:            domain.NewConflictError("email already exists", nil),
			operation:      "create user",
			expectedStatus: http.StatusConflict,
			expectedMsg:    "email already exists",
		},
		{
			name:           "handles AuthenticationError from service",
			err:            domain.NewAuthenticationError("invalid credentials", nil),
			operation:      "login",
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "invalid credentials",
		},
		{
			name:           "handles TimeoutError from service",
			err:            domain.NewTimeoutError("request timeout", nil),
			operation:      "fetch data",
			expectedStatus: http.StatusRequestTimeout,
			expectedMsg:    "request timeout",
		},
		{
			name:           "handles generic error as internal server error",
			err:            errors.New("database connection failed"),
			operation:      "fetch data",
			expectedStatus: http.StatusInternalServerError,
			expectedMsg:    "", // Message will be non-empty but not predictable
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			responseErr := HandleServiceError(ctx, tt.err, tt.operation, setupTestLogger())
			require.NoError(t, responseErr)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			var response ErrorResponse
			err := json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedMsg != "" {
				if tt.assertMsgContains {
					assert.Contains(t, response.Message, tt.expectedMsg)
				} else {
					assert.Equal(t, tt.expectedMsg, response.Message)
				}
			} else {
				assert.NotEmpty(t, response.Message)
			}
		})
	}
}

// TestRequestTimeout_Response tests the RequestTimeout() response helper
func TestRequestTimeout_Response(t *testing.T) {
	e := setupTestEcho()

	tests := []struct {
		name     string
		message  string
		wantCode int
	}{
		{
			name:     "returns request timeout with message",
			message:  "request timeout",
			wantCode: http.StatusRequestTimeout,
		},
		{
			name:     "returns request timeout with custom message",
			message:  "operation exceeded time limit",
			wantCode: http.StatusRequestTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			err := RequestTimeout(ctx, tt.message)
			require.NoError(t, err)

			assert.Equal(t, tt.wantCode, rec.Code)

			var response ErrorResponse
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.message, response.Message)
		})
	}
}
