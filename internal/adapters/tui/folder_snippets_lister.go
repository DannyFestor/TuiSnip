package tui

import (
	"context"

	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type FolderSnippetsLister interface {
	Run(ctx context.Context, in browse.SnippetsInFolderInput) ([]domain.Snippet, error)
}
