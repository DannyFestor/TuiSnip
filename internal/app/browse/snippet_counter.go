package browse

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type SnippetCounter interface {
	CountByFolder(ctx context.Context) (map[domain.FolderID]int, error)
}
