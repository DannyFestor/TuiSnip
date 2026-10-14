package library

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type EmptyInserter interface {
	InsertIntoEmpty(ctx context.Context, library domain.Library) error
}
