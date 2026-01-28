package traveller

import (
	"context"
	"lizobly/ctc-db-api/pkg/constants"
	"lizobly/ctc-db-api/pkg/controller"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/helpers"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type TravellerService interface {
	GetByID(ctx context.Context, id int) (res *domain.Traveller, err error)
	GetList(ctx context.Context, filter domain.ListTravellerRequest, params helpers.PaginationParams) (res helpers.PaginatedResponse[domain.TravellerListItemResponse], err error)
	Create(ctx context.Context, input domain.CreateTravellerRequest) (id int64, err error)
	Update(ctx context.Context, id int, input domain.UpdateTravellerRequest) (err error)
	Delete(ctx context.Context, id int) (err error)
}

type TravellerHandler struct {
	Service TravellerService
}

func NewTravellerHandler(e *echo.Group, svc TravellerService) *TravellerHandler {
	handler := &TravellerHandler{
		Service: svc,
	}
	group := e.Group("/travellers")

	group.GET("", handler.GetList)
	group.GET("/:id", handler.GetByID)
	group.POST("", handler.Create)
	group.PUT("/:id", handler.Update)
	group.DELETE("/:id", handler.Delete)

	return handler
}

// GetList godoc
//
//	@Summary		Get list
//	@Description	get traveller list with optional filters and pagination
//	@Tags			accounts
//	@Accept			json
//	@Produce		json
//	@Param			name		query	string	false	"Filter by name (case insensitive)"
//	@Param			influence	query	string	false	"Filter by influence name"
//	@Param			job			query	string	false	"Filter by job name"
//	@Param			page		query	int		false	"Page number (default 1)"
//	@Param			page_size	query	int		false	"Page size (default 10, max 100)"
//	@Success		200	{object}	helpers.PaginatedResponse[domain.TravellerListItemResponse]
//	@Failure		400	{object}	StandardAPIResponse
//	@Failure		500	{object}	StandardAPIResponse
//	@Router			/travellers [get]
func (h *TravellerHandler) GetList(ctx echo.Context) error {
	var filter domain.ListTravellerRequest
	err := ctx.Bind(&filter)
	if err != nil {
		return controller.ResponseError(ctx, http.StatusBadRequest, "invalid request body")
	}

	err = ctx.Validate(&filter)
	if err != nil {
		return controller.ResponseErrorValidation(ctx, err)
	}

	var params helpers.PaginationParams
	err = ctx.Bind(&params)
	if err != nil {
		return controller.ResponseError(ctx, http.StatusBadRequest, "invalid pagination parameters")
	}

	result, err := h.Service.GetList(ctx.Request().Context(), filter, params)
	if err != nil {
		return controller.HandleServiceError(ctx, err, "get data")
	}

	// Set cache headers for list responses
	helpers.SetListCacheHeaders(ctx)

	return controller.Ok(ctx, result)
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
func (h *TravellerHandler) GetByID(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return controller.ResponseError(ctx, http.StatusBadRequest, "invalid id parameter")
	}

	traveller, err := h.Service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		return controller.HandleServiceError(ctx, err, "get data")
	}

	// Set cache headers and check if client has valid cached version
	if helpers.SetCacheHeaders(ctx, traveller.ETag(), traveller.LastModified(), constants.CacheMaxAgeResource) {
		return helpers.RespondNotModified(ctx)
	}

	response := domain.ToTravellerResponse(traveller)
	return controller.Ok(ctx, response)
}

func (h *TravellerHandler) Create(ctx echo.Context) error {

	var newTraveller domain.CreateTravellerRequest
	err := ctx.Bind(&newTraveller)
	if err != nil {
		return controller.ResponseError(ctx, http.StatusBadRequest, "invalid request body")
	}

	err = ctx.Validate(&newTraveller)
	if err != nil {
		return controller.ResponseErrorValidation(ctx, err)
	}

	id, err := h.Service.Create(ctx.Request().Context(), newTraveller)
	if err != nil {
		return controller.HandleServiceError(ctx, err, "create data")
	}

	traveller, err := h.Service.GetByID(ctx.Request().Context(), int(id))
	if err != nil {
		return controller.HandleServiceError(ctx, err, "get created data")
	}

	// Set ETag and Last-Modified for created resource
	ctx.Response().Header().Set("ETag", traveller.ETag())
	ctx.Response().Header().Set("Last-Modified", traveller.LastModified())

	location := "/api/v1/travellers/" + strconv.FormatInt(id, 10)
	response := domain.ToTravellerResponse(traveller)
	return controller.Created(ctx, response, location)
}

func (h *TravellerHandler) Update(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return controller.ResponseError(ctx, http.StatusBadRequest, "invalid id parameter")
	}

	// Check for optimistic locking with If-Match header
	if ctx.Request().Header.Get("If-Match") != "" {
		// Get current state to verify ETag
		currentTraveller, err := h.Service.GetByID(ctx.Request().Context(), id)
		if err != nil {
			return controller.HandleServiceError(ctx, err, "get data")
		}

		// Prevent lost updates - resource was modified
		if !helpers.CheckETagMatch(ctx, currentTraveller.ETag()) {
			return helpers.RespondPreconditionFailed(ctx)
		}
	}

	var updateRequest domain.UpdateTravellerRequest
	err = ctx.Bind(&updateRequest)
	if err != nil {
		return controller.ResponseError(ctx, http.StatusBadRequest, "invalid request body")
	}

	err = ctx.Validate(&updateRequest)
	if err != nil {
		return controller.ResponseErrorValidation(ctx, err)
	}

	err = h.Service.Update(ctx.Request().Context(), id, updateRequest)
	if err != nil {
		return controller.HandleServiceError(ctx, err, "update data")
	}

	traveller, err := h.Service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		return controller.HandleServiceError(ctx, err, "get updated data")
	}

	// Set new ETag and Last-Modified for updated resource
	ctx.Response().Header().Set("ETag", traveller.ETag())
	ctx.Response().Header().Set("Last-Modified", traveller.LastModified())

	response := domain.ToTravellerResponse(traveller)
	return controller.Ok(ctx, response)
}

func (h *TravellerHandler) Delete(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return controller.ResponseError(ctx, http.StatusBadRequest, "invalid id parameter")
	}

	err = h.Service.Delete(ctx.Request().Context(), id)
	if err != nil {
		return controller.HandleServiceError(ctx, err, "delete data")
	}

	return controller.NoContent(ctx)
}
