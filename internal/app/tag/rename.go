package tag

import (
	"context"
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Rename struct {
	repo  RenameRepository
	clock Clock
}

func NewRename(repo RenameRepository, clock Clock) (*Rename, error) {
	err := errors.Join(
		domain.RequireDependency("repo", repo),
		domain.RequireDependency("clock", clock),
	)
	if err != nil {
		return nil, fmt.Errorf("tag.NewRename: %w", err)
	}

	return &Rename{repo: repo, clock: clock}, nil
}

func (r *Rename) Run(ctx context.Context, input RenameInput) (domain.Tag, error) {
	name, err := parseName(input.Name)
	if err != nil {
		return domain.Tag{}, fmt.Errorf("tag.Rename: %w", err)
	}

	stored, err := r.repo.Find(ctx, input.TagID)
	if err != nil {
		return domain.Tag{}, fmt.Errorf("tag.Rename: %w", err)
	}

	renamed, err := stored.Rename(name, r.clock.Now())
	if err != nil {
		return domain.Tag{}, fmt.Errorf("tag.Rename: %w", err)
	}

	survivor, err := r.repo.Rename(ctx, renamed)
	if err != nil {
		return domain.Tag{}, fmt.Errorf("tag.Rename: %w", err)
	}

	return survivor, nil
}
