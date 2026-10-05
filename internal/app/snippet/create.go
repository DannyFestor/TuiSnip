package snippet

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Create struct {
	inserter Inserter
	ids      IDGenerator
	clock    Clock
}

func NewCreate(inserter Inserter, ids IDGenerator, clock Clock) (*Create, error) {
	err := errors.Join(
		domain.RequireDependency("inserter", inserter),
		domain.RequireDependency("ids", ids),
		domain.RequireDependency("clock", clock),
	)
	if err != nil {
		return nil, fmt.Errorf("snippet.NewCreate: %w", err)
	}

	return &Create{inserter: inserter, ids: ids, clock: clock}, nil
}

func (c *Create) Run(ctx context.Context, input CreateInput) (domain.Snippet, error) {
	parsed, err := parseFields(input.Title, input.Description, input.Language, input.Content)
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("snippet.Create: %w", err)
	}

	snippet, err := c.filed(parsed, input, c.clock.Now())
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("snippet.Create: %w", err)
	}

	err = c.inserter.Insert(ctx, snippet)
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("snippet.Create: %w", err)
	}

	return snippet, nil
}

func (c *Create) filed(parsed fields, input CreateInput, now time.Time) (domain.Snippet, error) {
	fragment, err := domain.NewFragment(c.ids.NewFragmentID(), parsed.language, parsed.content, now, now)
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("new fragment: %w", err)
	}

	snippet, err := domain.NewSnippet(
		c.ids.NewSnippetID(),
		parsed.title,
		parsed.description,
		input.FolderID,
		[]domain.Fragment{fragment},
		input.Tags,
		now,
		now,
	)
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("new snippet: %w", err)
	}

	return snippet, nil
}
