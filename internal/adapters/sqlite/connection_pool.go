package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	moderncsqlite "modernc.org/sqlite"
)

func openConnectionPool(path string) (*sql.DB, error) {
	dsn, err := dataSourceName(path)
	if err != nil {
		return nil, err
	}

	connector, err := moderncsqlite.NewConnector(dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	return sql.OpenDB(connector), nil
}

func closeIfOpen(db *sql.DB, cause error) error {
	if db == nil {
		return cause
	}

	return errors.Join(cause, db.Close())
}
