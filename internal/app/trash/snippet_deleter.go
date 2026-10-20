package trash

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type SnippetDeleter interface {
	DeleteSnippet(ctx context.Context, id domain.SnippetID) error
}
