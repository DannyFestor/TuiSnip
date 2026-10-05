package folder

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Deleter interface {
	Delete(ctx context.Context, id domain.FolderID) error
}
