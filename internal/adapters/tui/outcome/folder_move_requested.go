package outcome

import "github.com/DannyFestor/TuiSnip/internal/app/folder"

type FolderMoveRequested struct {
	Input      folder.MoveInput
	Name       string
	TargetName string
}

func (FolderMoveRequested) isOutcome() {}
