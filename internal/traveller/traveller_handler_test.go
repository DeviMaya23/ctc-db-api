package traveller

import (
	"encoding/json"
	"lizobly/ctc-db-api/internal/traveller/mocks"
	"lizobly/ctc-db-api/pkg/constants"
	"lizobly/ctc-db-api/pkg/controller"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/helpers"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type TravellerHandlerSuite struct {
	suite.Suite

	e                *echo.Echo
	travellerService *mocks.MockTravellerService
	handler          *TravellerHandler
}

func TestTravellerHandlerSuite(t *testing.T) {
	suite.Run(t, new(TravellerHandlerSuite))
}

func (s *TravellerHandlerSuite) SetupTest() {
	s.e = echo.New()
	s.travellerService = new(mocks.MockTravellerService)
	s.handler = NewTravellerHandler(s.e.Group(""), s.travellerService)
}

func (s *TravellerHandlerSuite) TearDownTest() {
	s.travellerService.AssertExpectations(s.T())
}

func (s *TravellerHandlerSuite) TestTravellerHandler_NewHandler() {

	got := NewTravellerHandler(s.e.Group(""), s.travellerService)
	assert.Equal(s.T(), got, s.handler)

}

func (s *TravellerHandlerSuite) TestTravellerHandler_GetByID() {

	type args struct {
		pathID string
	}
	type want struct {
		traveller    domain.Traveller
		responseBody interface{}
		statusCode   int
	}

	traveller := domain.Traveller{}

	tests := []struct {
		name       string
		args       args
		want       want
		beforeTest func(ctx echo.Context, param args, want want)
	}{
		{
			name: "success get traveller",
			args: args{"1"},
			want: want{
				traveller: traveller,
				responseBody: controller.StandardAPIResponse{
					Message: "success",
					Data:    domain.ToTravellerResponse(traveller),
				},
				statusCode: http.StatusOK,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				id, err := strconv.Atoi(ctx.Param("id"))
				assert.Nil(s.T(), err)
				s.travellerService.On("GetByID", ctx.Request().Context(), id).Return(traveller, nil).Once()
			},
		},
		{
			name: "failed invalid id",
			args: args{""},
			want: want{
				traveller: traveller,
				responseBody: controller.StandardAPIResponse{
					Message: "error validation",
					Errors:  "id not found",
				},
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "failed error get traveller",
			args: args{"2"},
			want: want{
				traveller:  traveller,
				statusCode: http.StatusNotFound,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				id, err := strconv.Atoi(ctx.Param("id"))
				assert.Nil(s.T(), err)
				s.travellerService.On("GetByID", ctx.Request().Context(), id).Return(traveller, domain.NewNotFoundError("traveller", id)).Once()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {

			pathParam := map[string]string{"id": tt.args.pathID}
			rec, ctx := helpers.GetHTTPTestRecorder(s.T(), http.MethodGet, "/travellers/1", nil, nil, pathParam)

			if tt.beforeTest != nil {
				tt.beforeTest(ctx, tt.args, tt.want)
			}

			err := s.handler.GetByID(ctx)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.want.statusCode, ctx.Response().Status)

			if tt.want.responseBody != nil {

				wantRespBytes, err := json.Marshal(tt.want.responseBody)
				assert.NoError(s.T(), err)

				assert.Equal(s.T(), string(wantRespBytes), strings.TrimSpace(rec.Body.String()))

			}

		})
	}

}

func (s *TravellerHandlerSuite) TestTravellerHandler_Create() {

	type args struct {
		requestBody interface{}
	}
	type want struct {
		responseBody interface{}
		statusCode   int
	}

	req := domain.CreateTravellerRequest{
		Name:      "Fiore",
		Rarity:    5,
		Influence: "Fame",
		Job:       "Warrior",
	}

	createdTraveller := domain.Traveller{
		CommonModel: domain.CommonModel{ID: 1},
		Name:        "Fiore",
		Rarity:      5,
		InfluenceID: constants.GetInfluenceID("Fame"),
		Influence:   domain.Influence{Name: "Fame"},
		JobID:       constants.GetJobID("Warrior"),
		Job:         domain.Job{Name: "Warrior"},
	}

	tests := []struct {
		name       string
		args       args
		want       want
		beforeTest func(ctx echo.Context, param args, want want)
	}{
		{
			name: "success create traveller",
			args: args{req},
			want: want{
				responseBody: controller.StandardAPIResponse{
					Message: "success",
					Data:    domain.ToTravellerResponse(createdTraveller),
				},
				statusCode: http.StatusCreated,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				s.travellerService.On("Create", ctx.Request().Context(), param.requestBody).Return(int64(1), nil).Once()
				s.travellerService.On("GetByID", ctx.Request().Context(), 1).Return(createdTraveller, nil).Once()
			},
		},
		{
			name: "failed bind",
			args: args{`{"pe": "pe"}`},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "failed validation",
			args: args{domain.CreateTravellerRequest{Name: "Fiore"}},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "failed create",
			args: args{req},
			want: want{
				responseBody: controller.StandardAPIResponse{
					Message: "error create data",
					Errors:  gorm.ErrInvalidDB.Error(),
				},
				statusCode: http.StatusInternalServerError,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				s.travellerService.On("Create", ctx.Request().Context(), param.requestBody).Return(int64(0), gorm.ErrInvalidDB).Once()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			rec, ctx := helpers.GetHTTPTestRecorder(s.T(), http.MethodPost, "/travellers", tt.args.requestBody, nil, nil)

			if tt.beforeTest != nil {
				tt.beforeTest(ctx, tt.args, tt.want)
			}

			errHandler := s.handler.Create(ctx)
			assert.Nil(s.T(), errHandler)
			assert.Equal(s.T(), tt.want.statusCode, ctx.Response().Status)

			if tt.want.responseBody != nil {

				wantRespBytes, err := json.Marshal(tt.want.responseBody)
				assert.NoError(s.T(), err)
				assert.Equal(s.T(), string(wantRespBytes), strings.TrimSpace(rec.Body.String()))

			}

		})
	}

}

func (s *TravellerHandlerSuite) TestTravellerHandler_Update() {

	type args struct {
		pathID      string
		requestBody interface{}
		headers     map[string]string
	}
	type want struct {
		responseBody interface{}
		statusCode   int
	}

	updateRequest := domain.UpdateTravellerRequest{
		Name:      "Fiore",
		Rarity:    6,
		Influence: constants.InfluencePower,
		Job:       constants.JobMerchant,
	}

	updatedTraveller := domain.Traveller{
		CommonModel: domain.CommonModel{ID: 1},
		Name:        "Fiore",
		Rarity:      6,
		InfluenceID: constants.GetInfluenceID(constants.InfluencePower),
		Influence:   domain.Influence{Name: constants.InfluencePower},
		JobID:       constants.GetJobID(constants.JobMerchant),
		Job:         domain.Job{Name: constants.JobMerchant},
	}

	tests := []struct {
		name       string
		args       args
		want       want
		beforeTest func(ctx echo.Context, param args, want want)
	}{
		{
			name: "success update traveller",
			args: args{"1", updateRequest, nil},
			want: want{
				responseBody: controller.StandardAPIResponse{
					Message: "success",
					Data:    domain.ToTravellerResponse(updatedTraveller),
				},
				statusCode: http.StatusOK,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				s.travellerService.On("Update", ctx.Request().Context(), 1, updateRequest).Return(nil).Once()
				s.travellerService.On("GetByID", ctx.Request().Context(), 1).Return(updatedTraveller, nil).Once()
			},
		},
		{
			name: "failed precondition - ETag mismatch",
			args: args{"1", updateRequest, map[string]string{"If-Match": `"9999999999"`}},
			want: want{
				statusCode: http.StatusPreconditionFailed,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				currentTraveller := updatedTraveller
				// Unix timestamp 1704067230 (different from If-Match 9999999999)
				t, _ := time.Parse(time.RFC3339, "2024-01-01T00:20:30Z")
				currentTraveller.UpdatedAt = t
				s.travellerService.On("GetByID", ctx.Request().Context(), 1).Return(currentTraveller, nil).Once()
			},
		},
		{
			name: "failed invalid id",
			args: args{"", domain.UpdateTravellerRequest{}, nil},
			want: want{
				responseBody: controller.StandardAPIResponse{
					Message: "error validation",
					Errors:  "id not found",
				},
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "failed bind request body",
			args: args{"1", `asdf`, nil},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "failed update",
			args: args{"1", updateRequest, nil},
			want: want{
				responseBody: controller.StandardAPIResponse{
					Message: "error update data",
					Errors:  gorm.ErrInvalidDB.Error(),
				},
				statusCode: http.StatusInternalServerError,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				s.travellerService.On("Update", ctx.Request().Context(), 1, updateRequest).Return(gorm.ErrInvalidDB).Once()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {

			pathParam := map[string]string{"id": tt.args.pathID}
			rec, ctx := helpers.GetHTTPTestRecorder(s.T(), http.MethodPut, "/travellers/1", tt.args.requestBody, nil, pathParam)

			// Set headers if provided
			if tt.args.headers != nil {
				for key, value := range tt.args.headers {
					ctx.Request().Header.Set(key, value)
				}
			}

			if tt.beforeTest != nil {
				tt.beforeTest(ctx, tt.args, tt.want)
			}

			err := s.handler.Update(ctx)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.want.statusCode, ctx.Response().Status)

			if tt.want.responseBody != nil {

				wantRespBytes, err := json.Marshal(tt.want.responseBody)
				assert.NoError(s.T(), err)

				assert.Equal(s.T(), string(wantRespBytes), strings.TrimSpace(rec.Body.String()))

			}

		})
	}

}

func (s *TravellerHandlerSuite) TestTravellerHandler_Delete() {

	type args struct {
		pathID string
	}
	type want struct {
		responseBody interface{}
		statusCode   int
	}

	tests := []struct {
		name       string
		args       args
		want       want
		beforeTest func(ctx echo.Context, param args, want want)
	}{
		{
			name: "success delete traveller",
			args: args{"1"},
			want: want{
				statusCode: http.StatusNoContent,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				s.travellerService.On("Delete", ctx.Request().Context(), 1).Return(nil).Once()
			},
		},
		{
			name: "failed invalid id",
			args: args{""},
			want: want{
				responseBody: controller.StandardAPIResponse{
					Message: "error validation",
					Errors:  "id not found",
				},
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "failed delete",
			args: args{"1"},
			want: want{
				responseBody: controller.StandardAPIResponse{
					Message: "error delete data",
					Errors:  gorm.ErrInvalidDB.Error(),
				},
				statusCode: http.StatusInternalServerError,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				s.travellerService.On("Delete", ctx.Request().Context(), 1).Return(gorm.ErrInvalidDB)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {

			pathParam := map[string]string{"id": tt.args.pathID}
			rec, ctx := helpers.GetHTTPTestRecorder(s.T(), http.MethodDelete, "/travellers/1", nil, nil, pathParam)

			if tt.beforeTest != nil {
				tt.beforeTest(ctx, tt.args, tt.want)
			}

			err := s.handler.Delete(ctx)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.want.statusCode, ctx.Response().Status)

			if tt.want.responseBody != nil {

				wantRespBytes, err := json.Marshal(tt.want.responseBody)
				assert.NoError(s.T(), err)

				assert.Equal(s.T(), string(wantRespBytes), strings.TrimSpace(rec.Body.String()))

			}

		})
	}

}

func (s *TravellerHandlerSuite) TestTravellerHandler_GetList() {

	type args struct {
		queryParams map[string]string
	}
	type want struct {
		responseBody interface{}
		statusCode   int
	}

	tests := []struct {
		name       string
		args       args
		want       want
		beforeTest func(ctx echo.Context, param args, want want)
	}{
		{
			name: "success get list without filters",
			args: args{
				queryParams: map[string]string{},
			},
			want: want{
				statusCode: http.StatusOK,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				filter := domain.ListTravellerRequest{}
				response := helpers.PaginatedResponse[domain.TravellerListItemResponse]{
					Data:       []domain.TravellerListItemResponse{},
					Page:       1,
					PageSize:   10,
					Total:      0,
					TotalPages: 0,
				}
				s.travellerService.On("GetList", mock.Anything, filter, mock.MatchedBy(func(p helpers.PaginationParams) bool {
					return true
				})).Return(response, nil).Once()
			},
		},
		{
			name: "success get list with name filter",
			args: args{
				queryParams: map[string]string{"name": "Fiore"},
			},
			want: want{
				statusCode: http.StatusOK,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				filter := domain.ListTravellerRequest{Name: "Fiore"}
				response := helpers.PaginatedResponse[domain.TravellerListItemResponse]{
					Data: []domain.TravellerListItemResponse{
						{Name: "Fiore", Rarity: 5},
					},
					Page:       1,
					PageSize:   10,
					Total:      1,
					TotalPages: 1,
				}
				s.travellerService.On("GetList", mock.Anything, filter, mock.MatchedBy(func(p helpers.PaginationParams) bool {
					return true
				})).Return(response, nil).Once()
			},
		},
		{
			name: "success get list with multiple filters",
			args: args{
				queryParams: map[string]string{
					"name":      "Fiore",
					"influence": "Fame",
					"job":       "Warrior",
					"page":      "2",
					"page_size": "20",
				},
			},
			want: want{
				statusCode: http.StatusOK,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				filter := domain.ListTravellerRequest{
					Name:      "Fiore",
					Influence: "Fame",
					Job:       "Warrior",
				}
				paginationParams := helpers.PaginationParams{
					Page:     2,
					PageSize: 20,
				}
				response := helpers.PaginatedResponse[domain.TravellerListItemResponse]{
					Data: []domain.TravellerListItemResponse{
						{Name: "Fiore", Rarity: 5},
					},
					Page:       2,
					PageSize:   20,
					Total:      1,
					TotalPages: 1,
				}
				s.travellerService.On("GetList", mock.Anything, filter, paginationParams).Return(response, nil).Once()
			},
		},
		{
			name: "failed service error",
			args: args{
				queryParams: map[string]string{},
			},
			want: want{
				responseBody: controller.StandardAPIResponse{
					Message: "error get data",
					Errors:  gorm.ErrInvalidDB.Error(),
				},
				statusCode: http.StatusInternalServerError,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				filter := domain.ListTravellerRequest{}
				s.travellerService.On("GetList", mock.Anything, filter, mock.MatchedBy(func(p helpers.PaginationParams) bool {
					return true
				})).Return(helpers.PaginatedResponse[domain.TravellerListItemResponse]{}, gorm.ErrInvalidDB).Once()
			},
		},
		{
			name: "failed filter validation",
			args: args{
				queryParams: map[string]string{"influence": "InvalidInfluence"},
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Convert queryParams map to url.Values
			queryValues := url.Values{}
			for k, v := range tt.args.queryParams {
				queryValues.Set(k, v)
			}
			rec, ctx := helpers.GetHTTPTestRecorder(s.T(), http.MethodGet, "/travellers", nil, queryValues, nil)

			if tt.beforeTest != nil {
				tt.beforeTest(ctx, tt.args, tt.want)
			}

			err := s.handler.GetList(ctx)
			assert.Nil(s.T(), err)
			assert.Equal(s.T(), tt.want.statusCode, ctx.Response().Status)

			if tt.want.responseBody != nil {
				wantRespBytes, err := json.Marshal(tt.want.responseBody)
				assert.NoError(s.T(), err)
				assert.Equal(s.T(), string(wantRespBytes), strings.TrimSpace(rec.Body.String()))
			}
		})
	}
}
