// Package auth implements signup and authentication use cases.
package auth

import (
	"errors"
	"log/slog"
	"net/http"

	"ArifulProtik/TownHall/pkg/apperror"
	"ArifulProtik/TownHall/pkg/logger"
	"ArifulProtik/TownHall/pkg/response"
	"ArifulProtik/TownHall/pkg/validation"

	"github.com/labstack/echo/v5"
)

// Handler serves the auth HTTP endpoints.
type Handler struct {
	svc *Service
	log *slog.Logger
}

// RegisterRoutes mounts the auth endpoints on e.
func (h *Handler) RegisterRoutes(e *echo.Group) {
	e.GET("/health", h.HealthCheck)
	e.POST("/signup/email", h.SignupEmail)
}

// NewHandler builds an Handler around svc.
func NewHandler(svc *Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// HealthCheck reports service liveness and environment.
func (h *Handler) HealthCheck(c *echo.Context) error {
	return c.JSON(200, response.Map{"status": "ok", "env": h.svc.GetAppEnv()})
}

// SignupEmail validates the request, creates the user, and returns 201.
func (h *Handler) SignupEmail(c *echo.Context) error {
	rid, _ := c.Get(logger.RequestIDKey).(string)
	log := logger.WithRequestID(h.log, rid)

	var req SignupEmail
	if err := c.Bind(&req); err != nil {
		log.Warn("signup: bad request body", slog.Any("error", err))
		return c.JSON(http.StatusBadRequest, response.Map{"error": "invalid request body"})
	}
	if err := c.Validate(&req); err != nil {
		var ve *validation.Error
		if errors.As(err, &ve) {
			log.Warn("signup: validation failed", slog.Any("fields", ve.Fields))
			return c.JSON(http.StatusBadRequest, response.Map{"error": "validation failed", "fields": ve.Fields})
		}
		log.Warn("signup: validation failed", slog.Any("error", err))
		return c.JSON(http.StatusBadRequest, response.Map{"error": "validation failed"})
	}

	ctx := logger.ContextWithRequestID(c.Request().Context(), rid)
	u, err := h.svc.SignupEmail(ctx, req)
	if err != nil {
		var app *apperror.AppError
		if errors.As(err, &app) {
			if app.Status >= 500 {
				log.Error("signup: service error", slog.Any("error", err))
			} else {
				log.Warn("signup: rejected", slog.Int("status", app.Status), slog.String("code", app.Code))
			}
			return c.JSON(app.Status, response.Map{"error": app.Message, "code": app.Code})
		}
		log.Error("signup: unexpected error", slog.Any("error", err))
		return c.JSON(http.StatusInternalServerError, response.Map{"error": "internal server error"})
	}
	return c.JSON(http.StatusCreated, ToUserResponse(u))
}
