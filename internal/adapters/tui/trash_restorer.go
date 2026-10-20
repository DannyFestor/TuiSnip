package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/trash"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type TrashRestorer interface {
	Run(ctx context.Context, in trash.RestoreInput) (domain.Restoration, error)
}
