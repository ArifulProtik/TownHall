// Package platform owns external infrastructure clients (Postgres via Ent).
package platform

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	drv := entsql.OpenDB(dialect.Postgres, db)
	return ent.NewClient(ent.Driver(drv)), nil
}

// WithTx executes fn inside an Ent transaction, committing if fn returns nil or rolling back on error or panic.
func WithTx(ctx context.Context, client *ent.Client, fn func(tx *ent.Tx) error) error {
	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	if err := fn(tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%w: rolling back transaction: %w", err, rerr)
		}
		return err
	}
	return tx.Commit()
}

// AutoMigrate runs Ent schema creation with destructive options.
// Call only in development.
func AutoMigrate(ctx context.Context, client *ent.Client) error {
	return client.Schema.Create(
		ctx,
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	)
}

// Migrate runs Ent schema creation without destructive options.
// Safe for explicit opt-in (e.g. AUTO_MIGRATE=true) outside development.
func Migrate(ctx context.Context, client *ent.Client) error {
	return client.Schema.Create(ctx)
}
