package notification

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	"ArifulProtik/TownHall/ent"
	"ArifulProtik/TownHall/ent/notification"
	"ArifulProtik/TownHall/ent/user"
	"ArifulProtik/TownHall/pkg/apperror"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

// Service owns persistence + fan-out. Handlers depend on the ServiceAPI
// interface (see handler), never this concrete, so tests mock it.
type Service struct {
	db     *ent.Client
	broker Broker
}

func NewService(db *ent.Client, broker Broker) *Service {
	if broker == nil {
		broker = NewMemoryBroker()
	}
	return &Service{db: db, broker: broker}
}

// Subscribe exposes live fan-out to the SSE handler without leaking the
// broker itself.
func (s *Service) Subscribe(userID string) (<-chan Event, func()) {
	return s.broker.Subscribe(userID)
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return defaultListLimit
	}
	if limit > maxListLimit {
		return maxListLimit
	}
	return limit
}

func encodeCursor(id string) string { return base64.URLEncoding.EncodeToString([]byte(id)) }

func decodeCursor(cursor string) (string, bool) {
	if cursor == "" {
		return "", false
	}
	raw, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil || len(raw) == 0 {
		return "", false
	}
	return string(raw), true
}

func toItem(n *ent.Notification, actor Actor) Item {
	return Item{
		ID: n.ID, Type: string(n.Type), Actor: actor,
		EntityType: n.EntityType, EntityID: n.EntityID,
		Data: n.Data, ReadAt: n.ReadAt, CreatedAt: n.CreatedAt,
	}
}

// NotifyFollow persists the follow event (idempotent on retry via the
// unique recipient/type/entity key) then fans out to live subscribers.
// Publish failures never fail the follow itself.
func (s *Service) NotifyFollow(ctx context.Context, recipientID string, actor Actor, entityID string) (*ent.Notification, error) {
	return s.persistAndPublish(ctx, recipientID, actor, TypeFollow, "follow", entityID)
}

// NotifyFriend persists the follow-back event: the recipient is told the
// actor followed them back (they are now friends). Same idempotency and
// fan-out contract as NotifyFollow.
func (s *Service) NotifyFriend(ctx context.Context, recipientID string, actor Actor, entityID string) (*ent.Notification, error) {
	return s.persistAndPublish(ctx, recipientID, actor, TypeFriend, "friend", entityID)
}

func (s *Service) persistAndPublish(ctx context.Context, recipientID string, actor Actor, notifType, entityType, entityID string) (*ent.Notification, error) {
	// Data is jsonb in Postgres: a bare name is not valid JSON and the
	// insert fails there (sqlite tests accept anything). Snapshot the
	// actor as a JSON object so the bell renders without joins.
	snapshot, err := json.Marshal(actor)
	if err != nil {
		return nil, apperror.Internal()
	}
	n, err := s.db.Notification.Create().
		SetRecipientID(recipientID).
		SetActorID(actor.ID).
		SetType(notification.Type(notifType)).
		SetEntityType(entityType).
		SetEntityID(entityID).
		SetData(string(snapshot)).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			existing, qerr := s.db.Notification.Query().
				Where(notification.RecipientID(recipientID),
					notification.TypeEQ(notification.Type(notifType)),
					notification.EntityIDEQ(entityID)).
				Only(ctx)
			if qerr != nil {
				return nil, apperror.Internal()
			}
			return existing, nil
		}
		return nil, apperror.Internal()
	}
	_ = s.broker.Publish(ctx, Event{
		ID: n.ID, Type: notifType, Recipient: recipientID, ActorID: actor.ID,
		EntityType: entityType, EntityID: entityID, Data: n.Data, CreatedAt: n.CreatedAt,
	})
	return n, nil
}

func (s *Service) loadActors(ctx context.Context, ids []string) map[string]Actor {
	out := make(map[string]Actor, len(ids))
	if len(ids) == 0 {
		return out
	}
	users, err := s.db.User.Query().Where(user.IDIn(ids...)).All(ctx)
	if err != nil {
		return out
	}
	for _, u := range users {
		var username *string
		if u.Username != "" {
			username = &u.Username
		}
		out[u.ID] = Actor{ID: u.ID, Name: u.Name, Username: username, AvatarURL: u.AvatarURL}
	}
	for _, id := range ids {
		if _, ok := out[id]; !ok {
			out[id] = Actor{ID: id}
		}
	}
	return out
}

func (s *Service) List(ctx context.Context, userID string, limit int, cursor string) (*ListResponse, error) {
	limit = clampLimit(limit)
	q := s.db.Notification.Query().Where(notification.RecipientID(userID))
	if id, ok := decodeCursor(cursor); ok {
		q = q.Where(notification.IDLT(id))
	}
	// Newest first (UUIDv7 ids sort as creation order): the bell shows
	// fresh notifications on top, paging walks into history.
	rows, err := q.Order(ent.Desc(notification.FieldID)).Limit(limit + 1).All(ctx)
	if err != nil {
		return nil, apperror.Internal()
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	actorIDs := make([]string, 0, len(rows))
	for _, r := range rows {
		actorIDs = append(actorIDs, r.ActorID)
	}
	actors := s.loadActors(ctx, actorIDs)
	items := make([]Item, 0, len(rows))
	for _, r := range rows {
		items = append(items, toItem(r, actors[r.ActorID]))
	}
	resp := &ListResponse{Notifications: items, HasMore: hasMore}
	if hasMore && len(rows) > 0 {
		resp.NextCursor = encodeCursor(rows[len(rows)-1].ID)
	}
	return resp, nil
}

// ListSince returns notifications newer than lastID, oldest first,
// bounded by limit. SSE replay uses it: missed events resend in order.
// UUIDv7 ids sort as creation order, so IDGT(lastID) is "everything
// after", and ascending keeps the replay chronological.
func (s *Service) ListSince(ctx context.Context, userID, lastID string, limit int) ([]Item, error) {
	limit = clampLimit(limit)
	rows, err := s.db.Notification.Query().
		Where(notification.RecipientID(userID), notification.IDGT(lastID)).
		Order(ent.Asc(notification.FieldID)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, apperror.Internal()
	}
	actorIDs := make([]string, 0, len(rows))
	for _, r := range rows {
		actorIDs = append(actorIDs, r.ActorID)
	}
	actors := s.loadActors(ctx, actorIDs)
	items := make([]Item, 0, len(rows))
	for _, r := range rows {
		items = append(items, toItem(r, actors[r.ActorID]))
	}
	return items, nil
}

func (s *Service) UnreadCount(ctx context.Context, userID string) (int, error) {
	n, err := s.db.Notification.Query().
		Where(notification.RecipientID(userID), notification.ReadAtIsNil()).
		Count(ctx)
	if err != nil {
		return 0, apperror.Internal()
	}
	return n, nil
}

func (s *Service) MarkRead(ctx context.Context, userID, id string) error {
	n, err := s.db.Notification.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperror.NotFound("notification not found")
		}
		return apperror.Internal()
	}
	if n.RecipientID != userID {
		return apperror.NotFound("notification not found")
	}
	if n.ReadAt != nil {
		return nil
	}
	if _, err := n.Update().SetReadAt(time.Now()).Save(ctx); err != nil {
		return apperror.Internal()
	}
	return nil
}

func (s *Service) MarkAllRead(ctx context.Context, userID string) error {
	if _, err := s.db.Notification.Update().
		Where(notification.RecipientID(userID), notification.ReadAtIsNil()).
		SetReadAt(time.Now()).
		Save(ctx); err != nil {
		return apperror.Internal()
	}
	return nil
}
