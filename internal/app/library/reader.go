package library

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Reader interface {
	Read(ctx context.Context, path string) (domain.Library, error)
}
