package outcome

import "github.com/DannyFestor/TuiSnip/internal/app/folder"

type FolderRenameRequested struct {
	Input        folder.RenameInput
	PreviousName string
}

func (FolderRenameRequested) isOutcome() {}
