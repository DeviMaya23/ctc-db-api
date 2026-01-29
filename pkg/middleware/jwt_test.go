package middleware

import (
	"lizobly/ctc-db-api/pkg/domain"
	"lizobly/ctc-db-api/pkg/logging"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJWTMiddleware_Success(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", "test-secret-key")

	middleware := NewJWTMiddleware()

	assert.NotNil(t, middleware, "middleware should not be nil")
}

func TestNewJWTMiddleware_Panic_WhenSecretKeyNotSet(t *testing.T) {
	// Ensure JWT_SECRET_KEY is not set
	t.Setenv("JWT_SECRET_KEY", "")

	defer func() {
		r := recover()
		assert.NotNil(t, r, "should panic when JWT_SECRET_KEY is not set")
		assert.Equal(t, "JWT_SECRET_KEY is not set", r)
	}()

	NewJWTMiddleware()
}

func TestJWTMiddleware_TokenValidation(t *testing.T) {
	secretKey := "test-secret-key"

	tests := []struct {
		name           string
		setupToken     func() string
		expectError    bool
		expectUserID   string
		validateUserID bool
	}{
		{
			name: "valid token",
			setupToken: func() string {
				claims := &domain.JWTClaims{
					Username: "testuser",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte(secretKey))
				return tokenString
			},
			expectError:    false,
			expectUserID:   "testuser",
			validateUserID: true,
		},
		{
			name: "invalid token format",
			setupToken: func() string {
				return "invalid-token"
			},
			expectError:    true,
			validateUserID: false,
		},
		{
			name: "missing token",
			setupToken: func() string {
				return ""
			},
			expectError:    true,
			validateUserID: false,
		},
		{
			name: "expired token",
			setupToken: func() string {
				claims := &domain.JWTClaims{
					Username: "testuser",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte(secretKey))
				return tokenString
			},
			expectError:    true,
			validateUserID: false,
		},
		{
			name: "wrong signing key",
			setupToken: func() string {
				claims := &domain.JWTClaims{
					Username: "testuser",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte("wrong-secret-key"))
				return tokenString
			},
			expectError:    true,
			validateUserID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("JWT_SECRET_KEY", secretKey)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)

			tokenString := tt.setupToken()
			if tokenString != "" {
				req.Header.Set("Authorization", "Bearer "+tokenString)
			}

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			middleware := NewJWTMiddleware()
			handler := middleware(func(c echo.Context) error {
				if tt.validateUserID {
					userID := logging.GetUserID(c.Request().Context())
					assert.Equal(t, tt.expectUserID, userID)
				}
				return c.String(http.StatusOK, "success")
			})

			err := handler(c)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, rec.Code)
			}
		})
	}
}

func TestJWTMiddleware_Skipper_LoginPath(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", "test-secret-key")

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/login", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := NewJWTMiddleware()
	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "login success")
	})

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "login success", rec.Body.String())
}

func TestJWTMiddleware_SuccessHandler_InjectsUserID(t *testing.T) {
	secretKey := "test-secret-key"
	t.Setenv("JWT_SECRET_KEY", secretKey)

	tests := []struct {
		name     string
		username string
	}{
		{"standard user", "testuser"},
		{"admin user", "admin"},
		{"user with special chars", "user@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &domain.JWTClaims{
				Username: tt.username,
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
				},
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, err := token.SignedString([]byte(secretKey))
			require.NoError(t, err)

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/protected", nil)
			req.Header.Set("Authorization", "Bearer "+tokenString)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			middleware := NewJWTMiddleware()
			handler := middleware(func(c echo.Context) error {
				userID := logging.GetUserID(c.Request().Context())
				assert.Equal(t, tt.username, userID)
				return c.String(http.StatusOK, "success")
			})

			err = handler(c)
			assert.NoError(t, err)
		})
	}
}
