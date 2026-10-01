package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/pressly/goose/v3"

	"github.com/DannyFestor/TuiSnip/db/migrations"
)

type schemaMigration struct {
	db       *sql.DB
	migrator *goose.Provider
	options  Options
}

func migrate(ctx context.Context, db *sql.DB, options Options) error {
	migrator, err := goose.NewProvider(goose.DialectSQLite3, db, migrations.FS(), goose.WithLogger(goose.NopLogger()))
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	return schemaMigration{db: db, migrator: migrator, options: options}.run(ctx)
}

func (m schemaMigration) run(ctx context.Context) error {
	current, err := m.supportedVersion(ctx)
	if err != nil {
		return err
	}

	pending, err := m.migrator.HasPending(ctx)
	if err != nil {
		return fmt.Errorf("check pending migrations: %w", err)
	}

	if !pending {
		return nil
	}

	return m.upgrade(ctx, current)
}

func (m schemaMigration) supportedVersion(ctx context.Context) (int64, error) {
	current, err := m.migrator.GetDBVersion(ctx)
	if err != nil {
		return 0, fmt.Errorf("read schema version: %w", err)
	}

	known := m.newestKnownVersion()
	if current > known {
		return 0, NewerSchemaError{Database: current, Known: known}
	}

	return current, nil
}

func (m schemaMigration) newestKnownVersion() int64 {
	sources := m.migrator.ListSources()
	if len(sources) == 0 {
		return 0
	}

	return sources[len(sources)-1].Version
}

func (m schemaMigration) upgrade(ctx context.Context, current int64) error {
	if current > 0 {
		err := m.backUp(ctx, current)
		if err != nil {
			return err
		}
	}

	_, err := m.migrator.Up(ctx)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	m.options.Logger.InfoContext(ctx, "migrations applied",
		slog.Int64(keyFromVersion, current), slog.Int64(keyToVersion, m.newestKnownVersion()))

	return nil
}

func (m schemaMigration) backUp(ctx context.Context, current int64) error {
	backupPath, err := backUp(ctx, m.db, m.options, current)
	if err != nil {
		return err
	}

	m.options.Logger.InfoContext(ctx, "database backed up", slog.String(keyBackupPath, backupPath))

	return nil
}
