package profile

import (
	"context"
	"strings"
	"testing"

	"ArifulProtik/TownHall/ent"
	"ArifulProtik/TownHall/ent/enttest"
	"ArifulProtik/TownHall/ent/user"
	"ArifulProtik/TownHall/internal/filestore"
	"ArifulProtik/TownHall/pkg/apperror"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func newTestProfileService(t *testing.T) (*Service, *ent.Client) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:profileservice?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	svc := NewService(client, filestore.New("", t.TempDir()))
	return svc, client
}

func createTestUser(t *testing.T, client *ent.Client, username, email string) *ent.User {
	t.Helper()
	u, err := client.User.Create().
		SetName("Alex River").
		SetEmail(email).
		SetPassword("hashedpass123").
		SetUsername(username).
		SetBio("Software engineer & open source enthusiast").
		SetLocation("San Francisco, CA").
		SetWebsite("https://alexriver.dev").
		SetAvatarURL("https://images.example.com/avatar.png").
		SetBannerURL("https://images.example.com/banner.png").
		SetEmailVerified(true).
		SetProvider(user.ProviderEmail).
		Save(context.Background())
	require.NoError(t, err)
	return u
}

func TestProfileService_GetProfile(t *testing.T) {
	svc, client := newTestProfileService(t)
	ctx := context.Background()

	u := createTestUser(t, client, "alexriver", "alex@example.com")

	// 1. Get by username
	p1, err := svc.GetProfile(ctx, "", "alexriver")
	require.NoError(t, err)
	require.NotNil(t, p1)
	assert.Equal(t, u.ID, p1.ID)
	assert.Equal(t, "Alex River", p1.Name)
	assert.Equal(t, "alexriver", *p1.Username)
	assert.Equal(t, "Software engineer & open source enthusiast", p1.Bio)
	assert.Equal(t, "San Francisco, CA", p1.Location)
	assert.Equal(t, "https://alexriver.dev", p1.Website)
	assert.False(t, p1.IsSelf)
	assert.Empty(t, p1.Email) // Hidden from non-owner

	// 2. Get by username as owner
	p2, err := svc.GetProfile(ctx, u.ID, "alexriver")
	require.NoError(t, err)
	assert.True(t, p2.IsSelf)
	assert.Equal(t, "alex@example.com", p2.Email)

	// 3. Get by ID
	p3, err := svc.GetProfile(ctx, "", u.ID)
	require.NoError(t, err)
	assert.Equal(t, "alexriver", *p3.Username)

	// 4. Get by "me"
	pMe, err := svc.GetProfile(ctx, u.ID, "me")
	require.NoError(t, err)
	assert.True(t, pMe.IsSelf)
	assert.Equal(t, "alexriver", *pMe.Username)

	// 5. "me" unauthenticated -> unauthorized error
	_, err = svc.GetProfile(ctx, "", "me")
	require.Error(t, err)
	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 401, appErr.Status)

	// 6. User not found
	_, err = svc.GetProfile(ctx, "", "nonexistent_handle")
	require.Error(t, err)
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 404, appErr.Status)
}

func TestProfileService_FollowCounts(t *testing.T) {
	svc, client := newTestProfileService(t)
	ctx := context.Background()

	a := createTestUser(t, client, "countera", "countera@example.com")
	b := createTestUser(t, client, "counterb", "counterb@example.com")

	_, err := client.Follow.Create().SetFollowerID(a.ID).SetFollowingID(b.ID).Save(ctx)
	require.NoError(t, err)

	pb, err := svc.GetProfile(ctx, "", "counterb")
	require.NoError(t, err)
	assert.Equal(t, 1, pb.FollowersCount)
	assert.Equal(t, 0, pb.FollowingCount)

	pa, err := svc.GetProfile(ctx, "", "countera")
	require.NoError(t, err)
	assert.Equal(t, 0, pa.FollowersCount)
	assert.Equal(t, 1, pa.FollowingCount)
}

func TestProfileService_UpdateProfile(t *testing.T) {
	svc, client := newTestProfileService(t)
	ctx := context.Background()

	u := createTestUser(t, client, "updater", "updater@example.com")

	newName := "Alex Updated"
	newBio := "New bio description"
	newLoc := "New York, NY"
	newSite := "https://updated.dev"
	newAvatar := "https://images.example.com/new_avatar.png"
	newBanner := "https://images.example.com/new_banner.png"

	updated, err := svc.UpdateProfile(ctx, u.ID, UpdateRequest{
		Name:      &newName,
		Bio:       &newBio,
		Location:  &newLoc,
		Website:   &newSite,
		AvatarURL: &newAvatar,
		BannerURL: &newBanner,
	})
	require.NoError(t, err)
	assert.Equal(t, "Alex Updated", updated.Name)
	assert.Equal(t, "New bio description", updated.Bio)
	assert.Equal(t, "New York, NY", updated.Location)
	assert.Equal(t, "https://updated.dev", updated.Website)
	assert.Equal(t, newAvatar, updated.AvatarURL)
	assert.Equal(t, newBanner, updated.BannerURL)

	// Validation: Bio too long (>280)
	tooLongBio := string(make([]byte, 281))
	_, err = svc.UpdateProfile(ctx, u.ID, UpdateRequest{
		Bio: &tooLongBio,
	})
	require.Error(t, err)
	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Status)

	// Limits count characters, not bytes: 280 emoji (1120 bytes) must save,
	// 281 emoji must be rejected.
	multiByteOK := strings.Repeat("🚀", 280)
	updatedMulti, err := svc.UpdateProfile(ctx, u.ID, UpdateRequest{
		Bio: &multiByteOK,
	})
	require.NoError(t, err)
	assert.Equal(t, multiByteOK, updatedMulti.Bio)

	multiByteTooLong := strings.Repeat("🚀", 281)
	_, err = svc.UpdateProfile(ctx, u.ID, UpdateRequest{
		Bio: &multiByteTooLong,
	})
	require.Error(t, err)
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Status)
}

func TestProfileService_UploadFile_DelegatesToStore(t *testing.T) {
	svc, _ := newTestProfileService(t)
	ctx := context.Background()

	_, err := svc.UploadFile(ctx, "script.sh", []byte("echo hi"))
	require.Error(t, err)
	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Status)

	dummyPNG := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15c4")
	res, err := svc.UploadFile(ctx, "avatar.png", dummyPNG)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Contains(t, res.URL, "/uploads/")
	assert.Equal(t, "avatar.png", res.Name)
}
