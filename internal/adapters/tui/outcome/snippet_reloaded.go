package outcome

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type SnippetReloaded struct {
	ID        domain.SnippetID
	Selection browseselection.Selection
}

func (SnippetReloaded) isOutcome() {}
