package mainscreen

import "github.com/DannyFestor/TuiSnip/internal/domain"

type Remembered struct {
	SortOrder        domain.SortOrder
	CollapsedFolders []domain.FolderID
}
