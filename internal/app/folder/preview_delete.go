package folder

import (
	"context"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type PreviewDelete struct {
	repo PreviewDeleteRepository
}

func NewPreviewDelete(repo PreviewDeleteRepository) (*PreviewDelete, error) {
	err := domain.RequireDependency("repo", repo)
	if err != nil {
		return nil, fmt.Errorf("folder.NewPreviewDelete: %w", err)
	}

	return &PreviewDelete{repo: repo}, nil
}

func (p *PreviewDelete) Run(ctx context.Context, input PreviewDeleteInput) (DeletePreview, error) {
	stored, err := p.repo.Find(ctx, input.FolderID)
	if err != nil {
		return DeletePreview{}, fmt.Errorf("folder.PreviewDelete: %w", err)
	}

	subfolderCount, err := p.repo.CountSubfolders(ctx, input.FolderID)
	if err != nil {
		return DeletePreview{}, fmt.Errorf("folder.PreviewDelete: %w", err)
	}

	snippetCount, err := p.repo.CountSnippetsInSubtree(ctx, input.FolderID)
	if err != nil {
		return DeletePreview{}, fmt.Errorf("folder.PreviewDelete: %w", err)
	}

	return DeletePreview{Folder: stored, SubfolderCount: subfolderCount, SnippetCount: snippetCount}, nil
}
