package social

import (
	"context"
	"fmt"
	"testing"

	"ArifulProtik/TownHall/ent"
	"ArifulProtik/TownHall/ent/enttest"
	"ArifulProtik/TownHall/ent/user"
	"ArifulProtik/TownHall/pkg/apperror"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:socialservice?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	return NewService(client, nil)
}

func createSocialUser(t *testing.T, client *ent.Client, username, email string) *ent.User {
	t.Helper()
	u, err := client.User.Create().
		SetName("Test User").
		SetEmail(email).
		SetPassword("hashedpass123").
		SetUsername(username).
		SetProvider(user.ProviderEmail).
		Save(context.Background())
	require.NoError(t, err)
	return u
}

func TestFollow_MutualBecomesFriends(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	a := createSocialUser(t, svc.db, "a-follow", "a-follow@ex.com")
	b := createSocialUser(t, svc.db, "b-follow", "b-follow@ex.com")

	_, err := svc.Follow(ctx, a.ID, b.ID)
	require.NoError(t, err)

	st, err := svc.Follow(ctx, b.ID, a.ID)
	require.NoError(t, err)
	assert.True(t, st.IsFriend)
	assert.True(t, st.IsFollowing)
	assert.True(t, st.IsFollowedBy)
}

func TestFollow_Guards(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	a := createSocialUser(t, svc.db, "guard-a", "guard-a@ex.com")
	b := createSocialUser(t, svc.db, "guard-b", "guard-b@ex.com")

	// Self-follow is 400.
	_, err := svc.Follow(ctx, a.ID, a.ID)
	require.Error(t, err)
	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Status)

	// Missing target is 404.
	_, err = svc.Follow(ctx, a.ID, "does-not-exist-handle")
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 404, appErr.Status)

	// Duplicate follow is idempotent 200, not an error.
	_, err = svc.Follow(ctx, a.ID, b.ID)
	require.NoError(t, err)
	st, err := svc.Follow(ctx, a.ID, b.ID)
	require.NoError(t, err)
	assert.True(t, st.IsFollowing)
	assert.False(t, st.IsFriend)
	assert.Equal(t, 1, st.FollowersCount)

	// Unfollow of non-edge is idempotent 200.
	st, err = svc.Unfollow(ctx, b.ID, a.ID)
	require.NoError(t, err)
	assert.False(t, st.IsFollowing)

	// Unfollow breaks friendship.
	_, err = svc.Follow(ctx, b.ID, a.ID)
	require.NoError(t, err)
	st, err = svc.Unfollow(ctx, a.ID, b.ID)
	require.NoError(t, err)
	assert.False(t, st.IsFollowing)
	assert.True(t, st.IsFollowedBy)
	assert.False(t, st.IsFriend)
}

func TestFollow_StatusAnonymousAndCounts(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	a := createSocialUser(t, svc.db, "anon-a", "anon-a@ex.com")
	b := createSocialUser(t, svc.db, "anon-b", "anon-b@ex.com")

	_, err := svc.Follow(ctx, a.ID, b.ID)
	require.NoError(t, err)

	st, err := svc.GetStatus(ctx, "", b.ID)
	require.NoError(t, err)
	assert.False(t, st.IsFollowing)
	assert.False(t, st.IsFollowedBy)
	assert.False(t, st.IsFriend)
	assert.False(t, st.IsSelf)
	assert.Equal(t, 1, st.FollowersCount)
	assert.Equal(t, 0, st.FollowingCount)
	assert.Equal(t, 1, svc.FollowersCount(ctx, b.ID))
	assert.Equal(t, 1, svc.FollowingCount(ctx, a.ID))
}

func TestFollow_ListsAndFriends(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	target := createSocialUser(t, svc.db, "list-target", "list-target@ex.com")
	var first []string
	for i := 0; i < 5; i++ {
		u := createSocialUser(t, svc.db, fmt.Sprintf("fan-%d", i), fmt.Sprintf("fan-%d@ex.com", i))
		first = append(first, u.ID)
		_, err := svc.Follow(ctx, u.ID, target.ID)
		require.NoError(t, err)
	}
	// One-sided fan is not a friend; make fan-0 mutual.
	_, err := svc.Follow(ctx, target.ID, first[0])
	require.NoError(t, err)

	page1, err := svc.ListFollowers(ctx, target.ID, 2, "")
	require.NoError(t, err)
	require.Len(t, page1.Users, 2)
	assert.True(t, page1.HasMore)
	require.NotEmpty(t, page1.NextCursor)

	page2, err := svc.ListFollowers(ctx, target.ID, 2, page1.NextCursor)
	require.NoError(t, err)
	require.Len(t, page2.Users, 2)
	assert.NotEqual(t, page1.Users[0].ID, page2.Users[0].ID)

	friends, err := svc.ListFriends(ctx, target.ID, 10, "")
	require.NoError(t, err)
	require.Len(t, friends.Users, 1)
	assert.Equal(t, first[0], friends.Users[0].ID)

	following, err := svc.ListFollowing(ctx, target.ID, 10, "")
	require.NoError(t, err)
	require.Len(t, following.Users, 1)
}

