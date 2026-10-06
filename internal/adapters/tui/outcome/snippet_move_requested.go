package outcome

import "github.com/DannyFestor/TuiSnip/internal/app/snippet"

type SnippetMoveRequested struct {
	Input snippet.MoveInput
}

func (SnippetMoveRequested) isOutcome() {}
