package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUser_TableName tests table name method
func TestUser_TableName(t *testing.T) {
	user := User{}
	assert.Equal(t, "m_user", user.TableName())
}

// TestUserStructs tests basic user struct instantiation
func TestUserStructs(t *testing.T) {
	t.Run("User struct", func(t *testing.T) {
		user := User{
			Username: "testuser",
			Password: "hashedpassword",
			Token:    "jwt-token",
		}

		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "hashedpassword", user.Password)
		assert.Equal(t, "jwt-token", user.Token)
	})

	t.Run("LoginRequest struct", func(t *testing.T) {
		req := LoginRequest{
			Username: "testuser",
			Password: "password123",
		}

		assert.Equal(t, "testuser", req.Username)
		assert.Equal(t, "password123", req.Password)
	})

	t.Run("LoginResponse struct", func(t *testing.T) {
		resp := LoginResponse{
			Username: "testuser",
			Token:    "jwt-token",
		}

		assert.Equal(t, "testuser", resp.Username)
		assert.Equal(t, "jwt-token", resp.Token)
	})
}
