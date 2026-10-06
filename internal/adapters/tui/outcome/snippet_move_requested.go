package outcome

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type SnippetMoveRequested struct {
	Input     snippet.MoveInput
	Selection browseselection.Selection
	Selecting domain.SnippetID
}

func (SnippetMoveRequested) isOutcome() {}
