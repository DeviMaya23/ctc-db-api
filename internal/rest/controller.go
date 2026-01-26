package rest

import (
	"net/http"

	"lizobly/ctc-db-api/pkg/domain"
	pkgValidator "lizobly/ctc-db-api/pkg/validator"

	"github.com/go-playground/validator/v10"
	"github.com/iancoleman/strcase"

	"github.com/labstack/echo/v4"
)

type Controller struct {
}

type StandardAPIResponse struct {
	Message  string      `json:"message"`
	Data     interface{} `json:"data"`
	Errors   interface{} `json:"errors"`
	Metadata interface{} `json:"metadata"`
}

type ValidationErrorFields struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (c Controller) Ok(ctx echo.Context, message string, data, metadata interface{}) error {

	return ctx.JSON(http.StatusOK, StandardAPIResponse{
		Message:  message,
		Data:     data,
		Metadata: metadata,
	})
}

// Created returns 201 Created status with Location header
func (c Controller) Created(ctx echo.Context, message string, data interface{}, location string) error {
	if location != "" {
		ctx.Response().Header().Set("Location", location)
	}
	return ctx.JSON(http.StatusCreated, StandardAPIResponse{
		Message: message,
		Data:    data,
	})
}

// NoContent returns 204 No Content status with empty body
func (c Controller) NoContent(ctx echo.Context) error {
	return ctx.NoContent(http.StatusNoContent)
}

// NotFound returns 404 Not Found status
func (c Controller) NotFound(ctx echo.Context, message string) error {
	return ctx.JSON(http.StatusNotFound, StandardAPIResponse{
		Message: message,
	})
}

// InternalError returns 500 Internal Server Error status
func (c Controller) InternalError(ctx echo.Context, message string, errorData interface{}) error {
	return ctx.JSON(http.StatusInternalServerError, StandardAPIResponse{
		Message: message,
		Errors:  errorData,
	})
}

func (c Controller) ResponseError(ctx echo.Context, httpStatus int, message string, errorData interface{}) error {

	return ctx.JSON(httpStatus, StandardAPIResponse{
		Message: message,
		Errors:  errorData,
	})
}

func (c Controller) ResponseErrorValidation(ctx echo.Context, err error) error {

	// TODO : non go validator error
	// _, ok := err.(*echo.HTTPError)
	// if !ok {
	// 	report = echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	// }

	var errMsg []ValidationErrorFields
	validate := ctx.Get("validator").(*pkgValidator.CustomValidator)
	language := ctx.Request().Header.Get("Accept-Language")
	translator, _ := validate.Translator.FindTranslator(language)

	if castedObject, ok := err.(validator.ValidationErrors); ok {
		for _, e := range castedObject {
			errMsg = append(errMsg, ValidationErrorFields{
				Field:   strcase.ToSnake(e.Field()),
				Message: e.Translate(translator),
			})
		}
	}

	return ctx.JSON(http.StatusBadRequest, StandardAPIResponse{
		Message: "error validation",
		Errors:  errMsg,
	})
}

// HandleServiceError maps domain errors to appropriate HTTP responses
// Returns true if error was handled, false if it's an unhandled error (should return 500)
func (c Controller) HandleServiceError(ctx echo.Context, err error, operation string) error {
	if err == nil {
		return nil
	}

	// Import domain package at top of file to use domain.IsNotFoundError, etc.
	// Check domain errors and map to HTTP status
	if domain.IsNotFoundError(err) {
		return c.NotFound(ctx, err.Error())
	}
	if domain.IsValidationError(err) {
		return c.ResponseError(ctx, http.StatusBadRequest, "error "+operation, err.Error())
	}
	if domain.IsConflictError(err) {
		return c.ResponseError(ctx, http.StatusConflict, "error "+operation, err.Error())
	}
	if domain.IsAuthenticationError(err) {
		return c.ResponseError(ctx, http.StatusUnauthorized, "error "+operation, err.Error())
	}

	// Unhandled error - return 500
	return c.InternalError(ctx, "error "+operation, err.Error())
}
