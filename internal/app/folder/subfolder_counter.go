package folder

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type SubfolderCounter interface {
	CountSubfolders(ctx context.Context, id domain.FolderID) (int, error)
}
