package folder

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Inserter interface {
	Insert(ctx context.Context, folder domain.Folder) error
}
