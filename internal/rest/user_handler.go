package rest

import (
	"context"
	"errors"
	"lizobly/ctc-db-api/pkg/domain"
	"net/http"

	"github.com/labstack/echo/v4"
)

type UserService interface {
	Login(ctx context.Context, req domain.LoginRequest) (res domain.LoginResponse, err error)
}

type UserHandler struct {
	Controller
	Service UserService
}

func NewUserHandler(e *echo.Group, svc UserService) *UserHandler {
	handler := &UserHandler{
		Service: svc,
	}

	e.POST("/login", handler.Login)

	return handler
}

func (h *UserHandler) Login(ctx echo.Context) error {

	var request domain.LoginRequest

	err := ctx.Bind(&request)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	err = ctx.Validate(&request)
	if err != nil {
		return h.ResponseErrorValidation(ctx, err)
	}

	res, err := h.Service.Login(ctx.Request().Context(), request)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidPassword):
			return h.ResponseError(ctx, http.StatusUnauthorized, "error", err.Error())
		case errors.Is(err, domain.ErrUserNotFound):
			return h.ResponseError(ctx, http.StatusUnauthorized, "error", err.Error())
		default:
			return h.InternalError(ctx, "error", err.Error())
		}
	}

	return h.Ok(ctx, "success", res, nil)
}
