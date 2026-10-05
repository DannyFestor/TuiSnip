package snippet

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DannyFestor/TuiSnip/internal/domain"
	"github.com/DannyFestor/TuiSnip/internal/domain/value"
)

type Create struct {
	inserter Inserter
	ids      IDGenerator
	clock    Clock
}

type createFields struct {
	title       value.Title
	description value.Description
	content     value.Content
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

func (c *Create) Run(ctx context.Context, in CreateInput) (domain.Snippet, error) {
	fields, err := parseCreateInput(in)
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("snippet.Create: %w", err)
	}

	snippet, err := c.atRoot(fields, c.clock.Now())
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("snippet.Create: %w", err)
	}

	err = c.inserter.Insert(ctx, snippet)
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("snippet.Create: %w", err)
	}

	return snippet, nil
}

func (c *Create) atRoot(fields createFields, now time.Time) (domain.Snippet, error) {
	fragment, err := domain.NewFragment(c.ids.NewFragmentID(), value.PlainText(), fields.content, now, now)
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("new fragment: %w", err)
	}

	snippet, err := domain.NewSnippet(
		c.ids.NewSnippetID(),
		fields.title,
		fields.description,
		domain.FolderID{},
		[]domain.Fragment{fragment},
		nil,
		now,
		now,
	)
	if err != nil {
		return domain.Snippet{}, fmt.Errorf("new snippet: %w", err)
	}

	return snippet, nil
}

func parseCreateInput(in CreateInput) (createFields, error) {
	title, titleErr := value.NewTitle(in.Title)
	description, descriptionErr := value.NewDescription(in.Description)
	content, contentErr := value.NewContent(in.Content)

	return createFields{title: title, description: description, content: content}, errors.Join(
		domain.OnField(domain.FieldTitle, titleErr),
		domain.OnField(domain.FieldDescription, descriptionErr),
		domain.OnField(domain.FieldContent, contentErr),
	)
}
