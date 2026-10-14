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

// Emptiness is checked inside the write transaction, so a TUI saving meanwhile can't
// add a Tag between the check and the Import.
func (r *LibraryRepository) InsertIntoEmpty(ctx context.Context, library domain.Library) error {
	err := inWriteTransaction(ctx, r.db, func(queries *sqlcgen.Queries) error {
		emptyErr := requireEmpty(ctx, queries)
		if emptyErr != nil {
			return emptyErr
		}

		return insertLibrary(ctx, queries, library)
	})
	if err != nil {
		return fmt.Errorf("sqlite.LibraryRepository.InsertIntoEmpty: %w", err)
	}

	return nil
}

// The stored Tags are read inside the write transaction, so a Tag a TUI saves
// meanwhile is reused rather than failing the Import on its name.
func (r *LibraryRepository) InsertAlongside(ctx context.Context, library domain.Library) (domain.Library, error) {
	var written domain.Library

	err := inWriteTransaction(ctx, r.db, func(queries *sqlcgen.Queries) error {
		stored, listErr := queries.ListTags(ctx)
		if listErr != nil {
			return fmt.Errorf("list tags: %w", listErr)
		}

		storedTags, rebuildErr := tagsFromRows(stored)
		if rebuildErr != nil {
			return corrupt(rebuildErr)
		}

		written = library.ReusingTags(storedTags)

		return insertLibrary(ctx, queries, written)
	})
	if err != nil {
		return domain.Library{}, fmt.Errorf("sqlite.LibraryRepository.InsertAlongside: %w", err)
	}

	return written, nil
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

func requireEmpty(ctx context.Context, queries *sqlcgen.Queries) error {
	counters := []func(ctx context.Context) (int64, error){
		queries.CountFolders,
		queries.CountTags,
		queries.CountSnippets,
	}

	for _, count := range counters {
		held, err := count(ctx)
		if err != nil {
			return fmt.Errorf("count library records: %w", err)
		}

		if held > 0 {
			return domain.ErrLibraryNotEmpty
		}
	}

	return nil
}

func insertLibrary(ctx context.Context, queries *sqlcgen.Queries, library domain.Library) error {
	folders, err := library.FoldersParentsFirst()
	if err != nil {
		return fmt.Errorf("order folders: %w", err)
	}

	err = insertFolders(ctx, queries, folders)
	if err != nil {
		return err
	}

	err = insertTags(ctx, queries, library.Tags)
	if err != nil {
		return err
	}

	return insertSnippets(ctx, queries, library.Snippets)
}

func insertFolders(ctx context.Context, queries *sqlcgen.Queries, folders []domain.Folder) error {
	for _, folder := range folders {
		err := queries.InsertFolder(ctx, insertFolderParams(folder))
		if err != nil {
			return fmt.Errorf("insert folder: %w", err)
		}
	}

	return nil
}

func insertTags(ctx context.Context, queries *sqlcgen.Queries, tags []domain.Tag) error {
	for _, tag := range tags {
		err := nameClashAsTaken(queries.InsertTag(ctx, insertTagParams(tag)))
		if err != nil {
			return err
		}
	}

	return nil
}

func insertSnippets(ctx context.Context, queries *sqlcgen.Queries, snippets []domain.Snippet) error {
	for i := range snippets {
		err := insertSnippet(ctx, queries, snippets[i])
		if err != nil {
			return err
		}
	}

	return nil
}
