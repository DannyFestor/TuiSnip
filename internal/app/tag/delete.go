package tag

import (
	"context"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Delete struct {
	deleter Deleter
}

func NewDelete(deleter Deleter) (*Delete, error) {
	err := domain.RequireDependency("deleter", deleter)
	if err != nil {
		return nil, fmt.Errorf("tag.NewDelete: %w", err)
	}

	return &Delete{deleter: deleter}, nil
}

func (d *Delete) Run(ctx context.Context, input DeleteInput) error {
	err := d.deleter.Delete(ctx, input.TagID)
	if err != nil {
		return fmt.Errorf("tag.Delete: %w", err)
	}

	return nil
}
