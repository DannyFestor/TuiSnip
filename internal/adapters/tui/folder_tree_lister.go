package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
)

type FolderTreeLister interface {
	Run(ctx context.Context, in browse.FolderTreeInput) (browse.Tree, error)
}
