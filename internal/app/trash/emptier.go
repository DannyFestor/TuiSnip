package trash

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Emptier interface {
	Empty(ctx context.Context, items domain.TrashItemIDs) (domain.TrashEmptied, error)
}
