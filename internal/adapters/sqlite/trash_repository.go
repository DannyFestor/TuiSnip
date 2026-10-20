package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqlcgen"
	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqltype"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type TrashRepository struct {
	db       *sql.DB
	logger   *slog.Logger
	snippets *SnippetRepository
}

func NewTrashRepository(database *Database, logger *slog.Logger) *TrashRepository {
	return &TrashRepository{db: database.db, logger: logger, snippets: NewSnippetRepository(database, logger)}
}

func (r *TrashRepository) CountItems(ctx context.Context) (int, error) {
	var count int

	err := inReadTransaction(ctx, r.db, func(queries *sqlcgen.Queries) error {
		folders, foldersErr := queries.CountFolderTrashItems(ctx)
		if foldersErr != nil {
			return fmt.Errorf("count folder trash items: %w", foldersErr)
		}

		snippets, snippetsErr := queries.CountSnippetTrashItems(ctx)
		if snippetsErr != nil {
			return fmt.Errorf("count snippet trash items: %w", snippetsErr)
		}

		count = int(folders + snippets)

		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("sqlite.TrashRepository.CountItems: %w", err)
	}

	return count, nil
}

func (r *TrashRepository) Load(ctx context.Context) (domain.Library, error) {
	var rows libraryRows

	err := inReadTransaction(ctx, r.db, func(queries *sqlcgen.Queries) error {
		var selectErr error

		rows, selectErr = selectTrashRows(ctx, queries)

		return selectErr
	})
	if err != nil {
		return domain.Library{}, fmt.Errorf("sqlite.TrashRepository.Load: %w", err)
	}

	library, err := rows.rebuild(ctx, r.logger)
	if err != nil {
		return domain.Library{}, fmt.Errorf("sqlite.TrashRepository.Load: %w", err)
	}

	return library, nil
}

func (r *TrashRepository) Find(ctx context.Context, id domain.SnippetID) (domain.Snippet, error) {
	snippet, err := r.snippets.loadOne(ctx, snippetTrashItem(id))
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("sqlite.TrashRepository.Find: %w", err)
	}

	return snippet, nil
}

func (r *TrashRepository) RestoreSnippet(ctx context.Context, id domain.SnippetID) (domain.Restoration, error) {
	restoration, err := r.restored(ctx, func(queries *sqlcgen.Queries) (domain.Restoration, error) {
		return snippetRestorable(queries).restore(ctx, queries, columnID(id))
	})
	if err != nil {
		return domain.Restoration{}, fmt.Errorf("sqlite.TrashRepository.RestoreSnippet: %w", err)
	}

	return restoration, nil
}

func (r *TrashRepository) RestoreFolder(ctx context.Context, id domain.FolderID) (domain.Restoration, error) {
	restoration, err := r.restored(ctx, func(queries *sqlcgen.Queries) (domain.Restoration, error) {
		return folderRestorable(queries).restore(ctx, queries, columnID(id))
	})
	if err != nil {
		return domain.Restoration{}, fmt.Errorf("sqlite.TrashRepository.RestoreFolder: %w", err)
	}

	return restoration, nil
}

func (r *TrashRepository) DeleteSnippet(ctx context.Context, id domain.SnippetID) error {
	err := inWriteTransaction(ctx, r.db, func(queries *sqlcgen.Queries) error {
		return deleteSnippetTrashItem(ctx, queries, id)
	})
	if err != nil {
		return fmt.Errorf("sqlite.TrashRepository.DeleteSnippet: %w", err)
	}

	return nil
}

func (r *TrashRepository) DeleteFolder(ctx context.Context, id domain.FolderID) error {
	err := inWriteTransaction(ctx, r.db, func(queries *sqlcgen.Queries) error {
		return deleteFolderTrashItem(ctx, queries, id)
	})
	if err != nil {
		return fmt.Errorf("sqlite.TrashRepository.DeleteFolder: %w", err)
	}

	return nil
}

func (r *TrashRepository) Empty(ctx context.Context, items domain.TrashItemIDs) (domain.TrashEmptied, error) {
	var emptied domain.TrashEmptied

	err := inWriteTransaction(ctx, r.db, func(queries *sqlcgen.Queries) error {
		before, countErr := countRows(ctx, queries)
		if countErr != nil {
			return countErr
		}

		emptyErr := emptyTrash(ctx, queries, items)
		if emptyErr != nil {
			return emptyErr
		}

		after, countErr := countRows(ctx, queries)
		if countErr != nil {
			return countErr
		}

		emptied = before.emptiedTo(after)

		return nil
	})
	if err != nil {
		return domain.TrashEmptied{}, fmt.Errorf("sqlite.TrashRepository.Empty: %w", err)
	}

	return emptied, nil
}

