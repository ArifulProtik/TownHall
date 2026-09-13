package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsEndpointServesPrometheus(t *testing.T) {
	e := echo.New()
	RegisterMetrics(e)
	e.GET("/ping", func(c *echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	mreq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	mrec := httptest.NewRecorder()
	e.ServeHTTP(mrec, mreq)
	require.Equal(t, http.StatusOK, mrec.Code)
	body := mrec.Body.String()
	assert.Contains(t, body, "townhall_http_requests_total")
	assert.Contains(t, body, `route="/ping"`)
	assert.Contains(t, body, "townhall_http_request_duration_seconds")
}
