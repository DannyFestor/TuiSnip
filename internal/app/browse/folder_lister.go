package browse

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type FolderLister interface {
	List(ctx context.Context) ([]domain.Folder, error)
}
