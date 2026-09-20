package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// RefreshToken holds a hashed opaque refresh token for rotation and revocation.
type RefreshToken struct {
	ent.Schema
}

// Mixin of the RefreshToken.
func (RefreshToken) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

// Fields of the RefreshToken.
func (RefreshToken) Fields() []ent.Field {
	return []ent.Field{
		field.String("token_hash").NotEmpty().Unique(),
		field.Time("expires_at"),
		field.Time("revoked_at").Optional().Nillable(),
		field.String("replaced_by").Optional(),
	}
}

// Edges of the RefreshToken.
func (RefreshToken) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("refresh_tokens").Unique().Required(),
	}
}
