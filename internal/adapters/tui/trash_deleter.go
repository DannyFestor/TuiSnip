package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/trash"
)

type TrashDeleter interface {
	Run(ctx context.Context, in trash.DeletePermanentlyInput) error
}
