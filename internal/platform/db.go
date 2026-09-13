// Package platform owns external infrastructure clients (Postgres via Ent).
package platform

import (
	"context"
	"database/sql"

	"ArifulProtik/TownHall/ent"
	"ArifulProtik/TownHall/ent/migrate"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	// Blank import registers the pgx stdlib driver under the "pgx" name.
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Open creates an Ent client backed by Postgres via pgx stdlib.
func Open(ctx context.Context, databaseURL string) (*ent.Client, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	drv := entsql.OpenDB(dialect.Postgres, db)
	return ent.NewClient(ent.Driver(drv)), nil
}

// AutoMigrate runs Ent schema creation. Call only in development.
func AutoMigrate(ctx context.Context, client *ent.Client) error {
	return client.Schema.Create(
		ctx,
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	)
}
