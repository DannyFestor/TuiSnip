package library

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type SnippetsLabReader interface {
	Read(ctx context.Context, path string) (domain.Library, domain.ImportLosses, error)
}
