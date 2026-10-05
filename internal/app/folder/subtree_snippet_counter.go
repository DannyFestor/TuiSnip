package folder

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type SubtreeSnippetCounter interface {
	CountSnippetsInSubtree(ctx context.Context, id domain.FolderID) (int, error)
}
