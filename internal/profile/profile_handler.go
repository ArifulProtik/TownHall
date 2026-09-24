package profile

import (
	"context"
	"io"
	"net/http"
	"strings"

	"ArifulProtik/TownHall/ent"
	"ArifulProtik/TownHall/internal/auth"
	"ArifulProtik/TownHall/internal/filestore"
	"ArifulProtik/TownHall/pkg/response"

	"github.com/labstack/echo/v5"
)

// ServiceAPI is what the handler needs. *Service satisfies it, tests mock it.
type ServiceAPI interface {
	GetProfile(ctx context.Context, viewerID, handle string) (*Response, error)
	UpdateProfile(ctx context.Context, userID string, req UpdateRequest) (*ent.User, error)
	UploadFile(ctx context.Context, filename string, data []byte) (*filestore.Result, error)
	FollowCounts(ctx context.Context, userID string) (int, int)
}

type Handler struct {
	svc    ServiceAPI
	secret string
}

func NewHandler(svc ServiceAPI, secret string) *Handler {
	return &Handler{svc: svc, secret: secret}
}

func (h *Handler) RegisterRoutes(public *echo.Group, protected *echo.Group) {
	public.GET("/users/:handle", h.GetProfile)
	protected.PATCH("/profile", h.UpdateProfile)
	protected.POST("/profile/upload", h.UploadFile)
}

// GetProfile is public; a valid token just reveals a bit more about yourself.
func (h *Handler) GetProfile(c *echo.Context) error {
	handle := c.Param("handle")
	if strings.TrimSpace(handle) == "" {
		return response.BadRequest(c, "handle is required")
	}

	viewerID := auth.ViewerID(c.Request(), h.secret)
	profile, err := h.svc.GetProfile(c.Request().Context(), viewerID, handle)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(http.StatusOK, profile)
}

func (h *Handler) UpdateProfile(c *echo.Context) error {
	uid, ok := response.CurrentUserID(c)
	if !ok {
		return nil
	}

	var req UpdateRequest
	if !response.Bind(c, &req) {
		return nil
	}

	u, err := h.svc.UpdateProfile(c.Request().Context(), uid, req)
	if err != nil {
		return response.Error(c, err)
	}
	resp := ToResponse(u, uid)
	resp.FollowersCount, resp.FollowingCount = h.svc.FollowCounts(c.Request().Context(), u.ID)
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) UploadFile(c *echo.Context) error {
	if _, ok := response.CurrentUserID(c); !ok {
		return nil
	}

	file, err := c.FormFile("file")
	if err != nil {
		return response.BadRequest(c, "file is required")
	}

	src, err := file.Open()
	if err != nil {
		return response.Error(c, err)
	}
	defer src.Close()

	if file.Size > filestore.MaxUploadBytes {
		return c.JSON(http.StatusRequestEntityTooLarge, response.Map{"error": "file exceeds the 8MB limit", "code": "file_too_large"})
	}

	// Cap the read itself: the extra byte tells "exactly at the limit"
	// apart from "over it".
	data, err := io.ReadAll(io.LimitReader(src, filestore.MaxUploadBytes+1))
	if err != nil {
		return response.Error(c, err)
	}
	if int64(len(data)) > filestore.MaxUploadBytes {
		return c.JSON(http.StatusRequestEntityTooLarge, response.Map{"error": "file exceeds the 8MB limit", "code": "file_too_large"})
	}

	res, err := h.svc.UploadFile(c.Request().Context(), file.Filename, data)
	if err != nil {
		return response.Error(c, err)
	}
	return c.JSON(http.StatusOK, res)
}
