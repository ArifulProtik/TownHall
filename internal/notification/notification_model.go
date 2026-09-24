package notification

import "time"

// Types are the extensibility point: follow ships now, like/comment and
// others only add a value plus a publisher call — no broker change.
const (
	TypeFollow  = "follow"
	TypeLike    = "like"
	TypeComment = "comment"
	TypeMention = "mention"
	TypeFriend  = "friend"
)

// Event is the fan-out payload. ID is the ent Notification id so clients
// can resume with Last-Event-ID.
type Event struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Recipient  string    `json:"recipient_id"`
	ActorID    string    `json:"actor_id"`
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	Data       string    `json:"data,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type Actor struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Username  *string `json:"username,omitempty"`
	AvatarURL string  `json:"avatar_url,omitempty"`
}

type Item struct {
	ID         string     `json:"id"`
	Type       string     `json:"type"`
	Actor      Actor      `json:"actor"`
	EntityType string     `json:"entity_type"`
	EntityID   string     `json:"entity_id"`
	Data       string     `json:"data,omitempty"`
	ReadAt     *time.Time `json:"read_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type ListResponse struct {
	Notifications []Item `json:"notifications"`
	NextCursor    string `json:"next_cursor,omitempty"`
	HasMore       bool   `json:"has_more"`
}

type UnreadResponse struct {
	Count int `json:"count"`
}
