package response

import (
	"errors"
	"log/slog"
	"net/http"

	"ArifulProtik/TownHall/pkg/apperror"
	"ArifulProtik/TownHall/pkg/validation"

	"github.com/labstack/echo/v5"
)

type Map map[string]any

// UserIDKey is the echo-context key where the auth middleware stores the
// signed-in user id. Handlers read it via CurrentUserID, never directly.
const UserIDKey = "user_id"

// CurrentUserID returns the id the auth middleware stored, writing the 401
// itself when nobody is signed in. False means the response is done.
func CurrentUserID(c *echo.Context) (string, bool) {
	uid, _ := c.Get(UserIDKey).(string)
	if uid == "" {
		_ = Unauthorized(c)
		return "", false
	}
	return uid, true
}

// Error renders a service error as JSON. Expected failures (4xx) stay quiet:
// the access log already records them. Unexpected ones (5xx) log here, the
// only place request handling logs.
func Error(c *echo.Context, err error) error {
	var app *apperror.AppError
	if errors.As(err, &app) {
		if app.Status >= 500 {
			slog.Default().Error("request failed", slog.Any("error", err), slog.Int("status", app.Status))
		}
		return c.JSON(app.Status, Map{"error": app.Message, "code": app.Code})
	}
	slog.Default().Error("request failed", slog.Any("error", err))
	return c.JSON(http.StatusInternalServerError, Map{"error": "internal server error"})
}

// Bind reads JSON into v and validates it, writing the 400 itself.
// False means the response is done; the caller just returns nil.
func Bind(c *echo.Context, v any) bool {
	if err := c.Bind(v); err != nil {
		_ = c.JSON(http.StatusBadRequest, Map{"error": "invalid request body"})
		return false
	}
	if err := c.Validate(v); err != nil {
		var ve *validation.Error
		if errors.As(err, &ve) {
			_ = c.JSON(http.StatusBadRequest, Map{"error": "validation failed", "fields": ve.Fields})
		} else {
			_ = c.JSON(http.StatusBadRequest, Map{"error": "validation failed"})
		}
		return false
	}
	return true
}

func Unauthorized(c *echo.Context) error {
	return c.JSON(http.StatusUnauthorized, Map{"error": "unauthorized", "code": "unauthorized"})
}

func BadRequest(c *echo.Context, msg string) error {
	return c.JSON(http.StatusBadRequest, Map{"error": msg, "code": "bad_request"})
}
