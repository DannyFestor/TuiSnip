package system

import (
	"uuid"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type IDs struct{}

func NewIDs() IDs {
	return IDs{}
}

func (IDs) NewSnippetID() domain.SnippetID {
	return domain.SnippetID(uuid.NewV7())
}

func (IDs) NewFragmentID() domain.FragmentID {
	return domain.FragmentID(uuid.NewV7())
}

func (IDs) NewFolderID() domain.FolderID {
	return domain.FolderID(uuid.NewV7())
}
