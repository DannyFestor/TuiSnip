package outcome

import "github.com/DannyFestor/TuiSnip/internal/domain"

type SnippetRevealed struct {
	ID domain.SnippetID
}

func (SnippetRevealed) isOutcome() {}
