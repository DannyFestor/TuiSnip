package search

import (
	"context"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Query struct {
	searcher Searcher
}

func NewQuery(searcher Searcher) (*Query, error) {
	err := domain.RequireDependency("searcher", searcher)
	if err != nil {
		return nil, fmt.Errorf("search.NewQuery: %w", err)
	}

	return &Query{searcher: searcher}, nil
}

func (q *Query) Run(ctx context.Context, in QueryInput) ([]domain.SearchHit, error) {
	if in.Text == "" {
		return []domain.SearchHit{}, nil
	}

	hits, err := q.searcher.Search(ctx, in.Text)
	if err != nil {
		return nil, fmt.Errorf("search.Query: %w", err)
	}

	return hits, nil
}
