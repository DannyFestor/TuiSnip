package searchpopup

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpath"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/snippetlist"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Listing struct {
	Snippets []domain.Snippet
	Paths    folderpath.Paths
}

func shortPathIn(paths folderpath.Paths) snippetlist.Meta {
	return func(snippet domain.Snippet) string {
		return paths.Short(snippet.FolderID())
	}
}
