package outcome

import "github.com/DannyFestor/TuiSnip/internal/app/trash"

type TrashDeleteRequested struct {
	Input trash.DeletePermanentlyInput
	Name  string
}

func (TrashDeleteRequested) isOutcome() {}
