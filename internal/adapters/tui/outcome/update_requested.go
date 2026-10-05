package outcome

import "github.com/DannyFestor/TuiSnip/internal/app/snippet"

type UpdateRequested struct {
	Input snippet.UpdateInput
}

func (UpdateRequested) isOutcome() {}
