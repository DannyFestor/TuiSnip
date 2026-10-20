package trash

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type FolderDeleter interface {
	DeleteFolder(ctx context.Context, id domain.FolderID) error
}
