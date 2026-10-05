package browse

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type TagSnippetLister interface {
	ListWithTag(ctx context.Context, tagID domain.TagID, order domain.SortOrder) ([]domain.Snippet, error)
}
