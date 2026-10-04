package sqlite

import (
	"context"
	"database/sql"
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
		folder, err := folderFromRow(row)
		if err != nil {
			r.logger.WarnContext(
				ctx,
				"folder row is corrupt",
				slog.String(keyFolderID, entityID[domain.FolderID](row.ID).String()),
				slog.String(keyError, err.Error()),
			)

			continue
		}

		folders = append(folders, folder)
	}

	return folders
}
