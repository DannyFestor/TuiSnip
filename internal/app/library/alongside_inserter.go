package library

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type AlongsideInserter interface {
	InsertAlongside(ctx context.Context, library domain.Library) (domain.Library, error)
}
