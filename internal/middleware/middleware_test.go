package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"ArifulProtik/TownHall/pkg/logger"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestEcho(t *testing.T) (*echo.Echo, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	e := echo.New()
	Register(e, logger.NewWithWriter("test", "info", &buf))
	e.GET("/ping", func(c *echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})
	return e, &buf
}

func TestRegister_ServesRequestsWithRequestID(t *testing.T) {
	e, buf := newTestEcho(t)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "pong", rec.Body.String())
	assert.NotEmpty(t, rec.Header().Get(echo.HeaderXRequestID), "response must carry a request id")

	out := buf.String()
	assert.Contains(t, out, "method=GET")
	assert.Contains(t, out, "status=200")
	assert.Contains(t, out, "request_id=")
	assert.Contains(t, out, "GET /ping → 200")
}

func TestRegister_PassesThroughInboundRequestID(t *testing.T) {
	e, buf := newTestEcho(t)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(echo.HeaderXRequestID, "in-123")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "in-123", rec.Header().Get(echo.HeaderXRequestID))
	assert.Contains(t, buf.String(), "request_id=in-123")
}
