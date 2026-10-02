package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const dataDirPermissions = 0o700

type Database struct {
	db *sql.DB
}

func Open(ctx context.Context, options Options) (*Database, error) {
	options.Path = withoutURIScheme(options.Path)

	err := os.MkdirAll(filepath.Dir(options.Path), dataDirPermissions)
	if err != nil {
		return nil, fmt.Errorf("sqlite.Open: create data directory: %w", err)
	}

	lock, err := acquireStartupLock(ctx, options.Path, options.Logger)
	if err != nil {
		return nil, fmt.Errorf("sqlite.Open: %w", err)
	}

	db, err := openMigrated(ctx, options)

	err = errors.Join(err, lock.release())
	if err != nil {
		return nil, fmt.Errorf("sqlite.Open: %w", closeIfOpen(db, err))
	}

	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	err := d.db.Close()
	if err != nil {
		return fmt.Errorf("sqlite.Database.Close: %w", err)
	}

	return nil
}

func openMigrated(ctx context.Context, options Options) (*sql.DB, error) {
	db, err := openConnectionPool(options.Path)
	if err != nil {
		return nil, err
	}

	err = migrate(ctx, db, options)
	if err != nil {
		return nil, closeIfOpen(db, err)
	}

	return db, nil
}
