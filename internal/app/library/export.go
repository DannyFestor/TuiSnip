package library

import (
	"context"
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Export struct {
	loader Loader
	writer Writer
}

func NewExport(loader Loader, writer Writer) (*Export, error) {
	err := errors.Join(
		domain.RequireDependency("loader", loader),
		domain.RequireDependency("writer", writer),
	)
	if err != nil {
		return nil, fmt.Errorf("library.NewExport: %w", err)
	}

	return &Export{loader: loader, writer: writer}, nil
}

func (e *Export) Run(ctx context.Context, input ExportInput) (Counts, error) {
	loaded, err := e.loader.Load(ctx)
	if err != nil {
		return Counts{}, fmt.Errorf("library.Export: %w", err)
	}

	ordered := loaded.InCreatedOrder()

	err = e.writer.Write(ctx, input.Path, ordered)
	if err != nil {
		return Counts{}, fmt.Errorf("library.Export: %w", err)
	}

	return countsOf(ordered), nil
}
