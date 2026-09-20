// Package auth implements signup and authentication use cases.
package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

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
	e.POST("/signup", h.SignupEmail)
	e.POST("/login", h.Login)
	e.POST("/refresh", h.Refresh)
	protected := e.Group("", h.AuthMiddleware())
	protected.POST("/logout", h.Logout)
	protected.POST("/logout-all", h.LogoutAll)
}

// AuthMiddleware validates Bearer tokens for the protected auth routes.
func (h *Handler) AuthMiddleware() echo.MiddlewareFunc {
	return Middleware(h.svc.config.JWTSecret)
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
	if err := h.checkRateLimit(c, req.Email); err != nil {
		return err
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

// RefreshCookieName is the refresh-token cookie name.
const RefreshCookieName = "refresh_token"

// RefreshCookiePath scopes the refresh cookie to auth endpoints.
const RefreshCookiePath = "/api/v1/auth"

// Login authenticates email+password and issues tokens.
func (h *Handler) Login(c *echo.Context) error {
	rid, _ := c.Get(logger.RequestIDKey).(string)
	log := logger.WithRequestID(h.log, rid)

	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		log.Warn("login: bad request body", slog.Any("error", err))
		return c.JSON(http.StatusBadRequest, response.Map{"error": "invalid request body"})
	}
	if err := c.Validate(&req); err != nil {
		var ve *validation.Error
		if errors.As(err, &ve) {
			log.Warn("login: validation failed", slog.Any("fields", ve.Fields))
			return c.JSON(http.StatusBadRequest, response.Map{"error": "validation failed", "fields": ve.Fields})
		}
		log.Warn("login: validation failed", slog.Any("error", err))
		return c.JSON(http.StatusBadRequest, response.Map{"error": "validation failed"})
	}
	if err := h.checkRateLimit(c, req.Email); err != nil {
		return err
	}

	ctx := logger.ContextWithRequestID(c.Request().Context(), rid)
	pair, err := h.svc.Login(ctx, req)
	if err != nil {
		var app *apperror.AppError
		if errors.As(err, &app) {
			if app.Status >= 500 {
				log.Error("login: service error", slog.Any("error", err))
			} else {
				log.Warn("login: rejected", slog.Int("status", app.Status), slog.String("code", app.Code))
			}
			return c.JSON(app.Status, response.Map{"error": app.Message, "code": app.Code})
		}
		log.Error("login: unexpected error", slog.Any("error", err))
		return c.JSON(http.StatusInternalServerError, response.Map{"error": "internal server error"})
	}
	setRefreshCookie(c, h.svc.GetAppEnv() == "production", pair.RefreshRaw, pair.RefreshExp)
	return c.JSON(http.StatusOK, TokenResponse{
		AccessToken: pair.AccessToken,
		ExpiresIn:   int64(time.Until(pair.AccessExp).Seconds()),
	})
}

// Refresh rotates the refresh-token cookie and returns a new access token.
func (h *Handler) Refresh(c *echo.Context) error {
	rid, _ := c.Get(logger.RequestIDKey).(string)
	log := logger.WithRequestID(h.log, rid)

	ck, err := c.Request().Cookie(RefreshCookieName)
	if err != nil || ck == nil || ck.Value == "" {
		log.Warn("refresh: missing cookie")
		return c.JSON(http.StatusUnauthorized, response.Map{"error": "unauthorized", "code": "unauthorized"})
	}
	ctx := logger.ContextWithRequestID(c.Request().Context(), rid)
	pair, err := h.svc.Refresh(ctx, ck.Value)
	if err != nil {
		var app *apperror.AppError
		if errors.As(err, &app) {
			if app.Status >= 500 {
				log.Error("refresh: service error", slog.Any("error", err))
			} else {
				log.Warn("refresh: rejected", slog.Int("status", app.Status), slog.String("code", app.Code))
			}
			return c.JSON(app.Status, response.Map{"error": app.Message, "code": app.Code})
		}
		log.Error("refresh: unexpected error", slog.Any("error", err))
		return c.JSON(http.StatusInternalServerError, response.Map{"error": "internal server error"})
	}
	setRefreshCookie(c, h.svc.GetAppEnv() == "production", pair.RefreshRaw, pair.RefreshExp)
	return c.JSON(http.StatusOK, TokenResponse{
		AccessToken: pair.AccessToken,
		ExpiresIn:   int64(time.Until(pair.AccessExp).Seconds()),
	})
}

// Logout revokes the current refresh token and clears the cookie.
func (h *Handler) Logout(c *echo.Context) error {
	rid, _ := c.Get(logger.RequestIDKey).(string)
	log := logger.WithRequestID(h.log, rid)

	if ck, err := c.Request().Cookie(RefreshCookieName); err == nil && ck != nil && ck.Value != "" {
		ctx := logger.ContextWithRequestID(c.Request().Context(), rid)
		if err := h.svc.Logout(ctx, ck.Value); err != nil {
			var app *apperror.AppError
			if errors.As(err, &app) {
				log.Error("logout: service error", slog.Any("error", err))
				return c.JSON(app.Status, response.Map{"error": app.Message, "code": app.Code})
			}
			log.Error("logout: unexpected error", slog.Any("error", err))
			return c.JSON(http.StatusInternalServerError, response.Map{"error": "internal server error"})
		}
	}
	clearRefreshCookie(c, h.svc.GetAppEnv() == "production")
	return c.JSON(http.StatusOK, response.Map{"status": "ok"})
}

// LogoutAll revokes every refresh token for the authenticated user.
func (h *Handler) LogoutAll(c *echo.Context) error {
	rid, _ := c.Get(logger.RequestIDKey).(string)
	log := logger.WithRequestID(h.log, rid)

	uid, _ := c.Get(UserIDKey).(string)
	if uid == "" {
		log.Warn("logout-all: missing user id")
		return c.JSON(http.StatusUnauthorized, response.Map{"error": "unauthorized", "code": "unauthorized"})
	}
	ctx := logger.ContextWithRequestID(c.Request().Context(), rid)
	if err := h.svc.LogoutAll(ctx, uid); err != nil {
		var app *apperror.AppError
		if errors.As(err, &app) {
			log.Error("logout-all: service error", slog.Any("error", err))
			return c.JSON(app.Status, response.Map{"error": app.Message, "code": app.Code})
		}
		log.Error("logout-all: unexpected error", slog.Any("error", err))
		return c.JSON(http.StatusInternalServerError, response.Map{"error": "internal server error"})
	}
	clearRefreshCookie(c, h.svc.GetAppEnv() == "production")
	return c.JSON(http.StatusOK, response.Map{"status": "ok"})
}

// errRateLimited marks a request already rejected with 429.
var errRateLimited = errors.New("rate limit exceeded")

// checkRateLimit enforces email+IP quotas; on excess it writes the 429 itself.
func (h *Handler) checkRateLimit(c *echo.Context, email string) error {
	allowed, retry := h.svc.AllowAttempt(email, c.RealIP())
	if allowed {
		return nil
	}
	secs := int64(retry / time.Second)
	if secs < 1 || retry%time.Second != 0 {
		secs++
	}
	c.Response().Header().Set("Retry-After", strconv.FormatInt(secs, 10))
	if err := c.JSON(http.StatusTooManyRequests, response.Map{"error": "rate limit exceeded", "code": "rate_limited"}); err != nil {
		return err
	}
	return errRateLimited
}

// setRefreshCookie writes the refresh-token cookie scoped to auth endpoints.
func setRefreshCookie(c *echo.Context, prod bool, raw string, exp time.Time) {
	c.SetCookie(&http.Cookie{
		Name:     RefreshCookieName,
		Value:    raw,
		Path:     RefreshCookiePath,
		Expires:  exp,
		HttpOnly: true,
		Secure:   prod,
		SameSite: http.SameSiteStrictMode,
	})
}

// clearRefreshCookie expires the refresh-token cookie.
func clearRefreshCookie(c *echo.Context, prod bool) {
	c.SetCookie(&http.Cookie{
		Name:     RefreshCookieName,
		Value:    "",
		Path:     RefreshCookiePath,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0).UTC(),
		HttpOnly: true,
		Secure:   prod,
		SameSite: http.SameSiteStrictMode,
	})
}
