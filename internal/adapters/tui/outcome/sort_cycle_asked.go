package outcome

import "github.com/DannyFestor/TuiSnip/internal/domain"

type SortCycleAsked struct {
	FolderID  domain.FolderID
	Selecting domain.SnippetID
}

func (SortCycleAsked) isOutcome() {}
