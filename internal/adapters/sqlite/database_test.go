package sqlite_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite"
)

const (
	knownSchemaVersion = 5
	lockWaitInTest     = 200 * time.Millisecond
)

func TestOpen(t *testing.T) {
	t.Parallel()

	t.Run("creates a new database without a backup", func(t *testing.T) {
		t.Parallel()

		path := newDatabasePath(t)

		require.NoError(t, openAndClose(t, path))

		assert.Empty(t, backupNames(t, path))
	})

	t.Run("backs up a database with pending migrations before migrating it", func(t *testing.T) {
		t.Parallel()

		path := newDatabasePath(t)
		migrateRawToPreviousSchema(t, path)

		require.NoError(t, openAndClose(t, path))

		assert.Equal(t, []string{"tuisnip-20261001T120000Z-schema4.db"}, backupNames(t, path))
	})

	t.Run("treats a file: prefixed Path as the plain path", func(t *testing.T) {
		t.Parallel()

		path := newDatabasePath(t)
		migrateRawToPreviousSchema(t, path)

		require.NoError(t, openAndClose(t, "file:"+path))

		assert.FileExists(t, path+".lock")
		assert.Equal(t, sqlite.BackupDir(path), sqlite.BackupDir("file:"+path))
		assert.Equal(t, []string{"tuisnip-20261001T120000Z-schema4.db"}, backupNames(t, path))
	})

	t.Run("refuses to migrate when the backup fails", func(t *testing.T) {
		t.Parallel()

		path := newDatabasePath(t)
		migrateRawToPreviousSchema(t, path)
		require.NoError(t, os.WriteFile(sqlite.BackupDir(path), nil, 0o600))

		require.Error(t, openAndClose(t, path))

		require.NoError(t, os.Remove(sqlite.BackupDir(path)))
		require.NoError(t, openAndClose(t, path))
		assert.Equal(t, []string{"tuisnip-20261001T120000Z-schema4.db"}, backupNames(t, path))
	})

	t.Run("keeps the three newest backups and nothing else it did not write", func(t *testing.T) {
		t.Parallel()

		path := newDatabasePath(t)
		migrateRawToPreviousSchema(t, path)
		plantBackupDirFiles(t, path,
			"tuisnip-20260101T090000Z-schema2.db",
			"tuisnip-20260301T090000Z-schema3.db",
			"tuisnip-20260601T090000Z-schema3.db",
			"notes.txt",
		)

		require.NoError(t, openAndClose(t, path))

		assert.ElementsMatch(t, []string{
			"notes.txt",
			"tuisnip-20260301T090000Z-schema3.db",
			"tuisnip-20260601T090000Z-schema3.db",
			"tuisnip-20261001T120000Z-schema4.db",
		}, backupNames(t, path))
	})

	t.Run("skips the backup once the database is up to date", func(t *testing.T) {
		t.Parallel()

		path := newDatabasePath(t)
		migrateRawToPreviousSchema(t, path)
		require.NoError(t, openAndClose(t, path))

		require.NoError(t, openAndClose(t, path))

		assert.Len(t, backupNames(t, path), 1)
	})

	t.Run("waits for another instance's start-up lock until the context ends", func(t *testing.T) {
		t.Parallel()

		path := newDatabasePath(t)
		holdStartupLock(t, path)
		ctx, cancel := context.WithTimeout(t.Context(), lockWaitInTest)
		t.Cleanup(cancel)

		_, err := sqlite.Open(ctx, testOptions(path))

		require.ErrorIs(t, err, context.DeadlineExceeded)
		assert.NoFileExists(t, path)
	})

	t.Run("refuses a database created by a newer build", func(t *testing.T) {
		t.Parallel()

		path := newDatabasePath(t)
		require.NoError(t, openAndClose(t, path))
		execRaw(t, path, "INSERT INTO goose_db_version (version_id, is_applied) VALUES (99, 1)")

		_, err := sqlite.Open(t.Context(), testOptions(path))

		newer, ok := errors.AsType[sqlite.NewerSchemaError](err)
		require.True(t, ok, "want NewerSchemaError, got %v", err)
		assert.Equal(t, sqlite.NewerSchemaError{Database: 99, Known: knownSchemaVersion}, newer)
		assert.Equal(t,
			"The database was created by a newer TuiSnip (schema 99; this build knows 5). Upgrade TuiSnip.",
			newer.Error(),
		)
	})
}
