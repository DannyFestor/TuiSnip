package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type SnippetUpdater interface {
	Run(ctx context.Context, in snippet.UpdateInput) (domain.Snippet, error)
}
