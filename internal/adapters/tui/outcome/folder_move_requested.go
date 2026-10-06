package outcome

import "github.com/DannyFestor/TuiSnip/internal/app/folder"

type FolderMoveRequested struct {
	Input folder.MoveInput
}

func (FolderMoveRequested) isOutcome() {}
