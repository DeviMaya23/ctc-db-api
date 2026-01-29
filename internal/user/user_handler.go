package user

import (
	"context"
	"lizobly/ctc-db-api/pkg/controller"
	"lizobly/ctc-db-api/pkg/domain"
	"net/http"

	"github.com/labstack/echo/v4"
)

type UserService interface {
	Login(ctx context.Context, req domain.LoginRequest) (res domain.LoginResponse, err error)
}

type UserHandler struct {
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
		return controller.ResponseError(ctx, http.StatusBadRequest, "invalid request body")
	}

	err = ctx.Validate(&request)
	if err != nil {
		return controller.ResponseErrorValidation(ctx, err)
	}

	res, err := h.Service.Login(ctx.Request().Context(), request)
	if err != nil {
		return controller.HandleServiceError(ctx, err, "login")
	}

	return controller.Ok(ctx, res)
}
