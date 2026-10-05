package outcome

import "github.com/DannyFestor/TuiSnip/internal/domain"

type SnippetReloaded struct {
	ID       domain.SnippetID
	FolderID domain.FolderID
}

func (SnippetReloaded) isOutcome() {}
