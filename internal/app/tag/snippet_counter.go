package tag

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type SnippetCounter interface {
	CountSnippets(ctx context.Context, id domain.TagID) (int, error)
}
