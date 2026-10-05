package browse

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type TagSnippetCounter interface {
	CountByTag(ctx context.Context) (map[domain.TagID]int, error)
}
