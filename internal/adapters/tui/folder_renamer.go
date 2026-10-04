package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type FolderRenamer interface {
	Run(ctx context.Context, in folder.RenameInput) (domain.Folder, error)
}
