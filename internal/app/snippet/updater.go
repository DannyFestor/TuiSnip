package snippet

import (
	"context"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Updater interface {
	Update(ctx context.Context, snippet domain.Snippet, loadedUpdatedAt time.Time) error
}
