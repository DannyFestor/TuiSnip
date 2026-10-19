package outcome

import (
	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type FolderDeleteRequested struct {
	Input    folder.DeleteInput
	Name     string
	ParentID domain.FolderID
}

func (FolderDeleteRequested) isOutcome() {}
