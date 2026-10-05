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

func readRows[R any](
	ctx context.Context,
	db *sql.DB,
	query func(queries *sqlcgen.Queries, ctx context.Context) ([]R, error),
) ([]R, error) {
	var rows []R

	err := inReadTransaction(ctx, db, func(queries *sqlcgen.Queries) error {
		var queryErr error

		rows, queryErr = query(queries, ctx)
		if queryErr != nil {
			return fmt.Errorf("read rows: %w", queryErr)
		}

		return nil
	})

	return rows, err
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
