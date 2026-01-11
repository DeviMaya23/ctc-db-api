package rest

import (
	"context"
	"lizobly/ctc-db-api/pkg/domain"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type TravellerService interface {
	GetByID(ctx context.Context, id int) (res domain.Traveller, err error)
	Create(ctx context.Context, input domain.CreateTravellerRequest) (err error)
	Update(ctx context.Context, id int, input domain.UpdateTravellerRequest) (err error)
	Delete(ctx context.Context, id int) (err error)
}

type TravellerHandler struct {
	Controller
	Service TravellerService
}

func NewTravellerHandler(e *echo.Group, svc TravellerService) *TravellerHandler {
	handler := &TravellerHandler{
		Service: svc,
	}
	group := e.Group("/travellers")

	group.GET("/:id", handler.GetByID)
	group.POST("", handler.Create)
	group.PUT("/:id", handler.Update)
	group.DELETE("/:id", handler.Delete)

	return handler
}

// GetByID godoc
//
//	@Summary		Get by ID
//	@Description	get traveller information by ID
//	@Tags			accounts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Account ID"
//	@Success		200	{object}	domain.Traveller
//	@Failure		400	{object}	StandardAPIResponse
//	@Failure		404	{object}	StandardAPIResponse
//	@Failure		500	{object}	StandardAPIResponse
//	@Router			/travellers/{id} [get]
func (a *TravellerHandler) GetByID(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return a.ResponseError(ctx, http.StatusBadRequest, "error validation", "id not found")
	}

	traveller, err := a.Service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		return a.ResponseError(ctx, http.StatusBadRequest, "error get data", err.Error())
	}

	response := domain.ToTravellerResponse(traveller)
	return a.Ok(ctx, "success", response, nil)
}

func (a *TravellerHandler) Create(ctx echo.Context) error {

	var newTraveller domain.CreateTravellerRequest
	err := ctx.Bind(&newTraveller)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	err = ctx.Validate(&newTraveller)
	if err != nil {
		return a.ResponseErrorValidation(ctx, err)
	}

	err = a.Service.Create(ctx.Request().Context(), newTraveller)
	if err != nil {
		return a.ResponseError(ctx, http.StatusBadRequest, "error create data", err.Error())
	}

	return a.Ok(ctx, "success", newTraveller, nil)
}

func (a *TravellerHandler) Update(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return a.ResponseError(ctx, http.StatusBadRequest, "error validation", "id not found")
	}

	var updateRequest domain.UpdateTravellerRequest
	err = ctx.Bind(&updateRequest)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error())
	}

	err = ctx.Validate(&updateRequest)
	if err != nil {
		return a.ResponseErrorValidation(ctx, err)
	}

	err = a.Service.Update(ctx.Request().Context(), id, updateRequest)
	if err != nil {
		return a.ResponseError(ctx, http.StatusBadRequest, "error update data", err.Error())
	}

	traveller, err := a.Service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		return a.ResponseError(ctx, http.StatusBadRequest, "error get updated data", err.Error())
	}

	response := domain.ToTravellerResponse(traveller)
	return a.Ok(ctx, "success", response, nil)
}

func (a *TravellerHandler) Delete(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return a.ResponseError(ctx, http.StatusBadRequest, "error validation", "id not found")
	}

	err = a.Service.Delete(ctx.Request().Context(), id)
	if err != nil {
		return a.ResponseError(ctx, http.StatusBadRequest, "error delete data", err.Error())
	}

	return a.Ok(ctx, "success", nil, nil)
}
