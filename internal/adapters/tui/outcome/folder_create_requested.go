package outcome

import "github.com/DannyFestor/TuiSnip/internal/app/folder"

type FolderCreateRequested struct {
	Input      folder.CreateInput
	ParentName string
}

func (FolderCreateRequested) isOutcome() {}
