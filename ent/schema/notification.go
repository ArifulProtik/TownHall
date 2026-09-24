package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Notification is one persisted event for recipient_id. Actor snapshot lives
// in data JSON so the bell renders without joins. read_at NULL = unread.
// Type enum is the extensibility point: follow now, like/comment later.
type Notification struct {
	ent.Schema
}

func (Notification) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (Notification) Fields() []ent.Field {
	return []ent.Field{
		field.String("recipient_id").NotEmpty(),
		field.String("actor_id").NotEmpty(),
		field.Enum("type").Values("follow", "like", "comment", "mention", "friend").Default("follow"),
		field.String("entity_type").NotEmpty(),
		field.String("entity_id").NotEmpty(),
		field.Text("data").Optional().SchemaType(map[string]string{
			dialect.Postgres: "jsonb",
		}),
		field.Time("read_at").Optional().Nillable(),
	}
}

func (Notification) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("recipient_id", "created_at"),
		index.Fields("recipient_id", "read_at"),
		index.Fields("recipient_id", "type", "entity_id").Unique(),
	}
}
