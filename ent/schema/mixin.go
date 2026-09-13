// Package schema defines the Ent entity schemas.
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

// BaseMixin adds a UUIDv7 string id plus timestamptz timestamps to schemas.
type BaseMixin struct {
	mixin.Schema
}

// Fields of the BaseMixin.
func (BaseMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").NotEmpty().DefaultFunc(func() string {
			return uuid.Must(uuid.NewV7()).String()
		}),
		field.Time("created_at").Default(time.Now).SchemaType(map[string]string{
			dialect.Postgres: "timestamptz",
		}),
		field.Time("updated_at").Default(time.Now).SchemaType(map[string]string{
			dialect.Postgres: "timestamptz",
		}).UpdateDefault(time.Now),
	}
}
