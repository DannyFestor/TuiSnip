package trash

import (
	"context"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type List struct {
	loader Loader
}

func NewList(loader Loader) (*List, error) {
	err := domain.RequireDependency("loader", loader)
	if err != nil {
		return nil, fmt.Errorf("trash.NewList: %w", err)
	}

	return &List{loader: loader}, nil
}

func (l *List) Run(ctx context.Context, _ ListInput) ([]Item, error) {
	loaded, err := l.loader.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("trash.List: %w", err)
	}

	return shelfOf(loaded).items(), nil
}
