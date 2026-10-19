package trash

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type SnippetRestorer interface {
	RestoreSnippet(ctx context.Context, id domain.SnippetID) (domain.Restoration, error)
}
