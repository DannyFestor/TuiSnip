package tag

import (
	"context"
	"errors"
	"fmt"

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
		return nil, fmt.Errorf("tag.NewCreate: %w", err)
	}

	return &Create{inserter: inserter, ids: ids, clock: clock}, nil
}

func (c *Create) Run(ctx context.Context, input CreateInput) (domain.Tag, error) {
	name, err := parseName(input.Name)
	if err != nil {
		return domain.Tag{}, fmt.Errorf("tag.Create: %w", err)
	}

	now := c.clock.Now()

	created, err := domain.NewTag(c.ids.NewTagID(), name, now, now)
	if err != nil {
		return domain.Tag{}, fmt.Errorf("tag.Create: %w", err)
	}

	err = c.inserter.Insert(ctx, created)
	if err != nil {
		return domain.Tag{}, fmt.Errorf("tag.Create: %w", err)
	}

	return created, nil
}
