package folder

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type DescendantLister interface {
	ListDescendantIDs(ctx context.Context, id domain.FolderID) ([]domain.FolderID, error)
}
