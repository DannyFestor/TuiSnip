package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqlcgen"
)

type queryFunc func(queries *sqlcgen.Queries) error

func inWriteTransaction(ctx context.Context, db *sql.DB, run queryFunc) error {
	return inTransaction(ctx, db, &sql.TxOptions{Isolation: sql.LevelDefault, ReadOnly: false}, run)
}

func inReadTransaction(ctx context.Context, db *sql.DB, run queryFunc) error {
	return inTransaction(ctx, db, &sql.TxOptions{Isolation: sql.LevelDefault, ReadOnly: true}, run)
}

func inTransaction(ctx context.Context, db *sql.DB, options *sql.TxOptions, run queryFunc) error {
	tx, err := db.BeginTx(ctx, options)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	err = run(sqlcgen.New(tx))
	if err != nil {
		return errors.Join(err, tx.Rollback())
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
