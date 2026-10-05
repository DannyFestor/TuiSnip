package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type SnippetCapturer interface {
	Run(ctx context.Context, in snippet.CaptureInput) (value.Content, error)
}
