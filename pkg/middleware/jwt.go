package middleware

import (
	"lizobly/cotc-db-api/pkg/domain"
	"lizobly/cotc-db-api/pkg/logging"
	"os"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

func NewJWTMiddleware() echo.MiddlewareFunc {

	jwtSecretKey := os.Getenv("JWT_SECRET_KEY")
	cfg := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(domain.JWTClaims)
		},
		SigningKey: []byte(jwtSecretKey),
		Skipper: func(c echo.Context) bool {
			return c.Request().URL.Path == "/api/v1/login"
		},
		SuccessHandler: func(c echo.Context) {
			// Extract username from JWT claims and inject into context
			token := c.Get("user").(*jwt.Token)
			claims := token.Claims.(*domain.JWTClaims)

			// Enrich context with user ID for logging
			ctx := logging.WithUserID(c.Request().Context(), claims.Username)
			c.SetRequest(c.Request().WithContext(ctx))
		},
	}

	return echojwt.WithConfig(cfg)
}
