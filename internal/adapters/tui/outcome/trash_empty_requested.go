package outcome

import "github.com/DannyFestor/TuiSnip/internal/app/trash"

type TrashEmptyRequested struct {
	Input trash.EmptyInput
}

func (TrashEmptyRequested) isOutcome() {}
