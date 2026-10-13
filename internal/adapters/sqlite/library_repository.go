package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqlcgen"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type LibraryRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewLibraryRepository(database *Database, logger *slog.Logger) *LibraryRepository {
	return &LibraryRepository{db: database.db, logger: logger}
}

// The Library is read in one transaction, so a TUI writing meanwhile can't leave a
// Snippet pointing at a Folder the read missed. A corrupt row fails the whole load:
// skipping it, as Browse does, would leave a Library Import can't restore.
func (r *LibraryRepository) Load(ctx context.Context) (domain.Library, error) {
	rows, err := r.readRows(ctx)
	if err != nil {
		return domain.Library{}, fmt.Errorf("sqlite.LibraryRepository.Load: %w", err)
	}

	library, err := rows.rebuild(ctx, r.logger)
	if err != nil {
		return domain.Library{}, fmt.Errorf("sqlite.LibraryRepository.Load: %w", err)
	}

	return library, nil
}

func (r *LibraryRepository) readRows(ctx context.Context) (libraryRows, error) {
	var rows libraryRows

	err := inReadTransaction(ctx, r.db, func(queries *sqlcgen.Queries) error {
		var selectErr error

		rows, selectErr = selectLibraryRows(ctx, queries)

		return selectErr
	})

	return rows, err
}
