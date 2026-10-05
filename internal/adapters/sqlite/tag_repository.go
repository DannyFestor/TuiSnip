package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	moderncsqlite "modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqlcgen"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type TagRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewTagRepository(database *Database, logger *slog.Logger) *TagRepository {
	return &TagRepository{db: database.db, logger: logger}
}

func (r *TagRepository) Insert(ctx context.Context, tag domain.Tag) error {
	err := inWriteTransaction(ctx, r.db, func(queries *sqlcgen.Queries) error {
		return nameClashAsTaken(queries.InsertTag(ctx, insertTagParams(tag)))
	})
	if err != nil {
		return fmt.Errorf("sqlite.TagRepository.Insert: %w", err)
	}

	return nil
}

func (r *TagRepository) List(ctx context.Context) ([]domain.Tag, error) {
	rows, err := readRows(ctx, r.db, (*sqlcgen.Queries).ListTags)
	if err != nil {
		return nil, fmt.Errorf("sqlite.TagRepository.List: %w", err)
	}

	tags := make([]domain.Tag, 0, len(rows))

	for _, row := range rows {
		tag, rebuildErr := r.rebuild(ctx, row)
		if rebuildErr != nil {
			continue
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

func (r *TagRepository) rebuild(ctx context.Context, row sqlcgen.Tag) (domain.Tag, error) {
	tag, err := tagFromRow(row)
	if err != nil {
		r.logger.WarnContext(
			ctx,
			"tag row is corrupt",
			slog.String(keyTagID, entityID[domain.TagID](row.ID).String()),
			slog.String(keyError, corrupt(err).Error()),
		)
	}

	return tag, err
}

func nameClashAsTaken(err error) error {
	constraint, ok := errors.AsType[*moderncsqlite.Error](err)
	if ok && constraint.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
		return fmt.Errorf("insert tag: %w", domain.ErrTagNameTaken)
	}

	if err != nil {
		return fmt.Errorf("insert tag: %w", err)
	}

	return nil
}
