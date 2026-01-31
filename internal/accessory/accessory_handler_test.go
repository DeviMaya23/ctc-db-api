package accessory

import (
	"encoding/json"
	"lizobly/ctc-db-api/internal/accessory/mocks"
	"lizobly/ctc-db-api/pkg/controller"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/helpers"
	"lizobly/ctc-db-api/pkg/logging"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type AccessoryHandlerSuite struct {
	suite.Suite

	e                *echo.Echo
	accessoryService *mocks.MockAccessoryService
	handler          *AccessoryHandler
}

func TestAccessoryHandlerSuite(t *testing.T) {
	suite.Run(t, new(AccessoryHandlerSuite))
}

func (s *AccessoryHandlerSuite) SetupTest() {
	s.e = echo.New()
	s.accessoryService = new(mocks.MockAccessoryService)
	testLogger, _ := logging.NewDevelopmentLogger()
	s.handler = NewAccessoryHandler(s.e.Group(""), s.accessoryService, testLogger)
}

func (s *AccessoryHandlerSuite) TearDownTest() {
	s.accessoryService.AssertExpectations(s.T())
}

func (s *AccessoryHandlerSuite) TestAccessoryHandler_NewHandler() {
	testLogger, _ := logging.NewDevelopmentLogger()
	got := NewAccessoryHandler(s.e.Group(""), s.accessoryService, testLogger)
	assert.Equal(s.T(), s.accessoryService, got.Service)
	assert.NotNil(s.T(), got.logger)
}

func (s *AccessoryHandlerSuite) TestAccessoryHandler_GetList() {

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
				filter := domain.ListAccessoryRequest{}
				response := helpers.PaginatedResponse[domain.AccessoryListItemResponse]{
					Data:       []domain.AccessoryListItemResponse{},
					Page:       1,
					PageSize:   10,
					Total:      0,
					TotalPages: 0,
				}
				s.accessoryService.On("GetList", mock.Anything, filter, mock.MatchedBy(func(p helpers.PaginationParams) bool {
					return true
				})).Return(response, nil).Once()
			},
		},
		{
			name: "success get list with owner and effect filters",
			args: args{
				queryParams: map[string]string{
					"owner":  "Fiore",
					"effect": "Elemental",
				},
			},
			want: want{
				statusCode: http.StatusOK,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				filter := domain.ListAccessoryRequest{
					Owner:  "Fiore",
					Effect: "Elemental",
				}
				response := helpers.PaginatedResponse[domain.AccessoryListItemResponse]{
					Data: []domain.AccessoryListItemResponse{
						{
							Name:   "Crown of Wisdom",
							HP:     150,
							SP:     80,
							PAtk:   45,
							PDef:   30,
							EAtk:   60,
							EDef:   25,
							Spd:    12,
							Crit:   8,
							Effect: "Increases elemental damage by 15%",
							Owner:  "Fiore",
						},
					},
					Page:       1,
					PageSize:   10,
					Total:      1,
					TotalPages: 1,
				}
				s.accessoryService.On("GetList", mock.Anything, filter, mock.MatchedBy(func(p helpers.PaginationParams) bool {
					return true
				})).Return(response, nil).Once()
			},
		},
		{
			name: "success get list with ordering and pagination",
			args: args{
				queryParams: map[string]string{
					"order_by":  "hp",
					"order_dir": "desc",
					"page":      "2",
					"page_size": "20",
				},
			},
			want: want{
				statusCode: http.StatusOK,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				filter := domain.ListAccessoryRequest{
					OrderBy:  "hp",
					OrderDir: "desc",
				}
				paginationParams := helpers.PaginationParams{
					Page:     2,
					PageSize: 20,
				}
				response := helpers.PaginatedResponse[domain.AccessoryListItemResponse]{
					Data: []domain.AccessoryListItemResponse{
						{
							Name:   "Ring of Power",
							HP:     200,
							SP:     100,
							PAtk:   80,
							PDef:   50,
							EAtk:   70,
							EDef:   40,
							Spd:    15,
							Crit:   10,
							Effect: "Increases physical damage",
							Owner:  "Noctis",
						},
					},
					Page:       2,
					PageSize:   20,
					Total:      1,
					TotalPages: 1,
				}
				s.accessoryService.On("GetList", mock.Anything, filter, paginationParams).Return(response, nil).Once()
			},
		},
		{
			name: "failed invalid order_by validation",
			args: args{
				queryParams: map[string]string{
					"order_by": "invalid_field",
				},
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				// No mock setup - validation should fail before service call
			},
		},
		{
			name: "failed invalid order_dir validation",
			args: args{
				queryParams: map[string]string{
					"order_by":  "hp",
					"order_dir": "invalid_direction",
				},
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				// No mock setup - validation should fail before service call
			},
		},
		{
			name: "failed service error",
			args: args{
				queryParams: map[string]string{},
			},
			want: want{
				responseBody: controller.ErrorResponse{
					Message: "internal server error",
				},
				statusCode: http.StatusInternalServerError,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				filter := domain.ListAccessoryRequest{}
				s.accessoryService.On("GetList", mock.Anything, filter, mock.MatchedBy(func(p helpers.PaginationParams) bool {
					return true
				})).Return(helpers.PaginatedResponse[domain.AccessoryListItemResponse]{}, gorm.ErrInvalidDB).Once()
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
			rec, ctx := helpers.GetHTTPTestRecorder(s.T(), http.MethodGet, "/accessories", nil, queryValues, nil)

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
