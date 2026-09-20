package auth

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"ArifulProtik/TownHall/ent/enttest"
	"ArifulProtik/TownHall/ent/refreshtoken"
	"ArifulProtik/TownHall/ent/user"
	"ArifulProtik/TownHall/internal/config"
	"ArifulProtik/TownHall/pkg/apperror"
	"ArifulProtik/TownHall/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/mattn/go-sqlite3"
)

func newTestService(t *testing.T) (*Service, *bytes.Buffer) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:authservice?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	var buf bytes.Buffer
	svc := NewService(&config.Config{AppEnv: "test"}, client, logger.NewWithWriter("test", "info", &buf))
	return svc, &buf
}

func TestSignupEmail_HappyPath(t *testing.T) {
	svc, _ := newTestService(t)

	u, err := svc.SignupEmail(context.Background(), SignupEmail{
		Name:     " Joe ",
		Email:    "JOE@Example.com",
		Password: "password123",
	})
	require.NoError(t, err)
	require.NotNil(t, u)

	assert.Equal(t, "Joe", u.Name)
	assert.Equal(t, "joe@example.com", u.Email)
	assert.Equal(t, user.ProviderEmail, u.Provider)
	assert.False(t, u.EmailVerified)
	assert.NotEmpty(t, u.ID)

	// Password must be stored hashed, never plaintext.
	assert.NotEqual(t, "password123", u.Password)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("password123")))
}

func TestSignupEmail_DuplicateEmail(t *testing.T) {
	svc, buf := newTestService(t)
	ctx := logger.ContextWithRequestID(context.Background(), "req-dup-1")

	_, err := svc.SignupEmail(ctx, SignupEmail{
		Name:     "Joe",
		Email:    "joe@example.com",
		Password: "password123",
	})
	require.NoError(t, err)

	_, err = svc.SignupEmail(ctx, SignupEmail{
		Name:     "Joe Two",
		Email:    "joe@example.com",
		Password: "password123",
	})
	require.Error(t, err)

	appErr := &apperror.AppError{}
	ok := errors.As(err, &appErr)
	require.True(t, ok, "expected *apperror.AppError, got %T", err)
	assert.Equal(t, 409, appErr.Status)

	assert.Contains(t, buf.String(), "duplicate email")
	assert.Contains(t, buf.String(), "req-dup-1")
}

func newLoginService(t *testing.T) (*Service, context.Context) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:authlogin?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	var buf bytes.Buffer
	cfg := &config.Config{
		AppEnv:        "test",
		JWTSecret:     "test-secret-1234567890",
		JWTAccessTTL:  15 * time.Minute,
		JWTRefreshTTL: 720 * time.Hour,
	}
	svc := NewService(cfg, client, logger.NewWithWriter("test", "info", &buf))
	return svc, context.Background()
}

func mustSignup(ctx context.Context, t *testing.T, svc *Service) {
	t.Helper()
	_, err := svc.SignupEmail(ctx, SignupEmail{Name: "Joe", Email: "joe@example.com", Password: "password123"})
	require.NoError(t, err)
}

func TestLogin_Success(t *testing.T) {
	svc, ctx := newLoginService(t)
	mustSignup(ctx, t, svc)

	pair, err := svc.Login(ctx, LoginRequest{Email: "JOE@example.com", Password: "password123"})
	require.NoError(t, err)
	require.NotNil(t, pair)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshRaw)
	assert.WithinDuration(t, pair.RefreshExp, time.Now().Add(720*time.Hour), 2*time.Minute)

	sub, err := VerifyAccessToken("test-secret-1234567890", pair.AccessToken)
	require.NoError(t, err)
	assert.NotEmpty(t, sub)

	stored, err := svc.db.RefreshToken.Query().
		Where(refreshtoken.TokenHashEQ(HashRefreshToken(pair.RefreshRaw))).
		WithUser().
		Only(ctx)
	require.NoError(t, err)
	assert.Nil(t, stored.RevokedAt)
	assert.Equal(t, sub, stored.Edges.User.ID)
}

