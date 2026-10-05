package browseselection

import "github.com/DannyFestor/TuiSnip/internal/domain"

type Selection struct {
	folderID domain.FolderID
	tagID    domain.TagID
}

func InFolder(id domain.FolderID) Selection {
	return Selection{folderID: id, tagID: domain.TagID{}}
}

func WithTag(id domain.TagID) Selection {
	return Selection{folderID: domain.FolderID{}, tagID: id}
}

func (s Selection) Folder() (domain.FolderID, bool) {
	return s.folderID, s.tagID.IsNil()
}

func (s Selection) Tag() (domain.TagID, bool) {
	return s.tagID, !s.tagID.IsNil()
}
