package library

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Writer interface {
	Write(ctx context.Context, path string, library domain.Library) error
}
