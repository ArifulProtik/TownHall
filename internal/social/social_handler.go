package social

import (
	"net/http"
	"strconv"
	"strings"

	"ArifulProtik/TownHall/internal/auth"
	"ArifulProtik/TownHall/pkg/response"

	"github.com/labstack/echo/v5"
)

type Handler struct {
	svc    *Service
	secret string
}

func NewHandler(svc *Service, secret string) *Handler {
	return &Handler{svc: svc, secret: secret}
}

func (h *Handler) RegisterRoutes(public *echo.Group, protected *echo.Group) {
	public.GET("/users/:handle/follow/status", h.GetStatus)
	public.GET("/users/:handle/followers", h.ListFollowers)
	public.GET("/users/:handle/following", h.ListFollowing)
	public.GET("/users/:handle/friends", h.ListFriends)
	protected.POST("/users/:handle/follow", h.FollowUser)
	protected.DELETE("/users/:handle/follow", h.UnfollowUser)
}

func (h *Handler) FollowUser(c *echo.Context) error {
	uid, ok := response.CurrentUserID(c)
	if !ok {
		return nil
	}
	handle := c.Param("handle")
	if strings.TrimSpace(handle) == "" {
		return response.BadRequest(c, "handle is required")
	}
	st, err := h.svc.Follow(c.Request().Context(), uid, handle)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(http.StatusOK, st)
}

func (h *Handler) UnfollowUser(c *echo.Context) error {
	uid, ok := response.CurrentUserID(c)
	if !ok {
		return nil
	}
	handle := c.Param("handle")
	if strings.TrimSpace(handle) == "" {
		return response.BadRequest(c, "handle is required")
	}
	st, err := h.svc.Unfollow(c.Request().Context(), uid, handle)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(http.StatusOK, st)
}

// GetStatus is public; a valid token just personalizes is_following/is_friend.
func (h *Handler) GetStatus(c *echo.Context) error {
	handle := c.Param("handle")
	if strings.TrimSpace(handle) == "" {
		return response.BadRequest(c, "handle is required")
	}
	viewerID := auth.ViewerID(c.Request(), h.secret)
	st, err := h.svc.GetStatus(c.Request().Context(), viewerID, handle)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(http.StatusOK, st)
}

func listParams(c *echo.Context) (int, string) {
	limit := defaultListLimit
	if raw := strings.TrimSpace(c.QueryParam("limit")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}
	return clampLimit(limit), c.QueryParam("cursor")
}

func (h *Handler) ListFollowers(c *echo.Context) error {
	handle := c.Param("handle")
	if strings.TrimSpace(handle) == "" {
		return response.BadRequest(c, "handle is required")
	}
	limit, cursor := listParams(c)
	resp, err := h.svc.ListFollowers(c.Request().Context(), handle, limit, cursor)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) ListFollowing(c *echo.Context) error {
	handle := c.Param("handle")
	if strings.TrimSpace(handle) == "" {
		return response.BadRequest(c, "handle is required")
	}
	limit, cursor := listParams(c)
	resp, err := h.svc.ListFollowing(c.Request().Context(), handle, limit, cursor)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) ListFriends(c *echo.Context) error {
	handle := c.Param("handle")
	if strings.TrimSpace(handle) == "" {
		return response.BadRequest(c, "handle is required")
	}
	limit, cursor := listParams(c)
	resp, err := h.svc.ListFriends(c.Request().Context(), handle, limit, cursor)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(http.StatusOK, resp)
}
