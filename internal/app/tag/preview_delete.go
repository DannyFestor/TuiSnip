package tag

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
		return nil, fmt.Errorf("tag.NewPreviewDelete: %w", err)
	}

	return &PreviewDelete{repo: repo}, nil
}

func (p *PreviewDelete) Run(ctx context.Context, input PreviewDeleteInput) (DeletePreview, error) {
	stored, err := p.repo.Find(ctx, input.TagID)
	if err != nil {
		return DeletePreview{}, fmt.Errorf("tag.PreviewDelete: %w", err)
	}

	snippetCount, err := p.repo.CountSnippets(ctx, input.TagID)
	if err != nil {
		return DeletePreview{}, fmt.Errorf("tag.PreviewDelete: %w", err)
	}

	return DeletePreview{Tag: stored, SnippetCount: snippetCount}, nil
}
