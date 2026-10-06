package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type SnippetDuplicator interface {
	Run(ctx context.Context, in snippet.DuplicateInput) (domain.Snippet, error)
}
