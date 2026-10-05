package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
)

type TagLister interface {
	Run(ctx context.Context, in browse.TagListInput) ([]browse.TagCount, error)
}
