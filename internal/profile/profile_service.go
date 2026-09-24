// Package profile provides user profile viewing, updating, and media uploads.
package profile

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ArifulProtik/TownHall/ent"
	"ArifulProtik/TownHall/ent/user"
	"ArifulProtik/TownHall/internal/config"
	"ArifulProtik/TownHall/pkg/apperror"
	"ArifulProtik/TownHall/pkg/logger"

	"github.com/google/uuid"
)

// Service implements profile use cases.
type Service struct {
	cfg        *config.Config
	db         *ent.Client
	log        *slog.Logger
	httpClient *http.Client
	uploadsDir string
}

// NewService constructs a profile Service.
func NewService(cfg *config.Config, db *ent.Client, log *slog.Logger) *Service {
	return &Service{
		cfg:        cfg,
		db:         db,
		log:        log,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		uploadsDir: "./uploads",
	}
}

// SetUploadsDir overrides the local uploads directory (primarily for testing).
func (s *Service) SetUploadsDir(dir string) {
	s.uploadsDir = dir
}

// GetProfile retrieves a user profile by username, id, or "me".
func (s *Service) GetProfile(ctx context.Context, viewerID, handle string) (*Response, error) {
	log := logger.WithContext(ctx, s.log)
	cleanHandle := strings.TrimSpace(handle)
	if cleanHandle == "" {
		return nil, apperror.BadRequest("handle is required")
	}

	if strings.EqualFold(cleanHandle, "me") {
		if viewerID == "" {
			return nil, apperror.Unauthorized("unauthorized")
		}
		cleanHandle = viewerID
	}

	var u *ent.User
	var err error

	// Try querying by username first
	u, err = s.db.User.Query().Where(user.UsernameEQ(strings.ToLower(cleanHandle))).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// If not found by username, try searching by ID
			u, err = s.db.User.Get(ctx, cleanHandle)
			if err != nil {
				if ent.IsNotFound(err) {
					return nil, apperror.NotFound("user not found")
				}
				log.Error("get profile: query by id failed", slog.Any("error", err))
				return nil, apperror.Internal()
			}
		} else {
			log.Error("get profile: query by username failed", slog.Any("error", err))
			return nil, apperror.Internal()
		}
	}

	resp := ToResponse(u, viewerID)
	return &resp, nil
}

// UpdateProfile updates the profile fields of an authenticated user.
func (s *Service) UpdateProfile(ctx context.Context, userID string, req UpdateRequest) (*ent.User, error) {
	log := logger.WithContext(ctx, s.log)

	update := s.db.User.UpdateOneID(userID)

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if len(name) < 2 || len(name) > 100 {
			return nil, apperror.BadRequest("name must be between 2 and 100 characters")
		}
		update.SetName(name)
	}
	if req.Bio != nil {
		bio := strings.TrimSpace(*req.Bio)
		if len(bio) > 280 {
			return nil, apperror.BadRequest("bio must not exceed 280 characters")
		}
		update.SetBio(bio)
	}
	if req.AvatarURL != nil {
		avatarURL := strings.TrimSpace(*req.AvatarURL)
		if len(avatarURL) > 1000 {
			return nil, apperror.BadRequest("avatar url must not exceed 1000 characters")
		}
		update.SetAvatarURL(avatarURL)
	}
	if req.BannerURL != nil {
		bannerURL := strings.TrimSpace(*req.BannerURL)
		if len(bannerURL) > 1000 {
			return nil, apperror.BadRequest("banner url must not exceed 1000 characters")
		}
		update.SetBannerURL(bannerURL)
	}
	if req.Location != nil {
		loc := strings.TrimSpace(*req.Location)
		if len(loc) > 100 {
			return nil, apperror.BadRequest("location must not exceed 100 characters")
		}
		update.SetLocation(loc)
	}
	if req.Website != nil {
		site := strings.TrimSpace(*req.Website)
		if len(site) > 200 {
			return nil, apperror.BadRequest("website must not exceed 200 characters")
		}
		update.SetWebsite(site)
	}

	u, err := update.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperror.NotFound("user not found")
		}
		log.Error("update profile: save failed", slog.Any("error", err))
		return nil, apperror.Internal()
	}

	return u, nil
}

