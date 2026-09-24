package auth

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ArifulProtik/TownHall/pkg/response"

	"github.com/labstack/echo/v5"
)

type Handler struct {
	svc    *Service
	secure bool
	env    string
}

func NewHandler(svc *Service, secure bool, env string) *Handler {
	return &Handler{svc: svc, secure: secure, env: env}
}

func (h *Handler) RegisterRoutes(public *echo.Group, protected *echo.Group) {
	authPublic := public.Group("/auth")
	authPublic.GET("/health", h.HealthCheck)
	authPublic.POST("/signup", h.SignupEmail)
	authPublic.POST("/login", h.Login)
	authPublic.POST("/refresh", h.Refresh)

	authProtected := protected.Group("/auth")
	authProtected.GET("/me", h.Me)
	authProtected.GET("/check-username", h.CheckUsername)
	authProtected.POST("/onboarding", h.SetupUsername)
	authProtected.PUT("/username", h.SetupUsername)
	authProtected.PUT("/password", h.ChangePassword)
	authProtected.POST("/logout", h.Logout)
	authProtected.POST("/logout-all", h.LogoutAll)
}

func (h *Handler) HealthCheck(c *echo.Context) error {
	return c.JSON(200, response.Map{"status": "ok", "env": h.env})
}

func (h *Handler) SignupEmail(c *echo.Context) error {
	var req SignupEmail
	if !response.Bind(c, &req) {
		return nil
	}
	if err := h.checkRateLimit(c, req.Email); err != nil {
		return err
	}

	u, err := h.svc.SignupEmail(c.Request().Context(), req)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(http.StatusCreated, ToUserResponse(u))
}

func (h *Handler) Login(c *echo.Context) error {
	var req LoginRequest
	if !response.Bind(c, &req) {
		return nil
	}
	if err := h.checkRateLimit(c, req.Email); err != nil {
		return err
	}

	pair, err := h.svc.Login(c.Request().Context(), req)
	if err != nil {
		return response.Error(c, err)
	}
	setRefreshCookie(c, h.secure, pair.RefreshRaw, pair.RefreshExp)
	return c.JSON(http.StatusOK, TokenResponse{
		AccessToken: pair.AccessToken,
		ExpiresIn:   int64(time.Until(pair.AccessExp).Seconds()),
	})
}

func (h *Handler) Refresh(c *echo.Context) error {
	ck, err := c.Request().Cookie(RefreshCookieName)
	if err != nil || ck == nil || ck.Value == "" {
		return response.Unauthorized(c)
	}
	pair, err := h.svc.Refresh(c.Request().Context(), ck.Value)
	if err != nil {
		return response.Error(c, err)
	}
	setRefreshCookie(c, h.secure, pair.RefreshRaw, pair.RefreshExp)
	return c.JSON(http.StatusOK, TokenResponse{
		AccessToken: pair.AccessToken,
		ExpiresIn:   int64(time.Until(pair.AccessExp).Seconds()),
	})
}

func (h *Handler) Logout(c *echo.Context) error {
	if ck, err := c.Request().Cookie(RefreshCookieName); err == nil && ck != nil && ck.Value != "" {
		if err := h.svc.Logout(c.Request().Context(), ck.Value); err != nil {
			return response.Error(c, err)
		}
	}
	clearRefreshCookie(c, h.secure)
	return c.JSON(http.StatusOK, response.Map{"status": "ok"})
}

func (h *Handler) LogoutAll(c *echo.Context) error {
	uid, ok := response.CurrentUserID(c)
	if !ok {
		return nil
	}
	if err := h.svc.LogoutAll(c.Request().Context(), uid); err != nil {
		return response.Error(c, err)
	}
	clearRefreshCookie(c, h.secure)
	return c.JSON(http.StatusOK, response.Map{"status": "ok"})
}

func (h *Handler) Me(c *echo.Context) error {
	uid, ok := response.CurrentUserID(c)
	if !ok {
		return nil
	}
	u, err := h.svc.GetUser(c.Request().Context(), uid)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(http.StatusOK, ToUserResponse(u))
}

func (h *Handler) CheckUsername(c *echo.Context) error {
	uid, ok := response.CurrentUserID(c)
	if !ok {
		return nil
	}

	username := c.QueryParam("username")
	if strings.TrimSpace(username) == "" {
		return c.JSON(http.StatusBadRequest, response.Map{"error": "username query parameter is required"})
	}

	avail, reason, err := h.svc.CheckUsername(c.Request().Context(), uid, username)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(http.StatusOK, CheckUsernameResponse{Available: avail, Reason: reason})
}

func (h *Handler) SetupUsername(c *echo.Context) error {
	uid, ok := response.CurrentUserID(c)
	if !ok {
		return nil
	}

	var req SetupUsernameRequest
	if !response.Bind(c, &req) {
		return nil
	}

	u, err := h.svc.SetupUsername(c.Request().Context(), uid, req.Username)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(http.StatusOK, ToUserResponse(u))
}

func (h *Handler) ChangePassword(c *echo.Context) error {
	uid, ok := response.CurrentUserID(c)
	if !ok {
		return nil
	}

	var req ChangePasswordRequest
	if !response.Bind(c, &req) {
		return nil
	}

	if err := h.svc.ChangePassword(c.Request().Context(), uid, req.CurrentPassword, req.NewPassword); err != nil {
		return response.Error(c, err)
	}
	return c.JSON(http.StatusOK, response.Map{"status": "ok"})
}

// errRateLimited marks a request already rejected with 429.
// The response is already committed, so Echo must not write again;
// never return nil on the 429 path.
var errRateLimited = errors.New("rate limit exceeded")

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
