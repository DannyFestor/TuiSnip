package outcome

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type SnippetDeleteRequested struct {
	Input     snippet.DeleteInput
	Selection browseselection.Selection
	Selecting domain.SnippetID
}

func (SnippetDeleteRequested) isOutcome() {}
