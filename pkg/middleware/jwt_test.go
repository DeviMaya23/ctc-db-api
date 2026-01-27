package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
