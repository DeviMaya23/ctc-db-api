package user

import (
	"context"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/helpers"
	"lizobly/ctc-db-api/pkg/logging"
	"lizobly/ctc-db-api/pkg/telemetry"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel/attribute"
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
	// Start service span
	ctx, span := telemetry.StartServiceSpan(ctx, "service.user", "UserService.Login",
		attribute.String("user.username", req.Username),
	)
	defer telemetry.EndSpanWithError(span, err)

	s.logger.WithContext(ctx).Info("login attempt",
		zap.String("user.username", req.Username),
	)

	// Always run bcrypt comparison to prevent timing-based username enumeration
	// Use a dummy hash if user doesn't exist
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	dummyHash := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy" // bcrypt hash of "dummy"
	passwordHash := dummyHash
	userFound := true

	if err != nil {
		s.logger.WithContext(ctx).Warn("authentication failed",
			zap.String("user.username", req.Username),
		)
		userFound = false
	} else {
		passwordHash = user.Password
	}

	// Always run bcrypt comparison regardless of whether user exists
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil || !userFound {
		s.logger.WithContext(ctx).Warn("authentication failed",
			zap.String("user.username", req.Username),
		)
		err = domain.NewAuthenticationError("invalid credentials")
		return
	}

	jwtSecretKey := os.Getenv("JWT_SECRET_KEY")
	if jwtSecretKey == "" {
		s.logger.WithContext(ctx).Error("JWT_SECRET_KEY not configured")
		err = domain.NewInternalError("authentication configuration error")
		return
	}
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
