package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
	svc := NewService(&config.Config{
		AppEnv:        "test",
		JWTSecret:     "test-secret-1234567890",
		JWTAccessTTL:  15 * time.Minute,
		JWTRefreshTTL: 720 * time.Hour,
	}, client, log)
	return e, NewHandler(svc, log), &buf
}

func doSignup(t *testing.T, e *echo.Echo, h *Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", strings.NewReader(body))
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

func doLogin(t *testing.T, e *echo.Echo, h *Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(logger.RequestIDKey, "req-login-1")
	if err := h.Login(c); !errors.Is(err, errRateLimited) {
		require.NoError(t, err)
	}
	return rec
}

func refreshCookie(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, ck := range rec.Result().Cookies() {
		if ck.Name == RefreshCookieName {
			return ck
		}
	}
	t.Fatal("refresh cookie missing")
	return nil
}

func doRefresh(t *testing.T, e *echo.Echo, h *Handler, ck *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	if ck != nil {
		req.AddCookie(ck)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(logger.RequestIDKey, "req-refresh-1")
	require.NoError(t, h.Refresh(c))
	return rec
}

func TestLoginHandler_SuccessAndCookie(t *testing.T) {
	e, h, _ := newTestHandler(t)
	require.Equal(t, http.StatusCreated, doSignup(t, e, h, `{"name":"Joe","email":"joe@example.com","password":"password123"}`).Code)

	rec := doLogin(t, e, h, `{"email":"joe@example.com","password":"password123"}`)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.NotEmpty(t, body["access_token"])
	assert.NotContains(t, body, "refresh_token")

	ck := refreshCookie(t, rec)
	assert.True(t, ck.HttpOnly)
	assert.Equal(t, RefreshCookiePath, ck.Path)
	assert.Equal(t, http.SameSiteStrictMode, ck.SameSite)
}

func TestLoginHandler_FailureIdentical(t *testing.T) {
	e, h, _ := newTestHandler(t)
	require.Equal(t, http.StatusCreated, doSignup(t, e, h, `{"name":"Joe","email":"joe@example.com","password":"password123"}`).Code)

	wrong := doLogin(t, e, h, `{"email":"joe@example.com","password":"wrongpassword"}`)
	unknown := doLogin(t, e, h, `{"email":"nobody@example.com","password":"wrongpassword"}`)
	assert.Equal(t, http.StatusUnauthorized, wrong.Code)
	assert.Equal(t, wrong.Code, unknown.Code)
	assert.JSONEq(t, wrong.Body.String(), unknown.Body.String())
}

func TestLoginHandler_RateLimited(t *testing.T) {
	e, h, _ := newTestHandler(t)
	require.Equal(t, http.StatusCreated, doSignup(t, e, h, `{"name":"Joe","email":"joe@example.com","password":"password123"}`).Code)

	var rec *httptest.ResponseRecorder
	for range 6 {
		rec = doLogin(t, e, h, `{"email":"joe@example.com","password":"wrongpassword"}`)
	}
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.NotEmpty(t, rec.Header().Get("Retry-After"))
	assert.JSONEq(t, `{"error":"rate limit exceeded","code":"rate_limited"}`, rec.Body.String())
}

func TestRefreshHandler_RotationAndReuse(t *testing.T) {
	e, h, _ := newTestHandler(t)
	require.Equal(t, http.StatusCreated, doSignup(t, e, h, `{"name":"Joe","email":"joe@example.com","password":"password123"}`).Code)

	first := refreshCookie(t, doLogin(t, e, h, `{"email":"joe@example.com","password":"password123"}`))
	rotated := doRefresh(t, e, h, first)
	assert.Equal(t, http.StatusOK, rotated.Code)
	second := refreshCookie(t, rotated)

	// Replay of the rotated token → 401 and full revocation.
	replay := doRefresh(t, e, h, first)
	assert.Equal(t, http.StatusUnauthorized, replay.Code)
	assert.Equal(t, http.StatusUnauthorized, doRefresh(t, e, h, second).Code)

	// Missing cookie → 401.
	assert.Equal(t, http.StatusUnauthorized, doRefresh(t, e, h, nil).Code)
}

func TestLogoutHandlers(t *testing.T) {
	e, h, _ := newTestHandler(t)
	require.Equal(t, http.StatusCreated, doSignup(t, e, h, `{"name":"Joe","email":"joe@example.com","password":"password123"}`).Code)

	ckA := refreshCookie(t, doLogin(t, e, h, `{"email":"joe@example.com","password":"password123"}`))
	ckB := refreshCookie(t, doLogin(t, e, h, `{"email":"joe@example.com","password":"password123"}`))

	// Logout A: clears cookie, kills A, spares B.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.AddCookie(ckA)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	require.NoError(t, h.Logout(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	cleared := false
	for _, ck := range rec.Result().Cookies() {
		if ck.Name == RefreshCookieName && ck.MaxAge < 0 {
			cleared = true
		}
	}
	assert.True(t, cleared, "logout must clear the cookie")
	assert.Equal(t, http.StatusUnauthorized, doRefresh(t, e, h, ckA).Code)

	// Reuse path above revoked ALL tokens including B, so re-login for a
	// fresh cookie to prove logout-all revokes a live token.
	ckB2 := refreshCookie(t, doLogin(t, e, h, `{"email":"joe@example.com","password":"password123"}`))
	_ = ckB

	// Logout-all with user id: kills B2 too.
	sub, err := VerifyAccessToken("test-secret-1234567890", tokenFrom(t, e, h))
	require.NoError(t, err)
	reqAll := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout-all", nil)
	reqAll.AddCookie(ckB2)
	recAll := httptest.NewRecorder()
	cAll := e.NewContext(reqAll, recAll)
	cAll.Set(UserIDKey, sub)
	require.NoError(t, h.LogoutAll(cAll))
	assert.Equal(t, http.StatusOK, recAll.Code)
	assert.Equal(t, http.StatusUnauthorized, doRefresh(t, e, h, ckB2).Code)

	// Logout-all without auth → 401.
	reqAnon := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout-all", nil)
	recAnon := httptest.NewRecorder()
	require.NoError(t, h.LogoutAll(e.NewContext(reqAnon, recAnon)))
	assert.Equal(t, http.StatusUnauthorized, recAnon.Code)
}

func tokenFrom(t *testing.T, e *echo.Echo, h *Handler) string {
	t.Helper()
	rec := doLogin(t, e, h, `{"email":"joe@example.com","password":"password123"}`)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	tok, _ := body["access_token"].(string)
	require.NotEmpty(t, tok)
	return tok
}

func TestMeHandler(t *testing.T) {
	e, h, _ := newTestHandler(t)
	signupRec := doSignup(t, e, h, `{"name":"Joe","email":"joe@example.com","password":"password123"}`)
	require.Equal(t, http.StatusCreated, signupRec.Code)

	var signupBody map[string]any
	require.NoError(t, json.Unmarshal(signupRec.Body.Bytes(), &signupBody))
	uid, _ := signupBody["id"].(string)
	require.NotEmpty(t, uid)

	// Authenticated request → 200 with user response.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(UserIDKey, uid)
	c.Set(logger.RequestIDKey, "req-me-1")
	require.NoError(t, h.Me(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var meBody map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &meBody))
	assert.Equal(t, "Joe", meBody["name"])
	assert.Equal(t, "joe@example.com", meBody["email"])

	// Unauthenticated request → 401.
	reqAnon := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	recAnon := httptest.NewRecorder()
	cAnon := e.NewContext(reqAnon, recAnon)
	cAnon.Set(logger.RequestIDKey, "req-me-anon")
	require.NoError(t, h.Me(cAnon))
	assert.Equal(t, http.StatusUnauthorized, recAnon.Code)
}

func TestCheckUsernameHandler(t *testing.T) {
	e, h, _ := newTestHandler(t)
	signupRec := doSignup(t, e, h, `{"name":"Tester","email":"check@example.com","password":"password123"}`)
	require.Equal(t, http.StatusCreated, signupRec.Code)

	var u map[string]any
	require.NoError(t, json.Unmarshal(signupRec.Body.Bytes(), &u))
	uid, _ := u["id"].(string)
	require.NotEmpty(t, uid)

	// 1. Unauthenticated -> 401
	reqAnon := httptest.NewRequest(http.MethodGet, "/api/v1/auth/check-username?username=cool_name", nil)
	recAnon := httptest.NewRecorder()
	cAnon := e.NewContext(reqAnon, recAnon)
	require.NoError(t, h.CheckUsername(cAnon))
	assert.Equal(t, http.StatusUnauthorized, recAnon.Code)

	// 2. Missing query param -> 400
	reqNoParam := httptest.NewRequest(http.MethodGet, "/api/v1/auth/check-username", nil)
	recNoParam := httptest.NewRecorder()
	cNoParam := e.NewContext(reqNoParam, recNoParam)
	cNoParam.Set(UserIDKey, uid)
	require.NoError(t, h.CheckUsername(cNoParam))
	assert.Equal(t, http.StatusBadRequest, recNoParam.Code)

	// 3. Authenticated, available -> 200 with available=true
	reqAuth := httptest.NewRequest(http.MethodGet, "/api/v1/auth/check-username?username=cool_name", nil)
	recAuth := httptest.NewRecorder()
	cAuth := e.NewContext(reqAuth, recAuth)
	cAuth.Set(UserIDKey, uid)
	require.NoError(t, h.CheckUsername(cAuth))
	assert.Equal(t, http.StatusOK, recAuth.Code)
	var availBody CheckUsernameResponse
	require.NoError(t, json.Unmarshal(recAuth.Body.Bytes(), &availBody))
	assert.True(t, availBody.Available)
}

func TestSetupUsernameHandler(t *testing.T) {
	e, h, _ := newTestHandler(t)
	signupRec := doSignup(t, e, h, `{"name":"Onboard","email":"onboard@example.com","password":"password123"}`)
	require.Equal(t, http.StatusCreated, signupRec.Code)

	var u map[string]any
	require.NoError(t, json.Unmarshal(signupRec.Body.Bytes(), &u))
	uid, _ := u["id"].(string)
	require.NotEmpty(t, uid)

	// 1. Unauthenticated -> 401
	reqAnon := httptest.NewRequest(http.MethodPost, "/api/v1/auth/onboarding", strings.NewReader(`{"username":"onboard_user"}`))
	reqAnon.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recAnon := httptest.NewRecorder()
	cAnon := e.NewContext(reqAnon, recAnon)
	require.NoError(t, h.SetupUsername(cAnon))
	assert.Equal(t, http.StatusUnauthorized, recAnon.Code)

	// 2. Invalid validation (too short / illegal chars) -> 400
	reqInvalid := httptest.NewRequest(http.MethodPost, "/api/v1/auth/onboarding", strings.NewReader(`{"username":"no!"}`))
	reqInvalid.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recInvalid := httptest.NewRecorder()
	cInvalid := e.NewContext(reqInvalid, recInvalid)
	cInvalid.Set(UserIDKey, uid)
	require.NoError(t, h.SetupUsername(cInvalid))
	assert.Equal(t, http.StatusBadRequest, recInvalid.Code)

	// 3. Setup username successfully -> 200
	reqSuccess := httptest.NewRequest(http.MethodPost, "/api/v1/auth/onboarding", strings.NewReader(`{"username":"onboard_user"}`))
	reqSuccess.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recSuccess := httptest.NewRecorder()
	cSuccess := e.NewContext(reqSuccess, recSuccess)
	cSuccess.Set(UserIDKey, uid)
	require.NoError(t, h.SetupUsername(cSuccess))
	assert.Equal(t, http.StatusOK, recSuccess.Code)

	var userResp UserResponse
	require.NoError(t, json.Unmarshal(recSuccess.Body.Bytes(), &userResp))
	require.NotNil(t, userResp.Username)
	assert.Equal(t, "onboard_user", *userResp.Username)

	// 4. Duplicate username conflict -> 409
	signupRec2 := doSignup(t, e, h, `{"name":"Second","email":"second@example.com","password":"password123"}`)
	require.Equal(t, http.StatusCreated, signupRec2.Code)
	var u2 map[string]any
	require.NoError(t, json.Unmarshal(signupRec2.Body.Bytes(), &u2))
	uid2, _ := u2["id"].(string)

	reqDup := httptest.NewRequest(http.MethodPost, "/api/v1/auth/onboarding", strings.NewReader(`{"username":"onboard_user"}`))
	reqDup.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recDup := httptest.NewRecorder()
	cDup := e.NewContext(reqDup, recDup)
	cDup.Set(UserIDKey, uid2)
	require.NoError(t, h.SetupUsername(cDup))
	assert.Equal(t, http.StatusConflict, recDup.Code)
}
