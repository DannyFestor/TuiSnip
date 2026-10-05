package tag

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Renamer interface {
	Rename(ctx context.Context, tag domain.Tag) (domain.Tag, error)
}
