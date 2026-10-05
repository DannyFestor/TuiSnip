package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type CollapsedFoldersSaver interface {
	SaveCollapsedFolders(ctx context.Context, ids []domain.FolderID) error
}
