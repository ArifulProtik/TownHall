package eventbus

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBus_DispatchesByConcreteType(t *testing.T) {
	bus := New()
	var follows []FollowCreated
	var friends []FriendshipFormed
	bus.Subscribe(FollowCreated{}, func(_ context.Context, evt any) error {
		e, ok := evt.(FollowCreated)
		require.True(t, ok)
		follows = append(follows, e)
		return nil
	})
	bus.Subscribe(FriendshipFormed{}, func(_ context.Context, evt any) error {
		e, ok := evt.(FriendshipFormed)
		require.True(t, ok)
		friends = append(friends, e)
		return nil
	})

	require.NoError(t, bus.Publish(context.Background(), FollowCreated{RecipientID: "u1"}))
	require.NoError(t, bus.Publish(context.Background(), FriendshipFormed{RecipientID: "u2"}))

	require.Len(t, follows, 1)
	assert.Equal(t, "u1", follows[0].RecipientID)
	require.Len(t, friends, 1)
	assert.Equal(t, "u2", friends[0].RecipientID)
}

func TestBus_NilPublisherIsSilent(t *testing.T) {
	var p Publisher
	assert.Nil(t, p)
}
