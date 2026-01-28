package jwt

import (
	"context"
	"fmt"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/logging"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// TokenService handles JWT token generation and validation
type TokenService struct {
	secretKey []byte
	timeout   time.Duration
	logger    *logging.Logger
}

// NewTokenService creates a new token service
func NewTokenService(secretKey string, timeout time.Duration, logger *logging.Logger) *TokenService {
	return &TokenService{
		secretKey: []byte(secretKey),
		timeout:   timeout,
		logger:    logger.Named("jwt.token_service"),
	}
}

// GenerateToken creates a new JWT token for the given username
func (s *TokenService) GenerateToken(ctx context.Context, username string) (token string, expiresAt time.Time, err error) {
	expiresAt = time.Now().Add(s.timeout)

	claims := domain.JWTClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = jwtToken.SignedString(s.secretKey)
	if err != nil {
		s.logger.WithContext(ctx).Error("failed to sign JWT token",
			zap.String("user.username", username),
			zap.Error(err),
		)
		return "", time.Time{}, fmt.Errorf("generate token: %w", err)
	}

	s.logger.WithContext(ctx).Debug("JWT token generated",
		zap.String("user.username", username),
		zap.Time("token.expiration", expiresAt),
	)

	return token, expiresAt, nil
}
