package outcome

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
)

type SnippetDuplicateRequested struct {
	Input     snippet.DuplicateInput
	Selection browseselection.Selection
}

func (SnippetDuplicateRequested) isOutcome() {}
