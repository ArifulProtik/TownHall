// Package filestore stores uploaded images: UploadThing when configured,
// local disk otherwise. Any domain needing uploads uses this; nobody
// reimplements sniffing or naming.
package filestore

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ArifulProtik/TownHall/pkg/apperror"

	"github.com/google/uuid"
)

type Result struct {
	URL  string `json:"url"`
	Key  string `json:"key,omitempty"`
	Name string `json:"name,omitempty"`
	Size int64  `json:"size,omitempty"`
}

type Storage struct {
	token      string
	uploadsDir string
	httpClient *http.Client
}

func New(uploadthingToken, uploadsDir string) *Storage {
	return &Storage{
		token:      uploadthingToken,
		uploadsDir: uploadsDir,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// MaxUploadBytes caps uploads at 8 MiB. Handlers use it to cap reads
// before the bytes ever reach Upload.
const MaxUploadBytes = 8 * 1024 * 1024

func (s *Storage) Upload(ctx context.Context, filename string, data []byte) (*Result, error) {
	imageType, ok := sniffImageType(data)
	if !ok {
		return nil, apperror.BadRequest("only jpeg, png, webp, and gif images are supported")
	}
	if len(data) > MaxUploadBytes {
		return nil, apperror.BadRequest("file size exceeds maximum allowed limit of 8MB")
	}

	if token := strings.TrimSpace(s.token); token != "" {
		if res, err := s.uploadToUploadThing(ctx, token, filename, imageType, data); err == nil {
			return res, nil
		} else {
			slog.Default().Warn("uploadthing failed, falling back to local storage", slog.Any("error", err))
		}
	}

	return s.saveLocalFile(filename, imageType, data)
}

// sniffImageType trusts the bytes, never the client-sent Content-Type.
func sniffImageType(data []byte) (string, bool) {
	base, _, _ := strings.Cut(http.DetectContentType(data), ";")
	switch strings.ToLower(strings.TrimSpace(base)) {
	case "image/jpeg":
		return "image/jpeg", true
	case "image/png":
		return "image/png", true
	case "image/webp":
		return "image/webp", true
	case "image/gif":
		return "image/gif", true
	default:
		return "", false
	}
}

func storedExtension(imageType string) string {
	switch imageType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".bin"
	}
}

func (s *Storage) saveLocalFile(filename string, imageType string, data []byte) (*Result, error) {
	if err := os.MkdirAll(s.uploadsDir, 0o750); err != nil {
		return nil, err
	}

	// Random name plus the sniffed extension: nothing client-controlled
	// ends up on disk.
	storedName := uuid.Must(uuid.NewV7()).String() + storedExtension(imageType)
	targetPath := filepath.Join(s.uploadsDir, storedName)

	if err := os.WriteFile(targetPath, data, 0o600); err != nil {
		return nil, err
	}

	return &Result{
		URL:  "/uploads/" + storedName,
		Key:  storedName,
		Name: filepath.Base(filename),
		Size: int64(len(data)),
	}, nil
}
