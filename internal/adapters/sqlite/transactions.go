package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqlcgen"
	"github.com/DannyFestor/TuiSnip/internal/domain"
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

func readOne[R any](ctx context.Context, db *sql.DB, query func(queries *sqlcgen.Queries) (R, error)) (R, error) {
	var row R

	err := inReadTransaction(ctx, db, func(queries *sqlcgen.Queries) error {
		var queryErr error

		row, queryErr = query(queries)
		if errors.Is(queryErr, sql.ErrNoRows) {
			return domain.ErrNotFound
		}

		if queryErr != nil {
			return fmt.Errorf("read row: %w", queryErr)
		}

		return nil
	})

	return row, err
}

func findRebuilt[R, E any](
	ctx context.Context,
	db *sql.DB,
	get func(queries *sqlcgen.Queries) (R, error),
	rebuild func(ctx context.Context, row R, corruptLevel slog.Level) (E, error),
) (E, error) {
	row, err := readOne(ctx, db, get)
	if err != nil {
		var none E

		return none, err
	}

	return rebuild(ctx, row, slog.LevelError)
}

func readCount(ctx context.Context, db *sql.DB, query func(queries *sqlcgen.Queries) (int64, error)) (int, error) {
	count, err := readOne(ctx, db, query)

	return int(count), err
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
