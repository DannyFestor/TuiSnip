package folder

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Finder interface {
	Find(ctx context.Context, id domain.FolderID) (domain.Folder, error)
}
