// Package middleware wires the global Echo middleware stack.
package middleware

import (
	"fmt"
	"log/slog"

	"ArifulProtik/TownHall/pkg/logger"

	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"
)

// RequestID must precede RequestLogger so access logs carry the id.
func Register(e *echo.Echo, log *slog.Logger) {
	RegisterMetrics(e)
	e.Use(echomw.Recover())
	e.Use(echomw.RequestIDWithConfig(echomw.RequestIDConfig{
		RequestIDHandler: func(c *echo.Context, requestID string) {
			c.Set(logger.RequestIDKey, requestID)
		},
	}))
	e.Use(echomw.RequestLoggerWithConfig(echomw.RequestLoggerConfig{
		LogLatency:   true,
		LogMethod:    true,
		LogURIPath:   true,
		LogRoutePath: true,
		LogRequestID: true,
		LogStatus:    true,
		LogValuesFunc: func(_ *echo.Context, v echomw.RequestLoggerValues) error {
			msg := fmt.Sprintf("%s %s → %d · %dms", v.Method, v.URIPath, v.Status, v.Latency.Milliseconds())
			attrs := []any{
				slog.String("method", v.Method),
				slog.String("uri", v.URIPath),
				slog.String("route", v.RoutePath),
				slog.Int("status", v.Status),
				slog.Int64("latency_ms", v.Latency.Milliseconds()),
				slog.String("request_id", v.RequestID),
			}
			switch {
			case v.Error != nil || v.Status >= 500:
				attrs = append(attrs, slog.Any("error", v.Error))
				log.Error(msg, attrs...)
			case v.Status >= 400:
				log.Warn(msg, attrs...)
			default:
				log.Info(msg, attrs...)
			}
			return nil
		},
	}))
	e.Use(echomw.CORS("*"))
}
