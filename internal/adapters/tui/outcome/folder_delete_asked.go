package outcome

import "github.com/DannyFestor/TuiSnip/internal/domain"

type FolderDeleteAsked struct {
	ID domain.FolderID
}

func (FolderDeleteAsked) isOutcome() {}
