package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
)

type SnippetDeleter interface {
	Run(ctx context.Context, in snippet.DeleteInput) error
}
