package outcome

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type TagsChanged struct {
	Selection browseselection.Selection
	Selecting domain.SnippetID
}

func (TagsChanged) isOutcome() {}