func TestFollow_ListFriendsHasMoreHonest(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	target := createSocialUser(t, svc.db, "friend-target", "friend-target@ex.com")
	mutual := createSocialUser(t, svc.db, "friend-mutual", "friend-mutual@ex.com")

	_, err := svc.Follow(ctx, target.ID, mutual.ID)
	require.NoError(t, err)
	_, err = svc.Follow(ctx, mutual.ID, target.ID)
	require.NoError(t, err)

	// Exact fill with no more edges: must not promise another page.
	p1, err := svc.ListFriends(ctx, target.ID, 1, "")
	require.NoError(t, err)
	require.Len(t, p1.Users, 1)
	assert.False(t, p1.HasMore)
	assert.Empty(t, p1.NextCursor)
}

func TestFollow_CascadeOnUserDelete(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	a := createSocialUser(t, svc.db, "cascade-a", "cascade-a@ex.com")
	b := createSocialUser(t, svc.db, "cascade-b", "cascade-b@ex.com")

	_, err := svc.Follow(ctx, a.ID, b.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, svc.FollowersCount(ctx, b.ID))

	// Deleting the follower removes the edge instead of orphaning it.
	require.NoError(t, svc.db.User.DeleteOneID(a.ID).Exec(ctx))
	assert.Equal(t, 0, svc.FollowersCount(ctx, b.ID))
	assert.Equal(t, 0, svc.FollowingCount(ctx, a.ID))
}

func followBoth(t *testing.T, svc *Service, ctx context.Context, a, b *ent.User) {
	t.Helper()
	_, err := svc.Follow(ctx, a.ID, b.ID)
	require.NoError(t, err)
	_, err = svc.Follow(ctx, b.ID, a.ID)
	require.NoError(t, err)
}

func TestFollow_ListFriendsPaginatesAcrossPages(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	target := createSocialUser(t, svc.db, "page-target", "page-target@ex.com")
	m1 := createSocialUser(t, svc.db, "page-m1", "page-m1@ex.com")
	m2 := createSocialUser(t, svc.db, "page-m2", "page-m2@ex.com")
	m3 := createSocialUser(t, svc.db, "page-m3", "page-m3@ex.com")
	followBoth(t, svc, ctx, target, m1)
	followBoth(t, svc, ctx, target, m2)
	followBoth(t, svc, ctx, target, m3)

	p1, err := svc.ListFriends(ctx, target.ID, 2, "")
	require.NoError(t, err)
	require.Len(t, p1.Users, 2)
	assert.True(t, p1.HasMore)
	require.NotEmpty(t, p1.NextCursor)

	p2, err := svc.ListFriends(ctx, target.ID, 2, p1.NextCursor)
	require.NoError(t, err)
	require.Len(t, p2.Users, 1)
	assert.Equal(t, m3.ID, p2.Users[0].ID)
	assert.False(t, p2.HasMore)
}

func TestFollow_ListFriendsLosesNothingAfterOneSided(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	target := createSocialUser(t, svc.db, "skip-target", "skip-target@ex.com")
	// One-sided follow first so the first edge page is sparse.
	s1 := createSocialUser(t, svc.db, "skip-s1", "skip-s1@ex.com")
	_, err := svc.Follow(ctx, target.ID, s1.ID)
	require.NoError(t, err)
	m1 := createSocialUser(t, svc.db, "skip-m1", "skip-m1@ex.com")
	m2 := createSocialUser(t, svc.db, "skip-m2", "skip-m2@ex.com")
	m3 := createSocialUser(t, svc.db, "skip-m3", "skip-m3@ex.com")
	followBoth(t, svc, ctx, target, m1)
	followBoth(t, svc, ctx, target, m2)
	followBoth(t, svc, ctx, target, m3)

	p1, err := svc.ListFriends(ctx, target.ID, 2, "")
	require.NoError(t, err)
	require.Len(t, p1.Users, 2)
	assert.Equal(t, m1.ID, p1.Users[0].ID)
	assert.Equal(t, m2.ID, p1.Users[1].ID)
	assert.True(t, p1.HasMore)

	p2, err := svc.ListFriends(ctx, target.ID, 2, p1.NextCursor)
	require.NoError(t, err)
	require.Len(t, p2.Users, 1)
	assert.Equal(t, m3.ID, p2.Users[0].ID)
	assert.False(t, p2.HasMore)
}

type recordingNotifier struct {
	follows []string
	friends []string
}

func (f *recordingNotifier) NotifyFollow(_ context.Context, recipientID, _, _, _, _, _ string) error {
	f.follows = append(f.follows, recipientID)
	return nil
}

func (f *recordingNotifier) NotifyFriend(_ context.Context, recipientID, _, _, _, _, _ string) error {
	f.friends = append(f.friends, recipientID)
	return nil
}

func TestFollow_FollowBackNotifiesFriend(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:socialfriend?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	notifier := &recordingNotifier{}
	svc := NewService(client, notifier)
	ctx := context.Background()

	alice := createSocialUser(t, client, "alice", "alice@example.com")
	bob := createSocialUser(t, client, "bob", "bob@example.com")

	_, err := svc.Follow(ctx, alice.ID, bob.Username)
	require.NoError(t, err)
	assert.Equal(t, []string{bob.ID}, notifier.follows)
	assert.Empty(t, notifier.friends)

	_, err = svc.Follow(ctx, bob.ID, alice.Username)
	require.NoError(t, err)
	assert.Equal(t, []string{alice.ID}, notifier.friends)
	assert.Len(t, notifier.follows, 1)
}
