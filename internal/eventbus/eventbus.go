// Package eventbus carries domain events ("what happened") between
// product areas without domain-to-domain imports. It is in-process only:
// cross-instance fan-out to online SSE subscribers stays in
// notification's Broker. Dispatch is synchronous so tests stay
// deterministic; publishers ignore handler errors so a notification
// failure never fails the action that caused it.
package eventbus

import (
	"context"
	"reflect"
	"sync"
)

// Publisher is what domains depend on. *Bus satisfies it; tests fake it.
// Nil means silent: publishers skip Publish when they hold nil.
type Publisher interface {
	Publish(ctx context.Context, evt any) error
}

// Handler receives one event. Returning an error aborts dispatch of
// that event to later handlers; the publisher decides what to do.
type Handler func(ctx context.Context, evt any) error

// Actor is the snapshot every social event carries about who acted.
type Actor struct {
	ID        string
	Name      string
	Username  string
	AvatarURL string
}

// FollowCreated means actorID followed recipientID (one-way edge).
type FollowCreated struct {
	RecipientID string
	Actor       Actor
	EntityID    string
}

// FriendshipFormed means actorID followed recipientID back: the pair is
// now mutual. Social owns that decision; subscribers only translate.
type FriendshipFormed struct {
	RecipientID string
	Actor       Actor
	EntityID    string
}

type Bus struct {
	mu   sync.RWMutex
	subs map[reflect.Type][]Handler
}

func New() *Bus {
	return &Bus{subs: make(map[reflect.Type][]Handler)}
}

// Subscribe registers h for the concrete event type of prototype
// (e.g. FollowCreated{}). prototype's value is ignored, only its type.
func (b *Bus) Subscribe(prototype any, h Handler) {
	t := reflect.TypeOf(prototype)
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs[t] = append(b.subs[t], h)
}

// Publish dispatches evt to its concrete type's handlers in order.
func (b *Bus) Publish(ctx context.Context, evt any) error {
	t := reflect.TypeOf(evt)
	b.mu.RLock()
	handlers := append([]Handler(nil), b.subs[t]...)
	b.mu.RUnlock()
	for _, h := range handlers {
		if err := h(ctx, evt); err != nil {
			return err
		}
	}
	return nil
}
