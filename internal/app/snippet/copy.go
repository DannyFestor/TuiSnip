package snippet

import (
	"context"
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type Copy struct {
	finder Finder
	copier Copier
	shape  func(value.Content) value.Content
}

func NewCopy(finder Finder, copier Copier) (*Copy, error) {
	copyAction, err := copyShapedBy(finder, copier, unchanged)
	if err != nil {
		return nil, fmt.Errorf("snippet.NewCopy: %w", err)
	}

	return copyAction, nil
}

func NewCopyTrimmingTrailingNewline(finder Finder, copier Copier) (*Copy, error) {
	copyAction, err := copyShapedBy(finder, copier, value.Content.WithoutTrailingNewline)
	if err != nil {
		return nil, fmt.Errorf("snippet.NewCopyTrimmingTrailingNewline: %w", err)
	}

	return copyAction, nil
}

func (c *Copy) Run(ctx context.Context, in CopyInput) (CopyResult, error) {
	snippet, err := c.finder.Find(ctx, in.SnippetID)
	if err != nil {
		return CopyResult{}, fmt.Errorf("snippet.Copy: %w", err)
	}

	content := c.shape(snippet.FirstFragment().Content())

	delivery, err := c.copier.Copy(ctx, content.String())
	if err != nil {
		return CopyResult{}, fmt.Errorf("snippet.Copy: %w", err)
	}

	return CopyResult{Delivery: delivery, Content: content}, nil
}

func copyShapedBy(finder Finder, copier Copier, shape func(value.Content) value.Content) (*Copy, error) {
	err := errors.Join(
		domain.RequireDependency("finder", finder),
		domain.RequireDependency("copier", copier),
	)
	if err != nil {
		return nil, err
	}

	return &Copy{finder: finder, copier: copier, shape: shape}, nil
}

func unchanged(content value.Content) value.Content {
	return content
}
