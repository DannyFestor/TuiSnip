package search

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type Searcher interface {
	Search(ctx context.Context, query value.SearchQuery) ([]domain.SearchHit, error)
}
