package jwt

import (
	"context"
	"fmt"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/logging"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
		return "", time.Time{}, fmt.Errorf("generate token: %w", err)
	}

	return token, expiresAt, nil
}
