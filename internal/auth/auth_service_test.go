package auth

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"ArifulProtik/TownHall/ent/enttest"
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
