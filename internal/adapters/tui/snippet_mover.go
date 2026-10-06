package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type SnippetMover interface {
	Run(ctx context.Context, in snippet.MoveInput) (domain.Snippet, error)
}
