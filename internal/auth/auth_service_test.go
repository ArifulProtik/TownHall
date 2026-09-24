package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"ArifulProtik/TownHall/ent/enttest"
	"ArifulProtik/TownHall/ent/refreshtoken"
	"ArifulProtik/TownHall/ent/user"
	"ArifulProtik/TownHall/pkg/apperror"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/mattn/go-sqlite3"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:authservice?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	return NewService(client, "test-secret", 15*time.Minute, 720*time.Hour)
}

func TestSignupEmail_HappyPath(t *testing.T) {
	svc := newTestService(t)

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
	svc := newTestService(t)
	ctx := context.Background()

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
}

func newLoginService(t *testing.T) (*Service, context.Context) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:authlogin?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	svc := NewService(client, "test-secret-1234567890", 15*time.Minute, 720*time.Hour)
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

func TestService_CheckUsername(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	alice, err := svc.SignupEmail(ctx, SignupEmail{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "password123",
	})
	require.NoError(t, err)

	// Set Alice's username
	_, err = svc.SetupUsername(ctx, alice.ID, "alice_99")
	require.NoError(t, err)

	// Alice checking "alice_99" should be available for Alice
	avail, reason, err := svc.CheckUsername(ctx, alice.ID, "alice_99")
	require.NoError(t, err)
	assert.True(t, avail)
	assert.Empty(t, reason)

	// Bob checking "alice_99" should NOT be available
	bob, err := svc.SignupEmail(ctx, SignupEmail{
		Name:     "Bob",
		Email:    "bob@example.com",
		Password: "password123",
	})
	require.NoError(t, err)

	avail, reason, err = svc.CheckUsername(ctx, bob.ID, "alice_99")
	require.NoError(t, err)
	assert.False(t, avail)
	assert.Equal(t, "already_taken", reason)

	// Checking an unused username
	avail, reason, err = svc.CheckUsername(ctx, bob.ID, "bob_the_builder")
	require.NoError(t, err)
	assert.True(t, avail)
	assert.Empty(t, reason)

	// Checking short username
	avail, reason, err = svc.CheckUsername(ctx, bob.ID, "ab")
	require.NoError(t, err)
	assert.False(t, avail)
	assert.Equal(t, "invalid_length", reason)
}

func TestService_SetupUsername_Conflict(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	u1, err := svc.SignupEmail(ctx, SignupEmail{Name: "User 1", Email: "u1@example.com", Password: "password123"})
	require.NoError(t, err)
	_, err = svc.SetupUsername(ctx, u1.ID, "shared_name")
	require.NoError(t, err)

	u2, err := svc.SignupEmail(ctx, SignupEmail{Name: "User 2", Email: "u2@example.com", Password: "password123"})
	require.NoError(t, err)
	_, err = svc.SetupUsername(ctx, u2.ID, "shared_name")
	require.Error(t, err)
	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 409, appErr.Status)
}

func TestService_ChangePassword(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	u, err := svc.SignupEmail(ctx, SignupEmail{
		Name:     "Pass Tester",
		Email:    "pass@example.com",
		Password: "oldpassword123",
	})
	require.NoError(t, err)

	// Wrong current password
	err = svc.ChangePassword(ctx, u.ID, "wrongpassword", "newpassword123")
	require.Error(t, err)
	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Status)

	// Correct current password revokes pre-existing sessions
	_, err = svc.Login(ctx, LoginRequest{Email: "pass@example.com", Password: "oldpassword123"})
	require.NoError(t, err)

	err = svc.ChangePassword(ctx, u.ID, "oldpassword123", "newpassword123")
	require.NoError(t, err)

	n, err := svc.db.RefreshToken.Query().
		Where(
			refreshtoken.HasUserWith(user.IDEQ(u.ID)),
			refreshtoken.RevokedAtIsNil(),
		).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, n, "password change must revoke all refresh tokens")

	// Verify login with new password succeeds
	pair, err := svc.Login(ctx, LoginRequest{Email: "pass@example.com", Password: "newpassword123"})
	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
}
