package jwt

import (
	"context"
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/logging"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TokenServiceSuite struct {
	suite.Suite
	logger    *logging.Logger
	service   *TokenService
	secretKey string
	timeout   time.Duration
}

func TestTokenServiceSuite(t *testing.T) {
	suite.Run(t, new(TokenServiceSuite))
}

func (s *TokenServiceSuite) SetupSuite() {
	s.logger, _ = logging.NewDevelopmentLogger()
	s.secretKey = "test-secret-key-for-testing"
	s.timeout = 10 * time.Minute
	s.service = NewTokenService(s.secretKey, s.timeout, s.logger)
}

func (s *TokenServiceSuite) TestTokenService_NewService() {
	s.T().Run("creates token service successfully", func(t *testing.T) {
		logger, _ := logging.NewDevelopmentLogger()
		secretKey := "test-secret-key"
		timeout := 10 * time.Minute

		service := NewTokenService(secretKey, timeout, logger)

		assert.NotNil(t, service)
		assert.Equal(t, []byte(secretKey), service.secretKey)
		assert.Equal(t, timeout, service.timeout)
		assert.NotNil(t, service.logger)
	})
}

func (s *TokenServiceSuite) TestTokenService_GenerateToken() {
	s.T().Run("generates valid token successfully", func(t *testing.T) {
		ctx := context.Background()
		username := "testuser"

		token, expiresAt, err := s.service.GenerateToken(ctx, username)

		// Verify no error
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Verify expiration time is approximately correct (within 1 second)
		expectedExpiry := time.Now().Add(s.timeout)
		assert.WithinDuration(t, expectedExpiry, expiresAt, 1*time.Second)

		// Verify token can be parsed and contains correct claims
		parsedToken, err := jwt.ParseWithClaims(token, &domain.JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte(s.secretKey), nil
		})

		assert.NoError(t, err)
		assert.True(t, parsedToken.Valid)

		claims, ok := parsedToken.Claims.(*domain.JWTClaims)
		assert.True(t, ok)
		assert.Equal(t, username, claims.Username)
		assert.NotNil(t, claims.ExpiresAt)
	})

	s.T().Run("generates different tokens for same username", func(t *testing.T) {
		ctx := context.Background()
		username := "testuser"

		token1, _, err1 := s.service.GenerateToken(ctx, username)
		time.Sleep(1 * time.Second) // Ensure different expiration timestamps (JWT uses seconds)
		token2, _, err2 := s.service.GenerateToken(ctx, username)

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, token1, token2, "tokens should be different due to different expiration times")
	})

	s.T().Run("generates tokens with correct expiration", func(t *testing.T) {
		ctx := context.Background()
		username := "testuser"
		shortTimeout := 1 * time.Second
		shortService := NewTokenService(s.secretKey, shortTimeout, s.logger)

		token, expiresAt, err := shortService.GenerateToken(ctx, username)

		assert.NoError(t, err)

		// Parse and verify expiration
		parsedToken, err := jwt.ParseWithClaims(token, &domain.JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
			return []byte(s.secretKey), nil
		})

		assert.NoError(t, err)
		claims := parsedToken.Claims.(*domain.JWTClaims)

		// Verify expiration matches returned value
		assert.Equal(t, expiresAt.Unix(), claims.ExpiresAt.Unix())

		// Verify token expires in the future but soon
		assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
		assert.True(t, claims.ExpiresAt.Time.Before(time.Now().Add(2*time.Second)))
	})

	s.T().Run("handles different usernames correctly", func(t *testing.T) {
		ctx := context.Background()
		usernames := []string{"user1", "user2", "admin", "test@example.com"}

		for _, username := range usernames {
			token, _, err := s.service.GenerateToken(ctx, username)

			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			// Verify username in claims
			parsedToken, err := jwt.ParseWithClaims(token, &domain.JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
				return []byte(s.secretKey), nil
			})

			assert.NoError(t, err)
			claims := parsedToken.Claims.(*domain.JWTClaims)
			assert.Equal(t, username, claims.Username)
		}
	})
}

func (s *TokenServiceSuite) TestTokenService_GenerateToken_SigningMethod() {
	s.T().Run("uses HS256 signing method", func(t *testing.T) {
		ctx := context.Background()

		token, _, err := s.service.GenerateToken(ctx, "testuser")

		assert.NoError(t, err)

		// Parse token without validation to check algorithm
		parsedToken, _, err := new(jwt.Parser).ParseUnverified(token, &domain.JWTClaims{})
		assert.NoError(t, err)
		assert.Equal(t, "HS256", parsedToken.Method.Alg())
	})
}
