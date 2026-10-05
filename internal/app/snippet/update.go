package snippet

import (
	"context"
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Update struct {
	repo  UpdateRepository
	clock Clock
}

func NewUpdate(repo UpdateRepository, clock Clock) (*Update, error) {
	err := errors.Join(
		domain.RequireDependency("repo", repo),
		domain.RequireDependency("clock", clock),
	)
	if err != nil {
		return nil, fmt.Errorf("snippet.NewUpdate: %w", err)
	}

	return &Update{repo: repo, clock: clock}, nil
}

func (u *Update) Run(ctx context.Context, input UpdateInput) (domain.Snippet, error) {
	parsed, err := parseFields(input.Title, input.Description, input.Content)
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("snippet.Update: %w", err)
	}

	stored, err := u.repo.Find(ctx, input.SnippetID)
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("snippet.Update: %w", err)
	}

	edited, err := stored.Edit(parsed.title, parsed.description, parsed.content, u.clock.Now())
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("snippet.Update: %w", err)
	}

	err = u.repo.Update(ctx, edited, input.LoadedUpdatedAt)
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("snippet.Update: %w", err)
	}

	return edited, nil
}
