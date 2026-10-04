package searchpopup

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpath"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Listing struct {
	Snippets []domain.Snippet
	Paths    folderpath.Paths
}
