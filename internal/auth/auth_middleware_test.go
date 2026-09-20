package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ArifulProtik/TownHall/pkg/validation"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runMiddleware(secret, method, target, bearer string) *httptest.ResponseRecorder {
	e := echo.New()
	e.Validator = validation.New()
	var gotUser string
	h := Middleware(secret)(func(c *echo.Context) error {
		gotUser, _ = c.Get(UserIDKey).(string)
		return c.String(http.StatusOK, gotUser)
	})
	req := httptest.NewRequest(method, target, nil)
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	rec := httptest.NewRecorder()
	_ = h(e.NewContext(req, rec))
	return rec
}

func TestMiddleware_ValidBearer(t *testing.T) {
	tok, err := MintAccessToken("s3cret", "user-9", 15*time.Minute)
	require.NoError(t, err)
	rec := runMiddleware("s3cret", http.MethodGet, "/api/v1/protected", tok)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "user-9", rec.Body.String())
}

func TestMiddleware_Rejections(t *testing.T) {
	tok, err := MintAccessToken("s3cret", "user-9", 15*time.Minute)
	require.NoError(t, err)
	expired, err := MintAccessToken("s3cret", "user-9", -time.Minute)
	require.NoError(t, err)

	for name, bearer := range map[string]string{
		"missing":       "",
		"wrong secret":  mustMint(t, "other", "user-9"),
		"expired":       expired,
		"malformed":     "abc.def",
		"empty subject": mustMintEmpty(t, "s3cret"),
	} {
		rec := runMiddleware("s3cret", http.MethodGet, "/api/v1/protected", bearer)
		assert.Equal(t, http.StatusUnauthorized, rec.Code, name)
		assert.JSONEq(t, `{"error":"unauthorized"}`, rec.Body.String(), name)
	}
	_ = tok
}

func mustMint(t *testing.T, secret, sub string) string {
	t.Helper()
	tok, err := MintAccessToken(secret, sub, 15*time.Minute)
	require.NoError(t, err)
	return tok
}

func mustMintEmpty(t *testing.T, secret string) string {
	t.Helper()
	tok, err := MintAccessToken(secret, "", 15*time.Minute)
	require.NoError(t, err)
	return tok
}

func TestMiddleware_SkipPaths(t *testing.T) {
	for _, p := range []string{"/api/v1/auth/signup", "/api/v1/auth/login", "/api/v1/auth/refresh", "/api/v1/auth/health", "/metrics"} {
		rec := runMiddleware("s3cret", http.MethodPost, p, "")
		assert.Equal(t, http.StatusOK, rec.Code, p)
		assert.Empty(t, rec.Body.String(), p)
	}
}
