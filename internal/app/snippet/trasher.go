package snippet

import (
	"context"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Trasher interface {
	Trash(ctx context.Context, id domain.SnippetID, deletedAt time.Time) error
}
