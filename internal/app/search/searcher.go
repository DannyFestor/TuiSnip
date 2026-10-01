package search

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Searcher interface {
	Search(ctx context.Context, query string) ([]domain.SearchHit, error)
}
