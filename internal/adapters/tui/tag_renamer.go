package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/tag"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type TagRenamer interface {
	Run(ctx context.Context, in tag.RenameInput) (domain.Tag, error)
}
