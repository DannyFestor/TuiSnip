package snippet

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Copier interface {
	Copy(ctx context.Context, text string) (domain.CopyDelivery, error)
}
