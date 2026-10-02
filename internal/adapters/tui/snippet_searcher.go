package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/search"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type SnippetSearcher interface {
	Run(ctx context.Context, in search.QueryInput) ([]domain.SearchHit, error)
}
