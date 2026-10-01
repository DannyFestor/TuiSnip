package memsearch

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type SnippetLister interface {
	List(ctx context.Context) ([]domain.Snippet, error)
}
