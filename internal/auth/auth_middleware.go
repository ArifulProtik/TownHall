package auth

import (
	"ArifulProtik/TownHall/pkg/response"

	"github.com/labstack/echo/v5"
)

const UserIDKey = response.UserIDKey

// Middleware keeps requests with a valid bearer token and rejects the rest
// with the same generic 401, never explaining why.
func Middleware(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			uid, err := bearerUserID(c.Request(), secret)
			if err != nil {
				return response.Unauthorized(c)
			}
			c.Set(UserIDKey, uid)
			return next(c)
		}
	}
}
