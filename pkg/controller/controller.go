package controller

import (
	"errors"
	"net/http"

	"lizobly/ctc-db-api/pkg/domain"
	pkgValidator "lizobly/ctc-db-api/pkg/validator"

	"github.com/go-playground/validator/v10"
	"github.com/iancoleman/strcase"

	"github.com/labstack/echo/v4"
)

type DataResponse[T any] struct {
	Data T `json:"data"`
}

type ErrorResponse struct {
	Message string       `json:"message"`
	Errors  []FieldError `json:"errors,omitempty"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Ok returns 200 OK status with data
func Ok[T any](ctx echo.Context, data T) error {
	return ctx.JSON(http.StatusOK, DataResponse[T]{
		Data: data,
	})
}

// Created returns 201 Created status with Location header
func Created[T any](ctx echo.Context, data T, location string) error {
	if location != "" {
		ctx.Response().Header().Set("Location", location)
	}
	return ctx.JSON(http.StatusCreated, DataResponse[T]{
		Data: data,
	})
}

// NoContent returns 204 No Content status with empty body
func NoContent(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNoContent)
}

// NotFound returns 404 Not Found status
func NotFound(ctx echo.Context, message string) error {
	return ctx.JSON(http.StatusNotFound, ErrorResponse{
		Message: message,
	})
}

// InternalError returns 500 Internal Server Error status
func InternalError(ctx echo.Context, message string) error {
	return ctx.JSON(http.StatusInternalServerError, ErrorResponse{
		Message: message,
	})
}

// RequestTimeout returns 408 Request Timeout status
func RequestTimeout(ctx echo.Context, message string) error {
	return ctx.JSON(http.StatusRequestTimeout, ErrorResponse{
		Message: message,
	})
}

// ResponseError returns a JSON response with the specified HTTP status
func ResponseError(ctx echo.Context, httpStatus int, message string) error {
	return ctx.JSON(httpStatus, ErrorResponse{
		Message: message,
	})
}

// ResponseErrorValidation returns 400 Bad Request with validation error details
func ResponseErrorValidation(ctx echo.Context, err error) error {
	var errMsg []FieldError

	// Handle go-playground validator errors (from ctx.Validate)
	if castedObject, ok := err.(validator.ValidationErrors); ok {
		validate := ctx.Get("validator").(*pkgValidator.CustomValidator)
		language := ctx.Request().Header.Get("Accept-Language")
		translator, _ := validate.Translator.FindTranslator(language)

		for _, e := range castedObject {
			errMsg = append(errMsg, FieldError{
				Field:   strcase.ToSnake(e.Field()),
				Message: e.Translate(translator),
			})
		}
	} else if validationErr, ok := err.(*domain.ValidationError); ok {
		// Handle domain ValidationError from services
		for _, fieldErr := range validationErr.Errors {
			errMsg = append(errMsg, FieldError{
				Field:   strcase.ToSnake(fieldErr.Field),
				Message: fieldErr.Message,
			})
		}
	} else {
		// Fallback for unknown validation error types
		errMsg = append(errMsg, FieldError{
			Field:   "general",
			Message: err.Error(),
		})
	}

	return ctx.JSON(http.StatusBadRequest, ErrorResponse{
		Message: "validation failed",
		Errors:  errMsg,
	})
}

// HandleServiceError maps domain errors to appropriate HTTP responses
func HandleServiceError(ctx echo.Context, err error, operation string) error {
	if err == nil {
		return nil
	}

	// Struct domain errors - use errors.As()
	var nfe *domain.NotFoundError
	if errors.As(err, &nfe) {
		return NotFound(ctx, err.Error())
	}

	var ce *domain.ConflictError
	if errors.As(err, &ce) {
		return ResponseError(ctx, http.StatusConflict, ce.Message)
	}

	var ae *domain.AuthenticationError
	if errors.As(err, &ae) {
		return ResponseError(ctx, http.StatusUnauthorized, ae.Message)
	}

	var ve *domain.ValidationError
	if errors.As(err, &ve) {
		return ResponseErrorValidation(ctx, err)
	}

	var te *domain.TimeoutError
	if errors.As(err, &te) {
		return RequestTimeout(ctx, te.Message)
	}

	// Unmapped errors - return 500
	return InternalError(ctx, "internal server error")
}
