package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/trash"
)

type TrashLister interface {
	Run(ctx context.Context, in trash.ListInput) ([]trash.Item, error)
}
