package outcome

import "github.com/DannyFestor/TuiSnip/internal/domain"

type CollapsedFoldersChanged struct {
	IDs []domain.FolderID
}

func (CollapsedFoldersChanged) isOutcome() {}
