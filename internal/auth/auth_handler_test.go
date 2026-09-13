package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ArifulProtik/TownHall/ent/enttest"
	"ArifulProtik/TownHall/internal/config"
	"ArifulProtik/TownHall/pkg/logger"
	"ArifulProtik/TownHall/pkg/validation"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func newTestHandler(t *testing.T) (*echo.Echo, *Handler, *bytes.Buffer) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:authhandler?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })

	e := echo.New()
	e.Validator = validation.New()

	var buf bytes.Buffer
	log := logger.NewWithWriter("test", "info", &buf)
	svc := NewService(&config.Config{AppEnv: "test"}, client, log)
	return e, NewHandler(svc, log), &buf
}

func doSignup(t *testing.T, e *echo.Echo, h *Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup/email", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(logger.RequestIDKey, "req-handler-1")
	require.NoError(t, h.SignupEmail(c))
	return rec
}

func TestSignupEmailHandler_HappyPath(t *testing.T) {
	e, h, _ := newTestHandler(t)

	rec := doSignup(t, e, h, `{"name":"Joe","email":"joe@example.com","password":"password123"}`)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Joe", resp["name"])
	assert.Equal(t, "joe@example.com", resp["email"])
	assert.NotContains(t, resp, "password")
}

func TestSignupEmailHandler_ValidationError(t *testing.T) {
	e, h, buf := newTestHandler(t)

	rec := doSignup(t, e, h, `{"name":"Joe","email":"not-an-email","password":"password123"}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "validation failed", resp["error"])
	assert.Contains(t, resp, "fields")

	assert.Contains(t, buf.String(), "validation failed")
	assert.Contains(t, buf.String(), "req-handler-1")
}

func TestSignupEmailHandler_Duplicate(t *testing.T) {
	e, h, _ := newTestHandler(t)

	first := doSignup(t, e, h, `{"name":"Joe","email":"joe@example.com","password":"password123"}`)
	require.Equal(t, http.StatusCreated, first.Code)

	second := doSignup(t, e, h, `{"name":"Joe Two","email":"joe@example.com","password":"password123"}`)
	assert.Equal(t, http.StatusConflict, second.Code)
}