func (r *TrashRepository) restored(
	ctx context.Context, restore func(queries *sqlcgen.Queries) (domain.Restoration, error),
) (domain.Restoration, error) {
	var restoration domain.Restoration

	err := inWriteTransaction(ctx, r.db, func(queries *sqlcgen.Queries) error {
		var restoreErr error

		restoration, restoreErr = restore(queries)

		return restoreErr
	})

	return restoration, err
}

func selectTrashRows(ctx context.Context, queries *sqlcgen.Queries) (libraryRows, error) {
	folders, err := queries.ListLibraryFolders(ctx)
	if err != nil {
		return libraryRows{}, fmt.Errorf("list folders: %w", err)
	}

	snippets, err := trashSnippets().selectRows(ctx, queries)
	if err != nil {
		return libraryRows{}, err
	}

	return libraryRows{folders: folders, tags: nil, snippets: snippets}, nil
}

type restorable[R, P any] struct {
	get        func(ctx context.Context, id sqltype.ID) (R, error)
	remembered func(row R) *sqltype.ID
	formerPath func(row R) string
	paramsOf   func(into *sqltype.ID, id sqltype.ID) P
	write      func(ctx context.Context, params P) error
}

//nolint:dupl // a table of the Snippet columns that dupl matches against folderRestorable's; no logic repeats.
func snippetRestorable(queries *sqlcgen.Queries) restorable[sqlcgen.Snippet, sqlcgen.RestoreSnippetParams] {
	return restorable[sqlcgen.Snippet, sqlcgen.RestoreSnippetParams]{
		get:        queries.GetSnippetTrashItem,
		remembered: func(row sqlcgen.Snippet) *sqltype.ID { return row.FolderID },
		formerPath: func(row sqlcgen.Snippet) string { return row.FormerFolderPath },
		paramsOf: func(into *sqltype.ID, id sqltype.ID) sqlcgen.RestoreSnippetParams {
			return sqlcgen.RestoreSnippetParams{FolderID: into, ID: id}
		},
		write: queries.RestoreSnippet,
	}
}

//nolint:dupl // a table of the Folder columns that dupl matches against snippetRestorable's; no logic repeats.
func folderRestorable(queries *sqlcgen.Queries) restorable[sqlcgen.Folder, sqlcgen.RestoreFolderParams] {
	return restorable[sqlcgen.Folder, sqlcgen.RestoreFolderParams]{
		get:        queries.GetFolderTrashItem,
		remembered: func(row sqlcgen.Folder) *sqltype.ID { return row.ParentID },
		formerPath: func(row sqlcgen.Folder) string { return row.FormerParentPath },
		paramsOf: func(into *sqltype.ID, id sqltype.ID) sqlcgen.RestoreFolderParams {
			return sqlcgen.RestoreFolderParams{ParentID: into, ID: id}
		},
		write: queries.RestoreFolder,
	}
}

func (r restorable[R, P]) restore(
	ctx context.Context, queries *sqlcgen.Queries, id sqltype.ID,
) (domain.Restoration, error) {
	row, err := r.get(ctx, id)
	if err != nil {
		return domain.Restoration{}, trashItemErr(err)
	}

	restoration, err := restorationOf(ctx, queries, folderFromColumn(r.remembered(row)), r.formerPath(row))
	if err != nil {
		return domain.Restoration{}, err
	}

	err = r.write(ctx, r.paramsOf(folderColumn(restoration.FolderID), id))
	if err != nil {
		return domain.Restoration{}, fmt.Errorf("restore trash item: %w", err)
	}

	return restoration, nil
}

func restorationOf(
	ctx context.Context, queries *sqlcgen.Queries, remembered domain.FolderID, formerPathColumn string,
) (domain.Restoration, error) {
	formerPath, err := formerPathFromColumn(formerPathColumn)
	if err != nil {
		return domain.Restoration{}, corrupt(err)
	}

	if remembered.IsNil() {
		return domain.RestorationOf(remembered, formerPath, false), nil
	}

	live, err := folderIsLive(ctx, queries, remembered)
	if err != nil {
		return domain.Restoration{}, err
	}

	return domain.RestorationOf(remembered, formerPath, live), nil
}

func trashItemErr(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}

	return fmt.Errorf("get trash item: %w", err)
}
