package folder

import (
	"context"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Trasher interface {
	Trash(ctx context.Context, id domain.FolderID, deletedAt time.Time) error
}
