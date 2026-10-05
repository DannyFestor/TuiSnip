package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type TagSnippetsLister interface {
	Run(ctx context.Context, in browse.SnippetsWithTagInput) ([]domain.Snippet, error)
}
