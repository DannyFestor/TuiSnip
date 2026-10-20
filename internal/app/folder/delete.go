package folder

import (
	"context"
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Delete struct {
	trasher Trasher
	clock   Clock
}

func NewDelete(trasher Trasher, clock Clock) (*Delete, error) {
	err := errors.Join(
		domain.RequireDependency("trasher", trasher),
		domain.RequireDependency("clock", clock),
	)
	if err != nil {
		return nil, fmt.Errorf("folder.NewDelete: %w", err)
	}

	return &Delete{trasher: trasher, clock: clock}, nil
}

func (d *Delete) Run(ctx context.Context, input DeleteInput) error {
	err := d.trasher.Trash(ctx, input.FolderID, d.clock.Now())
	if err != nil {
		return fmt.Errorf("folder.Delete: %w", err)
	}

	return nil
}
