package snippet

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Inserter interface {
	Insert(ctx context.Context, snippet domain.Snippet) error
}
