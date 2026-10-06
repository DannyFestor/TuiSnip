package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqlcgen"
	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqltype"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type FolderRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewFolderRepository(database *Database, logger *slog.Logger) *FolderRepository {
	return &FolderRepository{db: database.db, logger: logger}
}

func (r *FolderRepository) Insert(ctx context.Context, folder domain.Folder) error {
	err := inWriteTransaction(ctx, r.db, func(queries *sqlcgen.Queries) error {
		return queries.InsertFolder(ctx, insertFolderParams(folder))
	})
	if err != nil {
		return fmt.Errorf("sqlite.FolderRepository.Insert: %w", err)
	}

	return nil
}

func (r *FolderRepository) Find(ctx context.Context, id domain.FolderID) (domain.Folder, error) {
	get := func(queries *sqlcgen.Queries) (sqlcgen.Folder, error) { return queries.GetFolder(ctx, columnID(id)) }

	folder, err := findRebuilt(ctx, r.db, get, r.rebuild)
	if err != nil {
		return domain.Folder{}, fmt.Errorf("sqlite.FolderRepository.Find: %w", err)
	}

	return folder, nil
}

func (r *FolderRepository) Update(ctx context.Context, folder domain.Folder) error {
	err := inWriteTransaction(ctx, r.db, func(queries *sqlcgen.Queries) error {
		updated, updateErr := queries.UpdateFolder(ctx, updateFolderParams(folder))
		if updateErr != nil {
			return fmt.Errorf("update folder: %w", updateErr)
		}

		if updated == 0 {
			return domain.ErrNotFound
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("sqlite.FolderRepository.Update: %w", err)
	}

	return nil
}

func (r *FolderRepository) Delete(ctx context.Context, id domain.FolderID) error {
	err := inWriteTransaction(ctx, r.db, func(queries *sqlcgen.Queries) error {
		deleted, deleteErr := queries.DeleteFolder(ctx, columnID(id))
		if deleteErr != nil {
			return fmt.Errorf("delete folder: %w", deleteErr)
		}

		if deleted == 0 {
			return domain.ErrNotFound
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("sqlite.FolderRepository.Delete: %w", err)
	}

	return nil
}

func (r *FolderRepository) ListDescendantIDs(ctx context.Context, id domain.FolderID) ([]domain.FolderID, error) {
	rows, err := readRows(ctx, r.db, func(queries *sqlcgen.Queries, ctx context.Context) ([]sqltype.ID, error) {
		return queries.ListDescendantFolderIDs(ctx, folderColumn(id))
	})
	if err != nil {
		return nil, fmt.Errorf("sqlite.FolderRepository.ListDescendantIDs: %w", err)
	}

	return folderIDsOf(rows), nil
}

// Another instance can change the tree between the Action's cycle check and this write,
// so the check runs again on the tree the transaction sees.
func (r *FolderRepository) Move(ctx context.Context, folder domain.Folder) error {
	err := inWriteTransaction(ctx, r.db, func(queries *sqlcgen.Queries) error {
		cycleErr := requireNoCycle(ctx, queries, folder)
		if cycleErr != nil {
			return cycleErr
		}

		return moveFolder(ctx, queries, folder)
	})
	if err != nil {
		return fmt.Errorf("sqlite.FolderRepository.Move: %w", err)
	}

	return nil
}

func (r *FolderRepository) CountSubfolders(ctx context.Context, id domain.FolderID) (int, error) {
	count, err := readCount(ctx, r.db, func(queries *sqlcgen.Queries) (int64, error) {
		return queries.CountSubfolders(ctx, columnID(id))
	})
	if err != nil {
		return 0, fmt.Errorf("sqlite.FolderRepository.CountSubfolders: %w", err)
	}

	return count, nil
}

func (r *FolderRepository) CountSnippetsInSubtree(ctx context.Context, id domain.FolderID) (int, error) {
	count, err := readCount(ctx, r.db, func(queries *sqlcgen.Queries) (int64, error) {
		return queries.CountSnippetsInSubtree(ctx, columnID(id))
	})
	if err != nil {
		return 0, fmt.Errorf("sqlite.FolderRepository.CountSnippetsInSubtree: %w", err)
	}

	return count, nil
}

func (r *FolderRepository) List(ctx context.Context) ([]domain.Folder, error) {
	var rows []sqlcgen.Folder

	err := inReadTransaction(ctx, r.db, func(queries *sqlcgen.Queries) error {
		var listErr error

		rows, listErr = queries.ListFolders(ctx)
		if listErr != nil {
			return fmt.Errorf("list folders: %w", listErr)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("sqlite.FolderRepository.List: %w", err)
	}

	return r.rebuildAll(ctx, rows), nil
}

func (r *FolderRepository) rebuildAll(ctx context.Context, rows []sqlcgen.Folder) []domain.Folder {
	folders := make([]domain.Folder, 0, len(rows))

	for _, row := range rows {
		folder, err := r.rebuild(ctx, row, slog.LevelWarn)
		if err != nil {
			continue
		}

		folders = append(folders, folder)
	}

	return folders
}

func (r *FolderRepository) rebuild(
	ctx context.Context,
	row sqlcgen.Folder,
	corruptLevel slog.Level,
) (domain.Folder, error) {
	folder, err := folderFromRow(row)
	if err != nil {
		r.logger.LogAttrs(
			ctx,
			corruptLevel,
			"folder row is corrupt",
			slog.String(keyFolderID, entityID[domain.FolderID](row.ID).String()),
			slog.String(keyError, err.Error()),
		)
	}

	return folder, err
}

func requireNoCycle(ctx context.Context, queries *sqlcgen.Queries, folder domain.Folder) error {
	if folder.AtRoot() {
		return nil
	}

	descendants, err := queries.ListDescendantFolderIDs(ctx, folderColumn(folder.ID()))
	if err != nil {
		return fmt.Errorf("list descendant folders: %w", err)
	}

	_, err = folder.MoveUnder(folder.ParentID(), folderIDsOf(descendants))
	if err != nil {
		return fmt.Errorf("check folder cycle: %w", err)
	}

	return nil
}

func moveFolder(ctx context.Context, queries *sqlcgen.Queries, folder domain.Folder) error {
	moved, err := queries.MoveFolder(ctx, moveFolderParams(folder))
	if err != nil {
		return fmt.Errorf("move folder: %w", err)
	}

	if moved == 0 {
		return domain.ErrNotFound
	}

	return nil
}
