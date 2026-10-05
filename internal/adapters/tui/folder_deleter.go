package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/folder"
)

type FolderDeleter interface {
	Run(ctx context.Context, in folder.DeleteInput) error
}
