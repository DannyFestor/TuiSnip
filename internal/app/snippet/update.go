package snippet

import (
	"context"
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Update struct {
	repo  UpdateRepository
	ids   IDGenerator
	clock Clock
}

func NewUpdate(repo UpdateRepository, ids IDGenerator, clock Clock) (*Update, error) {
	err := errors.Join(
		domain.RequireDependency("repo", repo),
		domain.RequireDependency("ids", ids),
		domain.RequireDependency("clock", clock),
	)
	if err != nil {
		return nil, fmt.Errorf("snippet.NewUpdate: %w", err)
	}

	return &Update{repo: repo, ids: ids, clock: clock}, nil
}

func (u *Update) Run(ctx context.Context, input UpdateInput) (domain.Snippet, error) {
	parsed, err := parseFields(input.raw())
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("snippet.Update: %w", err)
	}

	stored, err := u.repo.Find(ctx, input.SnippetID)
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("snippet.Update: %w", err)
	}

	edited, err := u.edited(stored, parsed, input)
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("snippet.Update: %w", err)
	}

	err = u.repo.Update(ctx, edited, input.LoadedUpdatedAt)
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("snippet.Update: %w", err)
	}

	return edited, nil
}

func (u *Update) edited(stored domain.Snippet, parsed fields, input UpdateInput) (domain.Snippet, error) {
	now := u.clock.Now()

	edited, err := stored.Edit(parsed.title, parsed.description, parsed.language, parsed.content, now)
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("edit: %w", err)
	}

	tags, err := parsed.tagsBeside(input.Tags, u.ids, now)
	if err != nil {
		return domain.Snippet{}, err
	}

	return edited.Retagged(tags), nil
}
