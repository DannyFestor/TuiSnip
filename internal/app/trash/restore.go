package trash

import (
	"context"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Restore struct {
	repo RestoreRepository
}

func NewRestore(repo RestoreRepository) (*Restore, error) {
	err := domain.RequireDependency("repo", repo)
	if err != nil {
		return nil, fmt.Errorf("trash.NewRestore: %w", err)
	}

	return &Restore{repo: repo}, nil
}

func (r *Restore) Run(ctx context.Context, input RestoreInput) (domain.Restoration, error) {
	if folderID, ok := input.Item.Folder(); ok {
		return restored(r.repo.RestoreFolder(ctx, folderID))
	}

	snippetID, _ := input.Item.Snippet()

	return restored(r.repo.RestoreSnippet(ctx, snippetID))
}

func restored(restoration domain.Restoration, err error) (domain.Restoration, error) {
	if err != nil {
		return domain.Restoration{}, fmt.Errorf("trash.Restore: %w", err)
	}

	return restoration, nil
}
