package outcome

import "github.com/DannyFestor/TuiSnip/internal/app/snippet"

type SaveRequested struct {
	Input snippet.CreateInput
}

func (SaveRequested) isOutcome() {}
