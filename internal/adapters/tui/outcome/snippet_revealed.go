package outcome

import "github.com/DannyFestor/TuiSnip/internal/domain"

type SnippetRevealed struct {
	ID       domain.SnippetID
	FolderID domain.FolderID
}

func (SnippetRevealed) isOutcome() {}
