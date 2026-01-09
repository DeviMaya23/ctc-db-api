package user

import (
	"context"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/helpers"
	"lizobly/ctc-db-api/pkg/logging"
	"lizobly/ctc-db-api/user/mocks"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserServiceSuite struct {
	suite.Suite
	userRepo *mocks.MockUserRepository
	svc      *UserService
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, new(UserServiceSuite))
}

func (s *UserServiceSuite) SetupSuite() {
	logger, _ := logging.NewDevelopmentLogger()

	s.userRepo = new(mocks.MockUserRepository)
	s.svc = NewUserService(s.userRepo, logger)

}

func (s *UserServiceSuite) TestUserService_NewService() {

	s.T().Run("success", func(t *testing.T) {
		logger, _ := logging.NewDevelopmentLogger()
		repo := new(mocks.MockUserRepository)
		NewUserService(repo, logger)
	})
}

func (s *UserServiceSuite) TestUserService_Login() {

	bcryptedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), 8)
	type args struct {
		request domain.LoginRequest
	}
	type want struct {
		user domain.User
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
				user: domain.User{
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
			want: want{
				err: domain.ErrUserNotFound,
			},
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
			want: want{
				err: domain.ErrInvalidPassword,
			},
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
				assert.Equal(s.T(), err, tt.want.err)
				return
			}

			assert.Nil(s.T(), err)

			jwtSecretKey := helpers.EnvWithDefault("JWT_SECRET_KEY", "2catnipsforisla")
			token, err := jwt.ParseWithClaims(got.Token, &domain.JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
				return []byte(jwtSecretKey), nil
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
