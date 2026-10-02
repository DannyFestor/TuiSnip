package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
)

type SnippetCopier interface {
	Run(ctx context.Context, in snippet.CopyInput) (snippet.CopyResult, error)
}
