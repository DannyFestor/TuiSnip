package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/folder"
)

type FolderDeletePreviewer interface {
	Run(ctx context.Context, in folder.PreviewDeleteInput) (folder.DeletePreview, error)
}
