package outcome

import "github.com/DannyFestor/TuiSnip/internal/domain"

type SnippetSaved struct {
	ID domain.SnippetID
}

func (SnippetSaved) isOutcome() {}
