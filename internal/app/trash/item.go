package trash

import (
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Item struct {
	Kind      ItemKind
	Snippet   domain.Snippet
	Folder    domain.Folder
	DeletedAt time.Time
	Path      []string
	Contents  Contents
}

func (i Item) ID() ItemID {
	if i.Kind.IsFolder() {
		return FolderItem(i.Folder.ID())
	}

	return SnippetItem(i.Snippet.ID())
}

func (i Item) Name() string {
	if i.Kind.IsFolder() {
		return i.Folder.Name().String()
	}

	return i.Snippet.Title().String()
}
