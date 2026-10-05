package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/tag"
)

type TagDeletePreviewer interface {
	Run(ctx context.Context, in tag.PreviewDeleteInput) (tag.DeletePreview, error)
}
