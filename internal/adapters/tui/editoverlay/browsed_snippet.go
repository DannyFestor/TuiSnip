package editoverlay

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type BrowsedSnippet struct {
	Snippet   domain.Snippet
	Selection browseselection.Selection
}
