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
		tag, rebuildErr := r.rebuild(ctx, row, slog.LevelWarn)
		if rebuildErr != nil {
			continue
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

func (r *TagRepository) Find(ctx context.Context, id domain.TagID) (domain.Tag, error) {
	get := func(queries *sqlcgen.Queries) (sqlcgen.Tag, error) { return queries.GetTag(ctx, columnID(id)) }

	tag, err := findRebuilt(ctx, r.db, get, r.rebuild)
	if err != nil {
		return domain.Tag{}, fmt.Errorf("sqlite.TagRepository.Find: %w", err)
	}

	return tag, nil
}

func (r *TagRepository) Rename(ctx context.Context, tag domain.Tag) (domain.Tag, error) {
	var survivor domain.Tag

	err := inWriteTransaction(ctx, r.db, func(queries *sqlcgen.Queries) error {
		var renameErr error

		survivor, renameErr = r.renamed(ctx, queries, tag)

		return renameErr
	})
	if err != nil {
		return domain.Tag{}, fmt.Errorf("sqlite.TagRepository.Rename: %w", err)
	}

	return survivor, nil
}

func (r *TagRepository) Delete(ctx context.Context, id domain.TagID) error {
	err := inWriteTransaction(ctx, r.db, func(queries *sqlcgen.Queries) error {
		return deleteTag(ctx, queries, id)
	})
	if err != nil {
		return fmt.Errorf("sqlite.TagRepository.Delete: %w", err)
	}

	return nil
}

func (r *TagRepository) CountSnippets(ctx context.Context, id domain.TagID) (int, error) {
	count, err := readCount(ctx, r.db, func(queries *sqlcgen.Queries) (int64, error) {
		return queries.CountSnippetsWithTag(ctx, columnID(id))
	})
	if err != nil {
		return 0, fmt.Errorf("sqlite.TagRepository.CountSnippets: %w", err)
	}

	return count, nil
}

func (r *TagRepository) renamed(ctx context.Context, queries *sqlcgen.Queries, tag domain.Tag) (domain.Tag, error) {
	clashing, err := queries.ListOtherTagsWithNameKey(ctx, otherTagsWithNameKeyParams(tag))
	if err != nil {
		return domain.Tag{}, fmt.Errorf("list tags with the name: %w", err)
	}

	if len(clashing) == 0 {
		return tag, updateTag(ctx, queries, tag)
	}

	return r.merged(ctx, queries, tag, clashing[0])
}

func (r *TagRepository) merged(
	ctx context.Context,
	queries *sqlcgen.Queries,
	renamed domain.Tag,
	existingRow sqlcgen.Tag,
) (domain.Tag, error) {
	existing, err := r.rebuild(ctx, existingRow, slog.LevelError)
	if err != nil {
		return domain.Tag{}, err
	}

	survivor, err := existing.Rename(renamed.Name(), renamed.UpdatedAt())
	if err != nil {
		return domain.Tag{}, fmt.Errorf("merge tag: %w", err)
	}

	err = updateTag(ctx, queries, survivor)
	if err != nil {
		return domain.Tag{}, err
	}

	err = queries.MoveSnippetTags(ctx, moveSnippetTagsParams(renamed.ID(), survivor.ID()))
	if err != nil {
		return domain.Tag{}, fmt.Errorf("move snippet tags: %w", err)
	}

	return survivor, deleteTag(ctx, queries, renamed.ID())
}

func (r *TagRepository) rebuild(ctx context.Context, row sqlcgen.Tag, corruptLevel slog.Level) (domain.Tag, error) {
	tag, err := tagFromRow(row)
	if err != nil {
		err = corrupt(err)
		r.logger.LogAttrs(
			ctx,
			corruptLevel,
			"tag row is corrupt",
			slog.String(keyTagID, entityID[domain.TagID](row.ID).String()),
			slog.String(keyError, err.Error()),
		)
	}

	return tag, err
}

func updateTag(ctx context.Context, queries *sqlcgen.Queries, tag domain.Tag) error {
	updated, err := queries.UpdateTag(ctx, updateTagParams(tag))
	if err != nil {
		return fmt.Errorf("update tag: %w", err)
	}

	if updated == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func deleteTag(ctx context.Context, queries *sqlcgen.Queries, id domain.TagID) error {
	deleted, err := queries.DeleteTag(ctx, columnID(id))
	if err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}

	if deleted == 0 {
		return domain.ErrNotFound
	}

	return nil
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
