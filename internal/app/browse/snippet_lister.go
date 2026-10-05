package browse

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type SnippetLister interface {
	ListInFolder(ctx context.Context, folderID domain.FolderID, order domain.SortOrder) ([]domain.Snippet, error)
}
