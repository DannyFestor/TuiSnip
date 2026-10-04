package outcome

import "github.com/DannyFestor/TuiSnip/internal/domain"

type FolderSelected struct {
	ID domain.FolderID
}

func (FolderSelected) isOutcome() {}
