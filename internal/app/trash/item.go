package trash

import (
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Item struct {
	Snippet   domain.Snippet
	Folder    domain.Folder
	DeletedAt time.Time
	Path      []string
	Contents  Contents
}

type Contents struct {
	SubfolderCount int
	SnippetCount   int
	Outline        Outline
}

type Outline struct {
	Name          string
	SnippetTitles []string
	Subfolders    []Outline
}

func (i Item) ID() ItemID {
	if i.IsFolder() {
		return FolderItem(i.Folder.ID())
	}

	return SnippetItem(i.Snippet.ID())
}

func (i Item) IsFolder() bool {
	return !i.Folder.ID().IsNil()
}

func (i Item) Name() string {
	if i.IsFolder() {
		return i.Folder.Name().String()
	}

	return i.Snippet.Title().String()
}
