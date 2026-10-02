package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type SnippetCreator interface {
	Run(ctx context.Context, in snippet.CreateInput) (domain.Snippet, error)
}
