package auth

import (
	"net/http"
	"strings"

	"ArifulProtik/TownHall/pkg/response"

	"github.com/labstack/echo/v5"
)

// UserIDKey is the echo-context key carrying the authenticated user id.
const UserIDKey = "user_id"

// publicPaths bypass the auth middleware (signup, login, refresh, health, metrics).
var publicPaths = map[string]struct{}{
	"/api/v1/auth/signup":  {},
	"/api/v1/auth/login":   {},
	"/api/v1/auth/refresh": {},
	"/api/v1/auth/health":  {},
	"/metrics":             {},
}

// Middleware validates Bearer access JWTs and injects the user id.
// Every failure returns a generic 401 without internal details.
func Middleware(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if _, ok := publicPaths[c.Request().URL.Path]; ok {
				return next(c)
			}
			parts := strings.SplitN(c.Request().Header.Get("Authorization"), " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") || parts[1] == "" {
				return c.JSON(http.StatusUnauthorized, response.Map{"error": "unauthorized"})
			}
			uid, err := VerifyAccessToken(secret, parts[1])
			if err != nil {
				return c.JSON(http.StatusUnauthorized, response.Map{"error": "unauthorized"})
			}
			c.Set(UserIDKey, uid)
			return next(c)
		}
	}
}
