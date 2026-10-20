package trash

import (
	"cmp"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type ItemID struct {
	kind      ItemKind
	snippetID domain.SnippetID
	folderID  domain.FolderID
}

func SnippetItem(id domain.SnippetID) ItemID {
	return ItemID{kind: ItemKindSnippet, snippetID: id, folderID: domain.FolderID{}}
}

func FolderItem(id domain.FolderID) ItemID {
	return ItemID{kind: ItemKindFolder, snippetID: domain.SnippetID{}, folderID: id}
}

func (i ItemID) Kind() ItemKind {
	return i.kind
}

func (i ItemID) Snippet() (domain.SnippetID, bool) {
	return i.snippetID, i.kind == ItemKindSnippet
}

func (i ItemID) Folder() (domain.FolderID, bool) {
	return i.folderID, i.kind == ItemKindFolder
}

func (i ItemID) String() string {
	if i.kind == ItemKindFolder {
		return i.folderID.String()
	}

	return i.snippetID.String()
}

func (i ItemID) compare(other ItemID) int {
	return cmp.Compare(i.String(), other.String())
}
