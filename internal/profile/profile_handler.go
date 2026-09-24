// Package profile provides user profile viewing, updating, and media uploads.
package profile

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"ArifulProtik/TownHall/internal/auth"
	"ArifulProtik/TownHall/pkg/logger"
	"ArifulProtik/TownHall/pkg/response"
	"ArifulProtik/TownHall/pkg/validation"

	"github.com/labstack/echo/v5"
)

// Handler serves profile and file upload HTTP endpoints.
type Handler struct {
	svc       *Service
	log       *slog.Logger
	jwtSecret string
}

// NewHandler builds a Handler around svc.
func NewHandler(svc *Service, jwtSecret string, log *slog.Logger) *Handler {
	return &Handler{
		svc:       svc,
		log:       log,
		jwtSecret: jwtSecret,
	}
}

// RegisterRoutes mounts profile endpoints on public and protected route groups.
func (h *Handler) RegisterRoutes(public *echo.Group, protected *echo.Group) {
	public.GET("/users/:handle", h.GetProfile)
	protected.PATCH("/profile", h.UpdateProfile)
	protected.POST("/profile/upload", h.UploadFile)
}

// GetProfile returns a user's profile by username or ID.
func (h *Handler) GetProfile(c *echo.Context) error {
	rid, _ := c.Get(logger.RequestIDKey).(string)
	log := logger.WithRequestID(h.log, rid)

	handle := c.Param("handle")
	if strings.TrimSpace(handle) == "" {
		return c.JSON(http.StatusBadRequest, response.Map{"error": "handle is required", "code": "bad_request"})
	}

	viewerID, _ := c.Get(auth.UserIDKey).(string)
	if viewerID == "" && h.jwtSecret != "" {
		authHeader := c.Request().Header.Get("Authorization")
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") && parts[1] != "" {
			if uid, err := auth.VerifyAccessToken(h.jwtSecret, parts[1]); err == nil {
				viewerID = uid
			}
		}
	}

	ctx := logger.ContextWithRequestID(c.Request().Context(), rid)
	profile, err := h.svc.GetProfile(ctx, viewerID, handle)
	if err != nil {
		return response.Error(c, log, err, "get profile")
	}

	return c.JSON(http.StatusOK, profile)
}

// UpdateProfile updates the authenticated user's profile.
func (h *Handler) UpdateProfile(c *echo.Context) error {
	rid, _ := c.Get(logger.RequestIDKey).(string)
	log := logger.WithRequestID(h.log, rid)

	uid, _ := c.Get(auth.UserIDKey).(string)
	if uid == "" {
		log.Warn("update profile: missing user id")
		return c.JSON(http.StatusUnauthorized, response.Map{"error": "unauthorized", "code": "unauthorized"})
	}

	var req UpdateRequest
	if err := c.Bind(&req); err != nil {
		log.Warn("update profile: bad request body", slog.Any("error", err))
		return c.JSON(http.StatusBadRequest, response.Map{"error": "invalid request body"})
	}
	req.normalize()
	if err := c.Validate(&req); err != nil {
		var ve *validation.Error
		if errors.As(err, &ve) {
			log.Warn("update profile: validation failed", slog.Any("fields", ve.Fields))
			return c.JSON(http.StatusBadRequest, response.Map{"error": "validation failed", "fields": ve.Fields})
		}
		log.Warn("update profile: validation failed", slog.Any("error", err))
		return c.JSON(http.StatusBadRequest, response.Map{"error": "validation failed"})
	}

	ctx := logger.ContextWithRequestID(c.Request().Context(), rid)
	u, err := h.svc.UpdateProfile(ctx, uid, req)
	if err != nil {
		return response.Error(c, log, err, "update profile")
	}

	resp := ToResponse(u, uid)
	return c.JSON(http.StatusOK, resp)
}

// maxUploadBytes caps buffered upload bodies at 8 MiB (see UploadFile).
const maxUploadBytes = 8 * 1024 * 1024

// UploadFile handles multipart image uploads.
func (h *Handler) UploadFile(c *echo.Context) error {
	rid, _ := c.Get(logger.RequestIDKey).(string)
	log := logger.WithRequestID(h.log, rid)

	uid, _ := c.Get(auth.UserIDKey).(string)
	if uid == "" {
		log.Warn("upload file: missing user id")
		return c.JSON(http.StatusUnauthorized, response.Map{"error": "unauthorized", "code": "unauthorized"})
	}

	file, err := c.FormFile("file")
	if err != nil {
		log.Warn("upload file: missing file field in form-data", slog.Any("error", err))
		return c.JSON(http.StatusBadRequest, response.Map{"error": "file is required", "code": "bad_request"})
	}

	src, err := file.Open()
	if err != nil {
		log.Error("upload file: open multipart file failed", slog.Any("error", err))
		return c.JSON(http.StatusInternalServerError, response.Map{"error": "failed to open uploaded file"})
	}
	defer src.Close()

	// Trust-but-verify the declared size, then cap the read itself: the
	// extra byte distinguishes "exactly at the limit" from "over it".
	if file.Size > maxUploadBytes {
		log.Warn("upload file: file too large", slog.Int64("size", file.Size))
		return c.JSON(http.StatusRequestEntityTooLarge, response.Map{"error": "file exceeds the 8MB limit", "code": "file_too_large"})
	}

	data, err := io.ReadAll(io.LimitReader(src, maxUploadBytes+1))
	if err != nil {
		log.Error("upload file: read multipart file failed", slog.Any("error", err))
		return c.JSON(http.StatusInternalServerError, response.Map{"error": "failed to read uploaded file"})
	}
	if int64(len(data)) > maxUploadBytes {
		log.Warn("upload file: file too large while reading", slog.Int("bytes_read", len(data)))
		return c.JSON(http.StatusRequestEntityTooLarge, response.Map{"error": "file exceeds the 8MB limit", "code": "file_too_large"})
	}

	ctx := logger.ContextWithRequestID(c.Request().Context(), rid)
	res, err := h.svc.UploadFile(ctx, file.Filename, data)
	if err != nil {
		return response.Error(c, log, err, "upload file")
	}

	return c.JSON(http.StatusOK, res)
}
