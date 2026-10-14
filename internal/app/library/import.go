package library

import (
	"context"
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Import struct {
	reader   Reader
	inserter EmptyInserter
}

func NewImport(reader Reader, inserter EmptyInserter) (*Import, error) {
	err := errors.Join(
		domain.RequireDependency("reader", reader),
		domain.RequireDependency("inserter", inserter),
	)
	if err != nil {
		return nil, fmt.Errorf("library.NewImport: %w", err)
	}

	return &Import{reader: reader, inserter: inserter}, nil
}

func (i *Import) Run(ctx context.Context, input ImportInput) (Counts, error) {
	read, err := i.reader.Read(ctx, input.Path)
	if err != nil {
		return Counts{}, fmt.Errorf("library.Import: %w", err)
	}

	counts, err := insertWhole(ctx, writtenAsGiven(i.inserter.InsertIntoEmpty), read)
	if err != nil {
		return Counts{}, fmt.Errorf("library.Import: %w", err)
	}

	return counts, nil
}
