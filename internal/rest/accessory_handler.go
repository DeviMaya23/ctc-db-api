package rest

import (
	"context"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/helpers"
	"net/http"

	"github.com/labstack/echo/v4"
)

type AccessoryService interface {
	GetList(ctx context.Context, filter domain.ListAccessoryRequest, params helpers.PaginationParams) (res helpers.PaginatedResponse[domain.AccessoryListItemResponse], err error)
}

type AccessoryHandler struct {
	Controller
	Service AccessoryService
}

func NewAccessoryHandler(e *echo.Group, svc AccessoryService) *AccessoryHandler {
	handler := &AccessoryHandler{
		Service: svc,
	}
	group := e.Group("/accessories")

	group.GET("", handler.GetList)

	return handler
}

// GetList godoc
//
//	@Summary		Get list of accessories
//	@Description	get accessory list with optional filters, ordering, and pagination
//	@Tags			accessories
//	@Accept			json
//	@Produce		json
//	@Param			owner	query	string	false	"Filter by traveller name (case insensitive)"
//	@Param			effect			query	string	false	"Filter by effect (case insensitive)"
//	@Param			order_by		query	string	false	"Order by field (hp, sp, patk, pdef, eatk, edef, spd, crit)"
//	@Param			order_dir		query	string	false	"Order direction (asc, desc)"
//	@Param			page			query	int		false	"Page number (default 1)"
//	@Param			page_size		query	int		false	"Page size (default 10, max 100)"
//	@Success		200	{object}	helpers.PaginatedResponse[domain.AccessoryListItemResponse]
//	@Failure		400	{object}	StandardAPIResponse
//	@Failure		500	{object}	StandardAPIResponse
//	@Router			/accessories [get]
func (h *AccessoryHandler) GetList(ctx echo.Context) error {
	var filter domain.ListAccessoryRequest
	err := ctx.Bind(&filter)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	err = ctx.Validate(&filter)
	if err != nil {
		return h.ResponseErrorValidation(ctx, err)
	}

	var params helpers.PaginationParams
	err = ctx.Bind(&params)
	if err != nil {
		return h.ResponseError(ctx, http.StatusBadRequest, "error validation", err.Error())
	}

	result, err := h.Service.GetList(ctx.Request().Context(), filter, params)
	if err != nil {
		return h.InternalError(ctx, "error get data", err.Error())
	}

	// Set cache headers for list responses
	helpers.SetListCacheHeaders(ctx)

	return h.Ok(ctx, "success", result, nil)
}
