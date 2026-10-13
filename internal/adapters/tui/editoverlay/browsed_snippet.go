package editoverlay

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type BrowsedSnippet struct {
	Snippet         domain.Snippet
	Selection       browseselection.Selection
	DefaultLanguage value.Language
}
