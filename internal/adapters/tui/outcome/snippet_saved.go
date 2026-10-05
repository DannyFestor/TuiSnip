package outcome

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type SnippetSaved struct {
	ID        domain.SnippetID
	Selection browseselection.Selection
}

func (SnippetSaved) isOutcome() {}
