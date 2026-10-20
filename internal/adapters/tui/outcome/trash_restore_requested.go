package outcome

import "github.com/DannyFestor/TuiSnip/internal/app/trash"

type TrashRestoreRequested struct {
	Input trash.RestoreInput
	Name  string
	Path  []string
}

func (TrashRestoreRequested) isOutcome() {}
