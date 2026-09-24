package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Follow is one directed follow edge follower_id -> following_id.
// Friendship is derived (mutual edges), never stored.
// Reserved: add `status` (pending/accepted) for private-account approvals.
type Follow struct {
	ent.Schema
}

func (Follow) Mixin() []ent.Mixin {
	return []ent.Mixin{BaseMixin{}}
}

func (Follow) Fields() []ent.Field {
	return []ent.Field{
		field.String("follower_id").NotEmpty(),
		field.String("following_id").NotEmpty(),
	}
}

func (Follow) Edges() []ent.Edge { return nil }

func (Follow) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("follower_id", "following_id").Unique(),
		index.Fields("follower_id"),
		index.Fields("following_id"),
	}
}
