package user

import (
	"context"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/helpers"
	"lizobly/ctc-db-api/pkg/logging"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (result domain.User, err error)
}

type UserService struct {
	userRepo UserRepository
	logger   *logging.Logger
}

func NewUserService(u UserRepository, logger *logging.Logger) *UserService {
	return &UserService{
		userRepo: u,
		logger:   logger.Named("service.user"),
	}
}

func (s UserService) Login(ctx context.Context, req domain.LoginRequest) (res domain.LoginResponse, err error) {
	s.logger.WithContext(ctx).Info("login attempt",
		zap.String("user.username", req.Username),
	)

	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		s.logger.WithContext(ctx).Warn("user not found",
			zap.String("user.username", req.Username),
		)
		err = domain.ErrUserNotFound
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		s.logger.WithContext(ctx).Warn("invalid password",
			zap.String("user.username", req.Username),
		)
		err = domain.ErrInvalidPassword
		return
	}

	jwtSecretKey := helpers.EnvWithDefault("JWT_SECRET_KEY", "2catnipsforisla")
	jwtTimeoutStr := helpers.EnvWithDefault("JWT_TIMEOUT", "10m")
	jwtTimeout, _ := time.ParseDuration(jwtTimeoutStr)

	exp := time.Now().Add(jwtTimeout)
	claims := domain.JWTClaims{
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(jwtSecretKey))
	if err != nil {
		s.logger.WithContext(ctx).Error("failed to generate JWT token",
			zap.String("user.username", req.Username),
			zap.String("error.message", err.Error()),
		)
		return
	}

	res.Username = req.Username
	res.Token = t

	s.logger.WithContext(ctx).Info("login successful",
		zap.String("user.username", req.Username),
		zap.Time("token.expiration", exp),
	)

	return
}
