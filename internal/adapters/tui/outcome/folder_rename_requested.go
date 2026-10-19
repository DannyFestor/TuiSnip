package outcome

import "github.com/DannyFestor/TuiSnip/internal/app/folder"

type FolderRenameRequested struct {
	Input   folder.RenameInput
	Current string
}

func (FolderRenameRequested) isOutcome() {}
