package folder

import (
	"context"
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Rename struct {
	repo  EditRepository
	clock Clock
}

func NewRename(repo EditRepository, clock Clock) (*Rename, error) {
	err := errors.Join(
		domain.RequireDependency("repo", repo),
		domain.RequireDependency("clock", clock),
	)
	if err != nil {
		return nil, fmt.Errorf("folder.NewRename: %w", err)
	}

	return &Rename{repo: repo, clock: clock}, nil
}

func (r *Rename) Run(ctx context.Context, input RenameInput) (domain.Folder, error) {
	name, err := parseName(input.Name)
	if err != nil {
		return domain.Folder{}, fmt.Errorf("folder.Rename: %w", err)
	}

	stored, err := r.repo.Find(ctx, input.FolderID)
	if err != nil {
		return domain.Folder{}, fmt.Errorf("folder.Rename: %w", err)
	}

	renamed, err := stored.Rename(name, r.clock.Now())
	if err != nil {
		return domain.Folder{}, fmt.Errorf("folder.Rename: %w", err)
	}

	err = r.repo.Update(ctx, renamed)
	if err != nil {
		return domain.Folder{}, fmt.Errorf("folder.Rename: %w", err)
	}

	return renamed, nil
}
