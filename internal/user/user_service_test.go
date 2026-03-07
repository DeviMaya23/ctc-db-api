package user

import (
	"context"
	"errors"
	"lizobly/ctc-db-api/internal/user/mocks"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/logging"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

type UserServiceSuite struct {
	suite.Suite
	userRepo     *mocks.MockUserRepository
	tokenService *mocks.MockTokenService
	svc          *userService
	logger       *logging.Logger
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(UserServiceSuite))
}

func (s *UserServiceSuite) SetupSuite() {
	s.logger, _ = logging.NewDevelopmentLogger()
}

func (s *UserServiceSuite) SetupTest() {
	s.userRepo = new(mocks.MockUserRepository)
	s.tokenService = mocks.NewMockTokenService(s.T())
	s.svc = NewUserService(s.userRepo, s.tokenService, s.logger)
}

func (s *UserServiceSuite) TearDownTest() {
	s.userRepo.AssertExpectations(s.T())
}

func (s *UserServiceSuite) TestUserService_NewService() {

	s.T().Run("success", func(t *testing.T) {
		logger, _ := logging.NewDevelopmentLogger()
		repo := new(mocks.MockUserRepository)
		tokenService := mocks.NewMockTokenService(s.T())
		NewUserService(repo, tokenService, logger)
	})
}

func (s *UserServiceSuite) TestUserService_Login() {

	bcryptedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), 8)
	type args struct {
		request domain.LoginRequest
	}
	type want struct {
		user *domain.User
		// response  domain.LoginResponse
		err error
	}
	tests := []struct {
		name       string
		args       args
		want       want
		wantErr    bool
		beforeTest func(ctx context.Context, args args, want want)
	}{
		{
			name: "success",
			args: args{request: domain.LoginRequest{
				Username: "isla",
				Password: "password",
			}},
			want: want{
				user: &domain.User{
					Username: "isla",
					Password: string(bcryptedPassword),
				},
			},
			wantErr: false,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.userRepo.On("GetByUsername", mock.Anything, args.request.Username).Return(want.user, want.err).Once()
				s.tokenService.On("GenerateToken", mock.Anything, args.request.Username).Return("valid-token", time.Now().Add(10*time.Minute), nil).Once()
			},
		},
		{
			name: "user not found",
			args: args{request: domain.LoginRequest{
				Username: "isla",
			}},
			want:    want{},
			wantErr: true,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.userRepo.On("GetByUsername", mock.Anything, args.request.Username).Return(want.user, domain.NewNotFoundError("user", args.request.Username, nil)).Once()
			},
		},
		{
			name: "invalid password",
			args: args{request: domain.LoginRequest{
				Username: "isla",
			}},
			want:    want{},
			wantErr: true,
			beforeTest: func(ctx context.Context, args args, want want) {
				s.userRepo.On("GetByUsername", mock.Anything, args.request.Username).Return(want.user, nil).Once()
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.TODO()

			if tt.beforeTest != nil {
				tt.beforeTest(ctx, tt.args, tt.want)
			}

			got, err := s.svc.Login(ctx, tt.args.request)
			if tt.wantErr {
				assert.Error(s.T(), err)
				var ae *domain.AuthenticationError
				assert.True(s.T(), errors.As(err, &ae), "expected AuthenticationError")
				return
			}

			assert.Nil(s.T(), err)
			assert.NotNil(s.T(), got)
			assert.Equal(s.T(), "valid-token", got.Token)

		})
	}
}
