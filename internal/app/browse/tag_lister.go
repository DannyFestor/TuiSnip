package browse

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type TagLister interface {
	List(ctx context.Context) ([]domain.Tag, error)
}
