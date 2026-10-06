package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type FolderMover interface {
	Run(ctx context.Context, in folder.MoveInput) (domain.Folder, error)
}
