package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/folder"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type FolderDefaultLanguageSetter interface {
	Run(ctx context.Context, in folder.SetDefaultLanguageInput) (domain.Folder, error)
}
