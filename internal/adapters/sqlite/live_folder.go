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

	live, err := queries.IsFolderLive(ctx, columnID(id))
	if err != nil {
		return fmt.Errorf("check folder is live: %w", err)
	}

	if !live {
		return domain.ErrFolderGone
	}

	return nil
}
