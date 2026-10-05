package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type SortOrderSaver interface {
	SaveSortOrder(ctx context.Context, order domain.SortOrder) error
}
