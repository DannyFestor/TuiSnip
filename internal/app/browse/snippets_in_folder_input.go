package browse

import "github.com/DannyFestor/TuiSnip/internal/domain"

type SnippetsInFolderInput struct {
	FolderID domain.FolderID
	Order    domain.SortOrder
}
