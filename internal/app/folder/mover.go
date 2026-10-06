package folder

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Mover interface {
	Move(ctx context.Context, folder domain.Folder) error
}
