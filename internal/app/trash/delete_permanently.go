package trash

import (
	"context"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type DeletePermanently struct {
	repo DeletePermanentlyRepository
}

func NewDeletePermanently(repo DeletePermanentlyRepository) (*DeletePermanently, error) {
	err := domain.RequireDependency("repo", repo)
	if err != nil {
		return nil, fmt.Errorf("trash.NewDeletePermanently: %w", err)
	}

	return &DeletePermanently{repo: repo}, nil
}

func (d *DeletePermanently) Run(ctx context.Context, input DeletePermanentlyInput) error {
	if folderID, ok := input.Item.Folder(); ok {
		return deleted(d.repo.DeleteFolder(ctx, folderID))
	}

	snippetID, _ := input.Item.Snippet()

	return deleted(d.repo.DeleteSnippet(ctx, snippetID))
}

func deleted(err error) error {
	if err != nil {
		return fmt.Errorf("trash.DeletePermanently: %w", err)
	}

	return nil
}
