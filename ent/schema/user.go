package schema

import (
	"errors"
	"strconv"
	"unicode/utf8"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
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

// runeLimit returns a validator counting characters, not bytes, so values
// meeting the character limit are accepted while overlong values are
// rejected before save.
func runeLimit(n int, what string) func(string) error {
	return func(s string) error {
		if utf8.RuneCountInString(s) > n {
			return errors.New(what + " must not exceed " + strconv.Itoa(n) + " characters")
		}
		return nil
	}
}

// varchar preserves the Postgres column size that MaxLen used to declare,
// without its byte-counting validator.
func varchar(n int) map[string]string {
	return map[string]string{dialect.Postgres: "varchar(" + strconv.Itoa(n) + ")"}
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
		field.String("bio").Optional().SchemaType(varchar(280)).Validate(runeLimit(280, "bio")),
		field.String("avatar_url").Optional().SchemaType(varchar(1000)).Validate(runeLimit(1000, "avatar url")),
		field.String("banner_url").Optional().SchemaType(varchar(1000)).Validate(runeLimit(1000, "banner url")),
		field.String("location").Optional().SchemaType(varchar(100)).Validate(runeLimit(100, "location")),
		field.String("website").Optional().SchemaType(varchar(200)).Validate(runeLimit(200, "website")),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("refresh_tokens", RefreshToken.Type),
	}
}
