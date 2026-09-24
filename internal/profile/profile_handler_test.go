package profile

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ArifulProtik/TownHall/ent"
	"ArifulProtik/TownHall/ent/enttest"
	"ArifulProtik/TownHall/ent/user"
	"ArifulProtik/TownHall/internal/auth"
	"ArifulProtik/TownHall/internal/config"
	"ArifulProtik/TownHall/pkg/logger"
	"ArifulProtik/TownHall/pkg/validation"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func newTestProfileHandler(t *testing.T) (*echo.Echo, *Handler, *ent.Client) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:profilehandler?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })

	var buf bytes.Buffer
	log := logger.NewWithWriter("test", "info", &buf)
	cfg := &config.Config{AppEnv: "test"}
	svc := NewService(cfg, client, log)
	svc.SetUploadsDir(t.TempDir())
	h := NewHandler(svc, "test-jwt-secret", log)

	e := echo.New()
	e.Validator = validation.New()
	return e, h, client
}

func TestProfileHandler_GetProfile(t *testing.T) {
	e, h, client := newTestProfileHandler(t)
	ctx := context.Background()

	u, err := client.User.Create().
		SetName("Jordan Fox").
		SetEmail("jordan@example.com").
		SetPassword("hashedpass123").
		SetUsername("jordan_fox").
		SetBio("Building TownHall").
		SetProvider(user.ProviderEmail).
		Save(ctx)
	require.NoError(t, err)

	// 1. Get by username without auth
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/jordan_fox", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/users/:handle")
	c.SetPathValues(echo.PathValues{{Name: "handle", Value: "jordan_fox"}})

	require.NoError(t, h.GetProfile(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Jordan Fox", resp.Name)
	assert.Equal(t, "jordan_fox", *resp.Username)
	assert.False(t, resp.IsSelf)

	// 2. Get by username with auth (owner)
	recAuth := httptest.NewRecorder()
	cAuth := e.NewContext(req, recAuth)
	cAuth.SetPath("/api/v1/users/:handle")
	cAuth.SetPathValues(echo.PathValues{{Name: "handle", Value: "jordan_fox"}})
	cAuth.Set(auth.UserIDKey, u.ID)

	require.NoError(t, h.GetProfile(cAuth))
	assert.Equal(t, http.StatusOK, recAuth.Code)

	var respAuth Response
	require.NoError(t, json.Unmarshal(recAuth.Body.Bytes(), &respAuth))
	assert.True(t, respAuth.IsSelf)
	assert.Equal(t, "jordan@example.com", respAuth.Email)

	// 3. Not found handle
	recNotFound := httptest.NewRecorder()
	cNotFound := e.NewContext(req, recNotFound)
	cNotFound.SetPath("/api/v1/users/:handle")
	cNotFound.SetPathValues(echo.PathValues{{Name: "handle", Value: "ghost"}})

	require.NoError(t, h.GetProfile(cNotFound))
	assert.Equal(t, http.StatusNotFound, recNotFound.Code)
}

func TestProfileHandler_UpdateProfile(t *testing.T) {
	e, h, client := newTestProfileHandler(t)
	ctx := context.Background()

	u, err := client.User.Create().
		SetName("Original Name").
		SetEmail("orig@example.com").
		SetPassword("hashedpass123").
		SetUsername("original").
		SetProvider(user.ProviderEmail).
		Save(ctx)
	require.NoError(t, err)

	// 1. Unauthenticated -> 401
	body := `{"name":"New Name","bio":"Updated bio"}`
	reqAnon := httptest.NewRequest(http.MethodPatch, "/api/v1/profile", strings.NewReader(body))
	reqAnon.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recAnon := httptest.NewRecorder()
	cAnon := e.NewContext(reqAnon, recAnon)

	require.NoError(t, h.UpdateProfile(cAnon))
	assert.Equal(t, http.StatusUnauthorized, recAnon.Code)

	// 2. Authenticated success -> 200
	reqAuth := httptest.NewRequest(http.MethodPatch, "/api/v1/profile", strings.NewReader(body))
	reqAuth.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recAuth := httptest.NewRecorder()
	cAuth := e.NewContext(reqAuth, recAuth)
	cAuth.Set(auth.UserIDKey, u.ID)

	require.NoError(t, h.UpdateProfile(cAuth))
	assert.Equal(t, http.StatusOK, recAuth.Code)

	var resp Response
	require.NoError(t, json.Unmarshal(recAuth.Body.Bytes(), &resp))
	assert.Equal(t, "New Name", resp.Name)
	assert.Equal(t, "Updated bio", resp.Bio)
}

func TestProfileHandler_UploadFile(t *testing.T) {
	e, h, client := newTestProfileHandler(t)
	ctx := context.Background()

	u, err := client.User.Create().
		SetName("Uploader").
		SetEmail("uploader@example.com").
		SetPassword("hashedpass123").
		SetUsername("uploader").
		SetProvider(user.ProviderEmail).
		Save(ctx)
	require.NoError(t, err)

	// Build multipart form body
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "test.png")
	require.NoError(t, err)
	_, err = part.Write([]byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15c4"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	// 1. Unauthenticated -> 401
	reqAnon := httptest.NewRequest(http.MethodPost, "/api/v1/profile/upload", bytes.NewReader(body.Bytes()))
	reqAnon.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	recAnon := httptest.NewRecorder()
	cAnon := e.NewContext(reqAnon, recAnon)

	require.NoError(t, h.UploadFile(cAnon))
	assert.Equal(t, http.StatusUnauthorized, recAnon.Code)

	// 2. Authenticated upload -> 200
	reqAuth := httptest.NewRequest(http.MethodPost, "/api/v1/profile/upload", bytes.NewReader(body.Bytes()))
	reqAuth.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	recAuth := httptest.NewRecorder()
	cAuth := e.NewContext(reqAuth, recAuth)
	cAuth.Set(auth.UserIDKey, u.ID)

	require.NoError(t, h.UploadFile(cAuth))
	assert.Equal(t, http.StatusOK, recAuth.Code)

	var uploadResp UploadResponse
	require.NoError(t, json.Unmarshal(recAuth.Body.Bytes(), &uploadResp))
	assert.NotEmpty(t, uploadResp.URL)
	assert.Equal(t, "test.png", uploadResp.Name)
}
