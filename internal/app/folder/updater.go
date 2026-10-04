package folder

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Updater interface {
	Update(ctx context.Context, folder domain.Folder) error
}
