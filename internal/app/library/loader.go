package library

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Loader interface {
	Load(ctx context.Context) (domain.Library, error)
}
