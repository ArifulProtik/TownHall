package notification

import (
	"context"
	"testing"

	"ArifulProtik/TownHall/internal/eventbus"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister_FollowCreatedPersists(t *testing.T) {
	svc := newTestService(t)
	bus := eventbus.New()
	Register(bus, svc)
	ctx := context.Background()
	recipient := mustUser(t, svc, "mia", "mia@example.com")
	actorID := mustUser(t, svc, "ned", "ned@example.com")

	err := bus.Publish(ctx, eventbus.FollowCreated{
		RecipientID: recipient,
		Actor:       eventbus.Actor{ID: actorID, Name: "Ned"},
		EntityID:    "edge-r1",
	})
	require.NoError(t, err)

	count, err := svc.UnreadCount(ctx, recipient)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	page, err := svc.List(ctx, recipient, 20, "")
	require.NoError(t, err)
	require.Len(t, page.Notifications, 1)
	assert.Equal(t, TypeFollow, page.Notifications[0].Type)
}
