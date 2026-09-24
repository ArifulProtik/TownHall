package notification

import (
	"context"
	"testing"
	"time"

	"ArifulProtik/TownHall/ent/enttest"
	"ArifulProtik/TownHall/ent/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:notifservice?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	return NewService(client, NewMemoryBroker())
}

func mustUser(t *testing.T, svc *Service, username, email string) string {
	t.Helper()
	u, err := svc.db.User.Create().
		SetName("Test").
		SetEmail(email).
		SetPassword("x").
		SetUsername(username).
		SetProvider(user.ProviderEmail).
		Save(context.Background())
	require.NoError(t, err)
	return u.ID
}

func TestNotifyFollow_PersistsAndFansOut(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	recipient := mustUser(t, svc, "bob", "bob@example.com")
	actorID := mustUser(t, svc, "alice", "alice@example.com")

	ch, unsub := svc.Subscribe(recipient)
	defer unsub()

	n, err := svc.NotifyFollow(ctx, recipient, Actor{ID: actorID, Name: "Alice"}, "edge-1")
	require.NoError(t, err)
	assert.Equal(t, recipient, n.RecipientID)

	select {
	case e := <-ch:
		assert.Equal(t, n.ID, e.ID)
		assert.Equal(t, TypeFollow, e.Type)
	default:
		t.Fatal("expected live event")
	}

	count, err := svc.UnreadCount(ctx, recipient)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestNotifyFollow_IdempotentOnRetry(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	recipient := mustUser(t, svc, "bob2", "bob2@example.com")
	actorID := mustUser(t, svc, "alice2", "alice2@example.com")

	first, err := svc.NotifyFollow(ctx, recipient, Actor{ID: actorID, Name: "Alice"}, "edge-9")
	require.NoError(t, err)
	second, err := svc.NotifyFollow(ctx, recipient, Actor{ID: actorID, Name: "Alice"}, "edge-9")
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID)
}

func TestList_MarkRead(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	recipient := mustUser(t, svc, "carol", "carol@example.com")
	actorID := mustUser(t, svc, "dave", "dave@example.com")

	n, err := svc.NotifyFollow(ctx, recipient, Actor{ID: actorID, Name: "Dave"}, "edge-2")
	require.NoError(t, err)

	page, err := svc.List(ctx, recipient, 20, "")
	require.NoError(t, err)
	require.Len(t, page.Notifications, 1)
	assert.Equal(t, "dave", *page.Notifications[0].Actor.Username)

	require.NoError(t, svc.MarkRead(ctx, recipient, n.ID))
	count, err := svc.UnreadCount(ctx, recipient)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestMarkAllRead(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	recipient := mustUser(t, svc, "erin", "erin@example.com")
	actorID := mustUser(t, svc, "frank", "frank@example.com")

	_, err := svc.NotifyFollow(ctx, recipient, Actor{ID: actorID, Name: "Frank"}, "e1")
	require.NoError(t, err)
	_, err = svc.NotifyFollow(ctx, recipient, Actor{ID: actorID, Name: "Frank"}, "e2")
	require.NoError(t, err)

	require.NoError(t, svc.MarkAllRead(ctx, recipient))
	count, err := svc.UnreadCount(ctx, recipient)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestMemoryBroker_PublishSubscribe(t *testing.T) {
	b := NewMemoryBroker()
	ch, unsub := b.Subscribe("u1")
	defer unsub()
	require.NoError(t, b.Publish(context.Background(), Event{ID: "1", Recipient: "u1"}))
	e := <-ch
	assert.Equal(t, "1", e.ID)
}

func TestNotifyFriend_PersistsFriendType(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	recipient := mustUser(t, svc, "gina", "gina@example.com")
	actorID := mustUser(t, svc, "hank", "hank@example.com")

	ch, unsub := svc.Subscribe(recipient)
	defer unsub()

	n, err := svc.NotifyFriend(ctx, recipient, Actor{ID: actorID, Name: "Hank"}, "edge-f1")
	require.NoError(t, err)
	assert.Equal(t, TypeFriend, string(n.Type))

	select {
	case e := <-ch:
		assert.Equal(t, TypeFriend, e.Type)
		assert.Equal(t, n.ID, e.ID)
	default:
		t.Fatal("expected live friend event")
	}
}

func TestList_NewestFirst(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	recipient := mustUser(t, svc, "ivan", "ivan@example.com")
	actorID := mustUser(t, svc, "judy", "judy@example.com")

	// Distinct ids need distinct timestamps: UUIDv7 embeds time.
	_, err := svc.NotifyFollow(ctx, recipient, Actor{ID: actorID, Name: "Judy"}, "e-old")
	require.NoError(t, err)
	time.Sleep(2 * time.Millisecond)
	_, err = svc.NotifyFollow(ctx, recipient, Actor{ID: actorID, Name: "Judy"}, "e-new")
	require.NoError(t, err)

	page, err := svc.List(ctx, recipient, 20, "")
	require.NoError(t, err)
	require.Len(t, page.Notifications, 2)
	assert.Equal(t, "e-new", page.Notifications[0].EntityID)
	assert.Equal(t, "e-old", page.Notifications[1].EntityID)
}
