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
	var row sqlcgen.Folder

	err := inReadTransaction(ctx, r.db, func(queries *sqlcgen.Queries) error {
		var getErr error

		row, getErr = queries.GetFolder(ctx, columnID(id))
		if errors.Is(getErr, sql.ErrNoRows) {
			return domain.ErrNotFound
		}

		if getErr != nil {
			return fmt.Errorf("get folder: %w", getErr)
		}

		return nil
	})
	if err != nil {
		return domain.Folder{}, fmt.Errorf("sqlite.FolderRepository.Find: %w", err)
	}

	folder, err := r.rebuild(ctx, row, slog.LevelError)
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
