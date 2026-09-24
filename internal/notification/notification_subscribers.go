package notification

import (
	"context"

	"ArifulProtik/TownHall/internal/eventbus"
)

// Register subscribes the notification service to the domain events it
// translates into notifications. New notification types add a case here:
// publishers and main.go never change per type.
func Register(bus *eventbus.Bus, svc *Service) {
	bus.Subscribe(eventbus.FollowCreated{}, func(ctx context.Context, evt any) error {
		e := evt.(eventbus.FollowCreated)
		_, err := svc.NotifyFollow(ctx, e.RecipientID, toActor(e.Actor), e.EntityID)
		return err
	})
	bus.Subscribe(eventbus.FriendshipFormed{}, func(ctx context.Context, evt any) error {
		e := evt.(eventbus.FriendshipFormed)
		_, err := svc.NotifyFriend(ctx, e.RecipientID, toActor(e.Actor), e.EntityID)
		return err
	})
}

func toActor(a eventbus.Actor) Actor {
	var username *string
	if a.Username != "" {
		username = &a.Username
	}
	return Actor{ID: a.ID, Name: a.Name, Username: username, AvatarURL: a.AvatarURL}
}
