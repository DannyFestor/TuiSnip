package tui

import (
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type snippetsLoadedMsg struct {
	snippets []domain.Snippet
	err      error
}

type copyFinishedMsg struct {
	result snippet.CopyResult
	err    error
}
