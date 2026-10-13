package mainscreen

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type SnippetsLoaded struct {
	Selection browseselection.Selection
	Snippets  []domain.Snippet
	Selecting domain.SnippetID
	Showing   domain.FragmentID
	Order     domain.SortOrder
}
