package snippet

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Mover interface {
	Move(ctx context.Context, snippet domain.Snippet) error
}
