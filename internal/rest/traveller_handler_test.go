package rest

import (
	"encoding/json"
	"lizobly/ctc-db-api/internal/rest/mocks"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/helpers"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
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

func (s *TravellerHandlerSuite) SetupSuite() {
	s.e = echo.New()
	s.travellerService = new(mocks.MockTravellerService)
	s.handler = NewTravellerHandler(s.e.Group(""), s.travellerService)

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
				responseBody: StandardAPIResponse{
					Message: "success",
					Data:    traveller,
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
				responseBody: StandardAPIResponse{
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
				traveller: traveller,
				responseBody: StandardAPIResponse{
					Message: "error get data",
					Errors:  gorm.ErrRecordNotFound.Error(),
				},
				statusCode: http.StatusBadRequest,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				id, err := strconv.Atoi(ctx.Param("id"))
				assert.Nil(s.T(), err)
				s.travellerService.On("GetByID", ctx.Request().Context(), id).Return(traveller, gorm.ErrRecordNotFound).Once()
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
				responseBody: StandardAPIResponse{
					Message: "success",
					Data:    req,
				},
				statusCode: http.StatusOK,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				s.travellerService.On("Create", ctx.Request().Context(), param.requestBody).Return(nil).Once()
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
				responseBody: StandardAPIResponse{
					Message: "error create data",
					Errors:  gorm.ErrInvalidDB.Error(),
				},
				statusCode: http.StatusBadRequest,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				s.travellerService.On("Create", ctx.Request().Context(), param.requestBody).Return(gorm.ErrInvalidDB).Once()
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
	}
	type want struct {
		responseBody interface{}
		statusCode   int
	}

	traveller := domain.Traveller{
		Name:   "Fiore",
		Rarity: 6,
	}
	result := domain.Traveller{
		Name:   "Fiore",
		Rarity: 6,
		CommonModel: domain.CommonModel{
			ID: int64(1),
		},
	}

	tests := []struct {
		name       string
		args       args
		want       want
		beforeTest func(ctx echo.Context, param args, want want)
	}{
		{
			name: "success update traveller",
			args: args{"1", traveller},
			want: want{
				responseBody: StandardAPIResponse{
					Message: "success",
					Data:    result,
				},
				statusCode: http.StatusOK,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				s.travellerService.On("Update", ctx.Request().Context(), &result).Return(nil).Once()
			},
		},
		{
			name: "failed invalid id",
			args: args{"", domain.Traveller{}},
			want: want{
				responseBody: StandardAPIResponse{
					Message: "error validation",
					Errors:  "id not found",
				},
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "failed bind request body",
			args: args{"1", `asdf`},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "failed update",
			args: args{"1", traveller},
			want: want{
				responseBody: StandardAPIResponse{
					Message: "error update data",
					Errors:  gorm.ErrInvalidDB.Error(),
				},
				statusCode: http.StatusBadRequest,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				s.travellerService.On("Update", ctx.Request().Context(), &result).Return(gorm.ErrInvalidDB)
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {

			pathParam := map[string]string{"id": tt.args.pathID}
			rec, ctx := helpers.GetHTTPTestRecorder(s.T(), http.MethodPut, "/travellers/1", tt.args.requestBody, nil, pathParam)

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
				responseBody: StandardAPIResponse{
					Message: "success",
				},
				statusCode: http.StatusOK,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				s.travellerService.On("Delete", ctx.Request().Context(), 1).Return(nil).Once()
			},
		},
		{
			name: "failed invalid id",
			args: args{""},
			want: want{
				responseBody: StandardAPIResponse{
					Message: "error validation",
					Errors:  "id not found",
				},
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "failed update",
			args: args{"1"},
			want: want{
				responseBody: StandardAPIResponse{
					Message: "error delete data",
					Errors:  gorm.ErrInvalidDB.Error(),
				},
				statusCode: http.StatusBadRequest,
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
