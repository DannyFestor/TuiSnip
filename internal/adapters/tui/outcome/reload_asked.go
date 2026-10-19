package outcome

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type ReloadAsked struct {
	Selection browseselection.Selection
	Selecting domain.SnippetID
}

func (ReloadAsked) isOutcome() {}
