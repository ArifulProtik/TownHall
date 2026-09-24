// Package response provides shorthand JSON response types.
package response

import (
	"errors"
	"log/slog"
	"net/http"

	"ArifulProtik/TownHall/pkg/apperror"

	"github.com/labstack/echo/v5"
)

// Map is a shorthand JSON object for handler responses.
type Map map[string]interface{}

// Error maps a service-layer error to its HTTP response. AppError values
// keep their status, message, and code; anything else becomes a 500.
// Statuses >= 500 log as errors, the rest as warnings. action names the
// operation in logs, e.g. "update profile".
func Error(c *echo.Context, log *slog.Logger, err error, action string) error {
	var app *apperror.AppError
	if errors.As(err, &app) {
		if app.Status >= 500 {
			log.Error(action+": service error", slog.Any("error", err))
		} else {
			log.Warn(action+": rejected", slog.Int("status", app.Status), slog.String("code", app.Code))
		}
		return c.JSON(app.Status, Map{"error": app.Message, "code": app.Code})
	}
	log.Error(action+": unexpected error", slog.Any("error", err))
	return c.JSON(http.StatusInternalServerError, Map{"error": "internal server error"})
}
