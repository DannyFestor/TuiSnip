package outcome

import "github.com/DannyFestor/TuiSnip/internal/domain"

type SnippetSaved struct {
	ID       domain.SnippetID
	FolderID domain.FolderID
}

func (SnippetSaved) isOutcome() {}
