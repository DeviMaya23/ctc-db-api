package user

import (
	"context"
	"lizobly/ctc-db-api/internal/user/mocks"
	"lizobly/ctc-db-api/pkg/domain"
	pkgJWT "lizobly/ctc-db-api/pkg/jwt"
	"lizobly/ctc-db-api/pkg/logging"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserServiceSuite struct {
	suite.Suite
	userRepo     *mocks.MockUserRepository
	tokenService *pkgJWT.TokenService
	svc          *userService
	logger       *logging.Logger
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(UserServiceSuite))
}

func (s *UserServiceSuite) SetupSuite() {
	s.logger, _ = logging.NewDevelopmentLogger()
	s.tokenService = pkgJWT.NewTokenService("test-secret-key-for-unit-tests", 10*time.Minute, s.logger)
}

func (s *UserServiceSuite) SetupTest() {

	s.userRepo = new(mocks.MockUserRepository)
	s.svc = NewUserService(s.userRepo, s.tokenService, s.logger)
}

func (s *UserServiceSuite) TearDownTest() {
	s.userRepo.AssertExpectations(s.T())
}

func (s *UserServiceSuite) TestUserService_NewService() {

	s.T().Run("success", func(t *testing.T) {
		logger, _ := logging.NewDevelopmentLogger()
		repo := new(mocks.MockUserRepository)
		tokenService := pkgJWT.NewTokenService("test-secret", 10*time.Minute, logger)
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
				s.userRepo.On("GetByUsername", mock.Anything, args.request.Username).Return(want.user, gorm.ErrRecordNotFound).Once()

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
				assert.True(s.T(), domain.IsAuthenticationError(err), "expected AuthenticationError")
				return
			}

			assert.Nil(s.T(), err)

			token, err := jwt.ParseWithClaims(got.Token, &domain.JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
				return []byte("test-secret-key-for-unit-tests"), nil
			})

			if err != nil {
				s.T().Errorf("error parsing token: %v", err)
			}

			if _, ok := token.Claims.(*domain.JWTClaims); !ok || !token.Valid {
				s.T().Error("invalid token claims")
			}

		})
	}
}
