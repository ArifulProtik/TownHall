// Package platform owns outside-world clients (Postgres via Ent).
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

// WithTx runs fn in a transaction. Auth Refresh rolls its own instead:
// it commits revoked-token cleanup on paths that still answer 401.
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

// AutoMigrate is destructive. Development only.
func AutoMigrate(ctx context.Context, client *ent.Client) error {
	return client.Schema.Create(
		ctx,
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	)
}

// Migrate creates schema without dropping anything (AUTO_MIGRATE=true).
func Migrate(ctx context.Context, client *ent.Client) error {
	return client.Schema.Create(ctx)
}
