package outcome

import "github.com/DannyFestor/TuiSnip/internal/app/folder"

type FolderRenameRequested struct {
	Input folder.RenameInput
}

func (FolderRenameRequested) isOutcome() {}
