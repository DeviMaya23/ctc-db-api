package user

import (
	"encoding/json"
	"lizobly/ctc-db-api/internal/user/mocks"
	"lizobly/ctc-db-api/pkg/controller"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/helpers"
	"net/http"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type UserHandlerSuite struct {
	suite.Suite

	e           *echo.Echo
	userService *mocks.MockUserService
	handler     *UserHandler
}

func TestUserHandlerSuite(t *testing.T) {
	suite.Run(t, new(UserHandlerSuite))
}

func (s *UserHandlerSuite) SetupTest() {
	s.e = echo.New()
	s.userService = new(mocks.MockUserService)
	s.handler = NewUserHandler(s.e.Group(""), s.userService)
}

func (s *UserHandlerSuite) TearDownTest() {
	s.userService.AssertExpectations(s.T())
}

func (s *UserHandlerSuite) TestUserHandler_NewHandler() {

	got := NewUserHandler(s.e.Group(""), s.userService)
	assert.Equal(s.T(), got, s.handler)

}

func (s *UserHandlerSuite) TestUserHandler_Login() {

	type args struct {
		requestBody interface{}
	}
	type want struct {
		loginResponse domain.LoginResponse
		responseBody  interface{}
		statusCode    int
	}

	login := domain.LoginRequest{
		Username: "user",
		Password: "pw",
	}
	resp := domain.LoginResponse{
		Username: "isla",
		Token:    "1234",
	}

	tests := []struct {
		name       string
		args       args
		want       want
		beforeTest func(ctx echo.Context, param args, want want)
	}{
		{
			name: "success get user",
			args: args{login},
			want: want{
				loginResponse: resp,
				responseBody: controller.DataResponse[domain.LoginResponse]{
					Data: resp,
				},
				statusCode: http.StatusOK,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				s.userService.On("Login", mock.Anything, param.requestBody).Return(want.loginResponse, nil).Once()

			},
		},
		{
			name: "failed bind",
			args: args{`asdf`},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "failed validation",
			args: args{domain.LoginRequest{}},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "failed authentication - invalid credentials",
			args: args{login},
			want: want{
				loginResponse: resp,
				responseBody: controller.ErrorResponse{
					Message: "invalid credentials",
				},
				statusCode: http.StatusUnauthorized,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				s.userService.On("Login", mock.Anything, param.requestBody).Return(want.loginResponse, domain.NewAuthenticationError("invalid credentials")).Once()

			},
		},
		{
			name: "failed internal server error",
			args: args{login},
			want: want{
				loginResponse: resp,
				responseBody: controller.ErrorResponse{
					Message: "internal server error",
				},
				statusCode: http.StatusInternalServerError,
			},
			beforeTest: func(ctx echo.Context, param args, want want) {
				s.userService.On("Login", mock.Anything, param.requestBody).Return(want.loginResponse, gorm.ErrCheckConstraintViolated).Once()

			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {

			rec, ctx := helpers.GetHTTPTestRecorder(s.T(), http.MethodPost, "/login", tt.args.requestBody, nil, nil)

			if tt.beforeTest != nil {
				tt.beforeTest(ctx, tt.args, tt.want)
			}

			err := s.handler.Login(ctx)
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
