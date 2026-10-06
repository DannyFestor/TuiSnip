package snippet

import (
	"context"
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Duplicate struct {
	repo  DuplicateRepository
	ids   IDGenerator
	clock Clock
}

func NewDuplicate(repo DuplicateRepository, ids IDGenerator, clock Clock) (*Duplicate, error) {
	err := errors.Join(
		domain.RequireDependency("repo", repo),
		domain.RequireDependency("ids", ids),
		domain.RequireDependency("clock", clock),
	)
	if err != nil {
		return nil, fmt.Errorf("snippet.NewDuplicate: %w", err)
	}

	return &Duplicate{repo: repo, ids: ids, clock: clock}, nil
}

func (d *Duplicate) Run(ctx context.Context, input DuplicateInput) (domain.Snippet, error) {
	original, err := d.repo.Find(ctx, input.SnippetID)
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("snippet.Duplicate: %w", err)
	}

	duplicate, err := original.Duplicate(d.ids.NewSnippetID(), d.ids.NewFragmentID(), d.clock.Now())
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("snippet.Duplicate: %w", err)
	}

	err = d.repo.Insert(ctx, duplicate)
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("snippet.Duplicate: %w", err)
	}

	return duplicate, nil
}
