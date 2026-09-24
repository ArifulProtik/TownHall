package social

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ArifulProtik/TownHall/pkg/response"
	"ArifulProtik/TownHall/pkg/validation"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestHandler(t *testing.T) (*echo.Echo, *Handler, *Service) {
	t.Helper()
	svc := newTestService(t)
	h := NewHandler(svc, "test-jwt-secret")
	e := echo.New()
	e.Validator = validation.New()
	return e, h, svc
}

func TestHandler_FollowRequiresAuth(t *testing.T) {
	e, h, _ := newTestHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/x/follow", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/users/:handle/follow")
	c.SetPathValues(echo.PathValues{{Name: "handle", Value: "x"}})

	require.NoError(t, h.FollowUser(c))
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestHandler_FollowUnfollowRoundTrip(t *testing.T) {
	e, h, svc := newTestHandler(t)
	a := createSocialUser(t, svc.db, "h-a", "h-a@ex.com")
	b := createSocialUser(t, svc.db, "h-b", "h-b@ex.com")

	newCtx := func(method, target string, actorID string) (*echo.Context, *httptest.ResponseRecorder) {
		req := httptest.NewRequest(method, "/api/v1/users/"+target+"/follow", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/users/:handle/follow")
		c.SetPathValues(echo.PathValues{{Name: "handle", Value: target}})
		if actorID != "" {
			c.Set(response.UserIDKey, actorID)
		}
		return c, rec
	}

	c1, rec1 := newCtx(http.MethodPost, b.ID, a.ID)
	require.NoError(t, h.FollowUser(c1))
	assert.Equal(t, http.StatusOK, rec1.Code)
	var st StatusResponse
	require.NoError(t, json.Unmarshal(rec1.Body.Bytes(), &st))
	assert.True(t, st.IsFollowing)
	assert.False(t, st.IsFriend)

	c2, rec2 := newCtx(http.MethodDelete, a.ID, b.ID)
	require.NoError(t, h.UnfollowUser(c2))
	assert.Equal(t, http.StatusOK, rec2.Code)

	// Anonymous status is 200 with is_*=false.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+b.ID+"/follow/status", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/users/:handle/follow/status")
	c.SetPathValues(echo.PathValues{{Name: "handle", Value: b.ID}})
	require.NoError(t, h.GetStatus(c))
	assert.Equal(t, http.StatusOK, rec.Code)
}
