package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type FolderCreator interface {
	Run(ctx context.Context, in folder.CreateInput) (domain.Folder, error)
}
