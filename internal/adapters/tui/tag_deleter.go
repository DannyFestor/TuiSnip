package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/tag"
)

type TagDeleter interface {
	Run(ctx context.Context, in tag.DeleteInput) error
}
