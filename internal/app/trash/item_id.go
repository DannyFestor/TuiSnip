package trash

import (
	"cmp"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type ItemID struct {
	snippetID domain.SnippetID
	folderID  domain.FolderID
}

func SnippetItem(id domain.SnippetID) ItemID {
	return ItemID{snippetID: id, folderID: domain.FolderID{}}
}

func FolderItem(id domain.FolderID) ItemID {
	return ItemID{snippetID: domain.SnippetID{}, folderID: id}
}

func (i ItemID) Snippet() (domain.SnippetID, bool) {
	return i.snippetID, !i.snippetID.IsNil()
}

func (i ItemID) Folder() (domain.FolderID, bool) {
	return i.folderID, !i.folderID.IsNil()
}

func (i ItemID) IsNil() bool {
	return i.snippetID.IsNil() && i.folderID.IsNil()
}

func (i ItemID) String() string {
	if id, ok := i.Folder(); ok {
		return id.String()
	}

	return i.snippetID.String()
}

func (i ItemID) compare(other ItemID) int {
	return cmp.Compare(i.String(), other.String())
}
