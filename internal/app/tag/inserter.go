package tag

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Inserter interface {
	Insert(ctx context.Context, tag domain.Tag) error
}
