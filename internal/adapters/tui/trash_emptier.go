package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/trash"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type TrashEmptier interface {
	Run(ctx context.Context, in trash.EmptyInput) (domain.TrashEmptied, error)
}