// UploadFile uploads a file to UploadThing if configured, otherwise falls back to local uploads directory.
func (s *Service) UploadFile(ctx context.Context, filename string, contentType string, data []byte) (*UploadResponse, error) {
	log := logger.WithContext(ctx, s.log)

	// Validate content type
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
		"image/gif":  true,
	}
	normalizedType := strings.ToLower(strings.TrimSpace(contentType))
	if !validTypes[normalizedType] {
		return nil, apperror.BadRequest("only jpeg, png, webp, and gif images are supported")
	}

	// Max 8MB
	if len(data) > 8*1024*1024 {
		return nil, apperror.BadRequest("file size exceeds maximum allowed limit of 8MB")
	}

	// If UploadThing is configured, attempt upload
	token := strings.TrimSpace(s.cfg.UploadthingToken)
	if token != "" {
		res, err := s.uploadToUploadThing(ctx, token, filename, normalizedType, data)
		if err == nil {
			return res, nil
		}
		log.Warn("uploadthing failed, falling back to local storage", slog.Any("error", err))
	}

	// Fallback to local file storage
	return s.saveLocalFile(filename, data)
}

type uploadThingFileReq struct {
	Name string `json:"name"`
	Size int    `json:"size"`
	Type string `json:"type"`
}

type uploadThingPrepareReq struct {
	Files              []uploadThingFileReq `json:"files"`
	ContentDisposition string               `json:"contentDisposition"`
}

type uploadThingFileResp struct {
	URL      string            `json:"url"`
	Fields   map[string]string `json:"fields"`
	FileURL  string            `json:"fileUrl"`
	UfsURL   string            `json:"ufsUrl"`
	Key      string            `json:"key"`
	FileName string            `json:"fileName"`
}

func (s *Service) uploadToUploadThing(ctx context.Context, rawToken, filename, contentType string, data []byte) (*UploadResponse, error) {
	apiKey := extractAPIKey(rawToken)
	if apiKey == "" {
		apiKey = rawToken
	}

	prepareReqBody := uploadThingPrepareReq{
		Files: []uploadThingFileReq{
			{
				Name: filename,
				Size: len(data),
				Type: contentType,
			},
		},
		ContentDisposition: "inline",
	}

	reqBytes, err := json.Marshal(prepareReqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal prepare upload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.uploadthing.com/v6/uploadFiles", bytes.NewReader(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("new prepare request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-uploadthing-api-key", apiKey)
	req.Header.Set("x-uploadthing-token", rawToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("prepare upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("uploadthing prepare returned status %d: %s", resp.StatusCode, string(respBody))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read prepare response: %w", err)
	}

	var files []uploadThingFileResp
	if err := json.Unmarshal(respBody, &files); err != nil {
		// Attempt parsing { "data": [...] }
		var wrapper struct {
			Data []uploadThingFileResp `json:"data"`
		}
		if wrapErr := json.Unmarshal(respBody, &wrapper); wrapErr == nil && len(wrapper.Data) > 0 {
			files = wrapper.Data
		} else {
			return nil, fmt.Errorf("unmarshal prepare response: %w", err)
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no upload targets returned from uploadthing")
	}

	target := files[0]
	// Upload binary to target URL
	uploadReq, err := http.NewRequestWithContext(ctx, http.MethodPut, target.URL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create s3 put request: %w", err)
	}
	uploadReq.Header.Set("Content-Type", contentType)

	uploadResp, err := s.httpClient.Do(uploadReq)
	if err != nil {
		return nil, fmt.Errorf("s3 upload failed: %w", err)
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode < 200 || uploadResp.StatusCode >= 300 {
		return nil, fmt.Errorf("s3 upload returned status: %d", uploadResp.StatusCode)
	}

	finalURL := target.UfsURL
	if finalURL == "" {
		finalURL = target.FileURL
	}
	if finalURL == "" && target.Key != "" {
		finalURL = "https://utfs.io/f/" + target.Key
	}

	return &UploadResponse{
		URL:  finalURL,
		Key:  target.Key,
		Name: filename,
		Size: int64(len(data)),
	}, nil
}

func extractAPIKey(token string) string {
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return token
	}
	var parsed struct {
		APIKey string `json:"apiKey"`
	}
	if err := json.Unmarshal(decoded, &parsed); err == nil && parsed.APIKey != "" {
		return parsed.APIKey
	}
	return token
}

func (s *Service) saveLocalFile(filename string, data []byte) (*UploadResponse, error) {
	if err := os.MkdirAll(s.uploadsDir, 0o750); err != nil {
		return nil, fmt.Errorf("create uploads dir: %w", err)
	}

	ext := filepath.Ext(filename)
	cleanName := filepath.Base(filename)
	safeID := uuid.Must(uuid.NewV7()).String()
	storedName := fmt.Sprintf("%s-%s", safeID, cleanName)
	if ext == "" {
		storedName += ".png"
	}
	targetPath := filepath.Join(s.uploadsDir, storedName)

	if err := os.WriteFile(targetPath, data, 0o600); err != nil {
		return nil, fmt.Errorf("write local file: %w", err)
	}

	return &UploadResponse{
		URL:  "/uploads/" + storedName,
		Key:  storedName,
		Name: cleanName,
		Size: int64(len(data)),
	}, nil
}
