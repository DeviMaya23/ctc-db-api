package user

import (
	"context"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/logging"
	"lizobly/ctc-db-api/pkg/telemetry"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (result domain.User, err error)
}

type TokenService interface {
	GenerateToken(ctx context.Context, username string) (token string, expiresAt time.Time, err error)
}

type UserService struct {
	userRepo     UserRepository
	tokenService TokenService
	logger       *logging.Logger
}

func NewUserService(u UserRepository, ts TokenService, logger *logging.Logger) *UserService {
	return &UserService{
		userRepo:     u,
		tokenService: ts,
		logger:       logger.Named("service.user"),
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
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	dummyHash := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy" // bcrypt hash of "dummy"
	passwordHash := dummyHash
	userFound := true

	if err != nil {
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

	// Generate JWT token
	token, expiresAt, err := s.tokenService.GenerateToken(ctx, user.Username)
	if err != nil {
		return res, err
	}

	res.Username = req.Username
	res.Token = token

	s.logger.WithContext(ctx).Info("login successful",
		zap.String("user.username", req.Username),
		zap.Time("token.expiration", expiresAt),
	)

	return
}
