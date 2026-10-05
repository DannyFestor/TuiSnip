package outcome

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type ExternalEditAsked struct {
	Content   string
	Language  value.Language
	Snippet   domain.Snippet
	Selection browseselection.Selection
}

func (ExternalEditAsked) isOutcome() {}
