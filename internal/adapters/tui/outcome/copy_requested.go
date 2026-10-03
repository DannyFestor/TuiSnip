package outcome

import "github.com/DannyFestor/TuiSnip/internal/domain"

type CopyRequested struct {
	ID domain.SnippetID
}

func (CopyRequested) isOutcome() {}
