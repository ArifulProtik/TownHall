// Package profile provides user profile viewing, updating, and media uploads.
package profile

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"ArifulProtik/TownHall/internal/auth"
	"ArifulProtik/TownHall/pkg/apperror"
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
		var app *apperror.AppError
		if errors.As(err, &app) {
			if app.Status >= 500 {
				log.Error("get profile: error", slog.Any("error", err))
			} else {
				log.Warn("get profile: rejected", slog.Int("status", app.Status), slog.String("code", app.Code))
			}
			return c.JSON(app.Status, response.Map{"error": app.Message, "code": app.Code})
		}
		log.Error("get profile: unexpected error", slog.Any("error", err))
		return c.JSON(http.StatusInternalServerError, response.Map{"error": "internal server error"})
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
		var app *apperror.AppError
		if errors.As(err, &app) {
			if app.Status >= 500 {
				log.Error("update profile: error", slog.Any("error", err))
			} else {
				log.Warn("update profile: rejected", slog.Int("status", app.Status), slog.String("code", app.Code))
			}
			return c.JSON(app.Status, response.Map{"error": app.Message, "code": app.Code})
		}
		log.Error("update profile: unexpected error", slog.Any("error", err))
		return c.JSON(http.StatusInternalServerError, response.Map{"error": "internal server error"})
	}

	resp := ToResponse(u, uid)
	return c.JSON(http.StatusOK, resp)
}

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

	data, err := io.ReadAll(src)
	if err != nil {
		log.Error("upload file: read multipart file failed", slog.Any("error", err))
		return c.JSON(http.StatusInternalServerError, response.Map{"error": "failed to read uploaded file"})
	}

	contentType := file.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = http.DetectContentType(data)
	}

	ctx := logger.ContextWithRequestID(c.Request().Context(), rid)
	res, err := h.svc.UploadFile(ctx, file.Filename, contentType, data)
	if err != nil {
		var app *apperror.AppError
		if errors.As(err, &app) {
			if app.Status >= 500 {
				log.Error("upload file: error", slog.Any("error", err))
			} else {
				log.Warn("upload file: rejected", slog.Int("status", app.Status), slog.String("code", app.Code))
			}
			return c.JSON(app.Status, response.Map{"error": app.Message, "code": app.Code})
		}
		log.Error("upload file: unexpected error", slog.Any("error", err))
		return c.JSON(http.StatusInternalServerError, response.Map{"error": "internal server error"})
	}

	return c.JSON(http.StatusOK, res)
}
