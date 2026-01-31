package traveller

import (
	"context"
	"lizobly/ctc-db-api/pkg/constants"
	"lizobly/ctc-db-api/pkg/controller"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/helpers"
	"lizobly/ctc-db-api/pkg/logging"
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
	logger  *logging.Logger
}

func NewTravellerHandler(e *echo.Group, svc TravellerService, logger *logging.Logger) *TravellerHandler {
	handler := &TravellerHandler{
		Service: svc,
		logger:  logger.Named("handler.traveller"),
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
//	@Tags			travellers
//	@Accept			json
//	@Produce		json
//	@Param			name		query	string	false	"Filter by name (case insensitive)"
//	@Param			influence	query	string	false	"Filter by influence name"
//	@Param			job			query	string	false	"Filter by job name"
//	@Param			page		query	int		false	"Page number (default 1)"
//	@Param			page_size	query	int		false	"Page size (default 10, max 100)"
//	@Success		200	{object}	helpers.PaginatedResponse[domain.TravellerListItemResponse]
//	@Failure		400	{object}	controller.ErrorResponse
//	@Failure		500	{object}	controller.ErrorResponse
//	@Router			/travellers [get]
//	@Security		BearerAuth
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
		return controller.HandleServiceError(ctx, err, "get traveller list", h.logger)
	}

	// Set cache headers for list responses
	helpers.SetListCacheHeaders(ctx)

	return controller.Ok(ctx, result)
}

// GetByID godoc
//
//	@Summary		Get by ID
//	@Description	get traveller information by ID
//	@Tags			travellers
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Traveller ID"
//	@Success		200	{object}	domain.TravellerResponse
//	@Header			200	{string}	ETag	"Entity tag for caching"
//	@Header			200	{string}	Last-Modified	"Last modified timestamp"
//	@Failure		400	{object}	controller.ErrorResponse
//	@Failure		404	{object}	controller.ErrorResponse
//	@Failure		500	{object}	controller.ErrorResponse
//	@Router			/travellers/{id} [get]
//	@Security		BearerAuth
func (h *TravellerHandler) GetByID(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return controller.ResponseError(ctx, http.StatusBadRequest, "invalid id parameter")
	}

	traveller, err := h.Service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		return controller.HandleServiceError(ctx, err, "get traveller by id", h.logger)
	}

	// Set cache headers and check if client has valid cached version
	if helpers.SetCacheHeaders(ctx, traveller.ETag(), traveller.LastModified(), constants.CacheMaxAgeResource) {
		return helpers.RespondNotModified(ctx)
	}

	response := domain.ToTravellerResponse(traveller)
	return controller.Ok(ctx, response)
}

// Create godoc
//
//	@Summary		Create traveller
//	@Description	create a new traveller with optional accessory
//	@Tags			travellers
//	@Accept			json
//	@Produce		json
//	@Param			body	body		domain.CreateTravellerRequest	true	"Traveller data"
//	@Success		201	{object}	domain.TravellerResponse
//	@Header			201	{string}	Location	"URI of the created resource"
//	@Header			201	{string}	ETag	"Entity tag for caching"
//	@Header			201	{string}	Last-Modified	"Last modified timestamp"
//	@Failure		400	{object}	controller.ErrorResponse
//	@Failure		409	{object}	controller.ErrorResponse
//	@Failure		500	{object}	controller.ErrorResponse
//	@Router			/travellers [post]
//	@Security		BearerAuth
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
		return controller.HandleServiceError(ctx, err, "create traveller", h.logger)
	}

	traveller, err := h.Service.GetByID(ctx.Request().Context(), int(id))
	if err != nil {
		return controller.HandleServiceError(ctx, err, "get created traveller", h.logger)
	}

	// Set ETag and Last-Modified for created resource
	ctx.Response().Header().Set("ETag", traveller.ETag())
	ctx.Response().Header().Set("Last-Modified", traveller.LastModified())

	location := "/api/v1/travellers/" + strconv.FormatInt(id, 10)
	response := domain.ToTravellerResponse(traveller)
	return controller.Created(ctx, response, location)
}

// Update godoc
//
//	@Summary		Update traveller
//	@Description	update an existing traveller by ID with optimistic locking support via If-Match header
//	@Tags			travellers
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Traveller ID"
//	@Param			body	body		domain.UpdateTravellerRequest	true	"Updated traveller data"
//	@Param			If-Match	header	string	false	"ETag for optimistic locking"
//	@Success		200	{object}	domain.TravellerResponse
//	@Header			200	{string}	ETag	"Updated entity tag"
//	@Header			200	{string}	Last-Modified	"Updated timestamp"
//	@Failure		400	{object}	controller.ErrorResponse
//	@Failure		404	{object}	controller.ErrorResponse
//	@Failure		412	{object}	controller.ErrorResponse	"Precondition Failed - resource was modified"
//	@Failure		500	{object}	controller.ErrorResponse
//	@Router			/travellers/{id} [put]
//	@Security		BearerAuth
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
			return controller.HandleServiceError(ctx, err, "get traveller for etag check", h.logger)
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
		return controller.HandleServiceError(ctx, err, "update traveller", h.logger)
	}

	traveller, err := h.Service.GetByID(ctx.Request().Context(), id)
	if err != nil {
		return controller.HandleServiceError(ctx, err, "get updated traveller", h.logger)
	}

	// Set new ETag and Last-Modified for updated resource
	ctx.Response().Header().Set("ETag", traveller.ETag())
	ctx.Response().Header().Set("Last-Modified", traveller.LastModified())

	response := domain.ToTravellerResponse(traveller)
	return controller.Ok(ctx, response)
}

// Delete godoc
//
//	@Summary		Delete traveller
//	@Description	soft delete a traveller by ID
//	@Tags			travellers
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Traveller ID"
//	@Success		204	"No Content"
//	@Failure		400	{object}	controller.ErrorResponse
//	@Failure		404	{object}	controller.ErrorResponse
//	@Failure		500	{object}	controller.ErrorResponse
//	@Router			/travellers/{id} [delete]
//	@Security		BearerAuth
func (h *TravellerHandler) Delete(ctx echo.Context) error {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return controller.ResponseError(ctx, http.StatusBadRequest, "invalid id parameter")
	}

	err = h.Service.Delete(ctx.Request().Context(), id)
	if err != nil {
		return controller.HandleServiceError(ctx, err, "delete traveller", h.logger)
	}

	return controller.NoContent(ctx)
}
