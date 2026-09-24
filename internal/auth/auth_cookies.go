package auth

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
)

const RefreshCookieName = "refresh_token"

const RefreshCookiePath = "/api/v1/auth"

// refreshCookie builds the refresh-token cookie; an empty value clears it.
func refreshCookie(secure bool, value string, exp time.Time) *http.Cookie {
	jar := &http.Cookie{
		Name:     RefreshCookieName,
		Path:     RefreshCookiePath,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	}
	if value == "" {
		jar.MaxAge = -1
		jar.Expires = time.Unix(0, 0).UTC()
		return jar
	}
	jar.Value = value
	jar.Expires = exp
	return jar
}

func setRefreshCookie(c *echo.Context, secure bool, raw string, exp time.Time) {
	c.SetCookie(refreshCookie(secure, raw, exp))
}

func clearRefreshCookie(c *echo.Context, secure bool) {
	c.SetCookie(refreshCookie(secure, "", time.Time{}))
}