func TestLogin_FailureIdentical(t *testing.T) {
	svc, ctx := newLoginService(t)
	mustSignup(ctx, t, svc)

	_, errWrong := svc.Login(ctx, LoginRequest{Email: "joe@example.com", Password: "wrongpassword"})
	_, errUnknown := svc.Login(ctx, LoginRequest{Email: "nobody@example.com", Password: "wrongpassword"})
	require.Error(t, errWrong)
	require.Error(t, errUnknown)

	wrongApp, unknownApp := &apperror.AppError{}, &apperror.AppError{}
	require.ErrorAs(t, errWrong, &wrongApp)
	require.ErrorAs(t, errUnknown, &unknownApp)
	assert.Equal(t, 401, wrongApp.Status)
	assert.Equal(t, wrongApp.Status, unknownApp.Status)
	assert.Equal(t, wrongApp.Message, unknownApp.Message)
	assert.Equal(t, "invalid credentials", wrongApp.Message)
}

func TestRefresh_Rotation(t *testing.T) {
	svc, ctx := newLoginService(t)
	mustSignup(ctx, t, svc)

	first, err := svc.Login(ctx, LoginRequest{Email: "joe@example.com", Password: "password123"})
	require.NoError(t, err)

	second, err := svc.Refresh(ctx, first.RefreshRaw)
	require.NoError(t, err)

	third, err := svc.Refresh(ctx, second.RefreshRaw)
	require.NoError(t, err)
	assert.NotEqual(t, second.RefreshRaw, third.RefreshRaw)
}

func TestRefresh_ReuseRevokesAll(t *testing.T) {
	svc, ctx := newLoginService(t)
	mustSignup(ctx, t, svc)

	first, err := svc.Login(ctx, LoginRequest{Email: "joe@example.com", Password: "password123"})
	require.NoError(t, err)
	second, err := svc.Refresh(ctx, first.RefreshRaw)
	require.NoError(t, err)

	// Replay the already-rotated token: reuse → 401 and full revocation.
	_, err = svc.Refresh(ctx, first.RefreshRaw)
	appErr := &apperror.AppError{}
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 401, appErr.Status)

	_, err = svc.Refresh(ctx, second.RefreshRaw)
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 401, appErr.Status)

	n, err := svc.db.RefreshToken.Query().Where(refreshtoken.RevokedAtIsNil()).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}

func TestRefresh_ExpiredAndUnknown(t *testing.T) {
	svc, ctx := newLoginService(t)
	mustSignup(ctx, t, svc)

	_, err := svc.Refresh(ctx, "never-issued-token")
	appErr := &apperror.AppError{}
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 401, appErr.Status)

	u, err := svc.db.User.Query().Only(ctx)
	require.NoError(t, err)
	_, err = svc.db.RefreshToken.Create().
		SetTokenHash(HashRefreshToken("expired-raw")).
		SetExpiresAt(time.Now().Add(-time.Hour)).
		SetUserID(u.ID).
		Save(ctx)
	require.NoError(t, err)

	_, err = svc.Refresh(ctx, "expired-raw")
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 401, appErr.Status)
}

func TestLogout_SingleAndAll(t *testing.T) {
	svc, ctx := newLoginService(t)
	mustSignup(ctx, t, svc)

	a, err := svc.Login(ctx, LoginRequest{Email: "joe@example.com", Password: "password123"})
	require.NoError(t, err)
	b, err := svc.Login(ctx, LoginRequest{Email: "joe@example.com", Password: "password123"})
	require.NoError(t, err)

	require.NoError(t, svc.Logout(ctx, a.RefreshRaw))
	_, err = svc.Refresh(ctx, b.RefreshRaw)
	require.NoError(t, err)
	_, err = svc.Refresh(ctx, a.RefreshRaw)
	require.Error(t, err)

	c, err := svc.Login(ctx, LoginRequest{Email: "joe@example.com", Password: "password123"})
	require.NoError(t, err)
	u, err := svc.db.User.Query().Only(ctx)
	require.NoError(t, err)
	require.NoError(t, svc.LogoutAll(ctx, u.ID))
	_, err = svc.Refresh(ctx, c.RefreshRaw)
	require.Error(t, err)
	require.NoError(t, svc.Logout(ctx, "nope"))
}
