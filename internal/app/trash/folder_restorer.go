package trash

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type FolderRestorer interface {
	RestoreFolder(ctx context.Context, id domain.FolderID) (domain.Restoration, error)
}
