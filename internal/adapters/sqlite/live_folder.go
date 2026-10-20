package sqlite

import (
	"context"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/adapters/sqlite/sqlcgen"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

func requireLiveFolder(ctx context.Context, queries *sqlcgen.Queries, id domain.FolderID) error {
	if id.IsNil() {
		return nil
	}

	live, err := folderIsLive(ctx, queries, id)
	if err != nil {
		return err
	}

	if !live {
		return domain.ErrFolderGone
	}

	return nil
}

func folderIsLive(ctx context.Context, queries *sqlcgen.Queries, id domain.FolderID) (bool, error) {
	live, err := queries.IsFolderLive(ctx, columnID(id))
	if err != nil {
		return false, fmt.Errorf("check folder is live: %w", err)
	}

	return live, nil
}
