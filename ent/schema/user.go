package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Mixin of the User.
func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("email").NotEmpty().Unique(),
		field.String("password").NotEmpty(),
		field.String("username").Optional().Unique(),
		field.Enum("provider").Values("google", "github", "email").Default("google"),
		field.Bool("email_verified").Default(false),
		field.String("bio").Optional().MaxLen(280),
		field.String("avatar_url").Optional().MaxLen(1000),
		field.String("banner_url").Optional().MaxLen(1000),
		field.String("location").Optional().MaxLen(100),
		field.String("website").Optional().MaxLen(200),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("refresh_tokens", RefreshToken.Type),
	}
}
