package platform_test

import (
	"context"
	"errors"
	"testing"

	"ArifulProtik/TownHall/ent"
	"ArifulProtik/TownHall/ent/enttest"
	"ArifulProtik/TownHall/ent/user"
	"ArifulProtik/TownHall/internal/platform"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithTx_CommitOnSuccess(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:withtx_commit?mode=memory&cache=shared&_fk=1")
	defer client.Close()
	ctx := context.Background()

	err := platform.WithTx(ctx, client, func(tx *ent.Tx) error {
		_, err := tx.User.Create().
			SetName("Tx User").
			SetEmail("tx@example.com").
			SetPassword("hash").
			Save(ctx)
		return err
	})
	require.NoError(t, err)

	exists, err := client.User.Query().Where(user.EmailEQ("tx@example.com")).Exist(ctx)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestWithTx_RollbackOnError(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:withtx_rollback?mode=memory&cache=shared&_fk=1")
	defer client.Close()
	ctx := context.Background()

	expectedErr := errors.New("aborted")
	err := platform.WithTx(ctx, client, func(tx *ent.Tx) error {
		_, err := tx.User.Create().
			SetName("Tx User 2").
			SetEmail("tx2@example.com").
			SetPassword("hash").
			Save(ctx)
		if err != nil {
			return err
		}
		return expectedErr
	})
	require.ErrorIs(t, err, expectedErr)

	exists, err := client.User.Query().Where(user.EmailEQ("tx2@example.com")).Exist(ctx)
	require.NoError(t, err)
	assert.False(t, exists)
}
