package library

import (
	"context"
	"errors"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type ImportSnippetsLab struct {
	reader   SnippetsLabReader
	inserter AlongsideInserter
}

func NewImportSnippetsLab(reader SnippetsLabReader, inserter AlongsideInserter) (*ImportSnippetsLab, error) {
	err := errors.Join(
		domain.RequireDependency("reader", reader),
		domain.RequireDependency("inserter", inserter),
	)
	if err != nil {
		return nil, fmt.Errorf("library.NewImportSnippetsLab: %w", err)
	}

	return &ImportSnippetsLab{reader: reader, inserter: inserter}, nil
}

func (i *ImportSnippetsLab) Run(ctx context.Context, input ImportSnippetsLabInput) (SnippetsLabReport, error) {
	read, losses, err := i.reader.Read(ctx, input.Path)
	if err != nil {
		return SnippetsLabReport{}, fmt.Errorf("library.ImportSnippetsLab: %w", err)
	}

	counts, err := insertWhole(ctx, i.inserter.InsertAlongside, read)
	if err != nil {
		return SnippetsLabReport{}, fmt.Errorf("library.ImportSnippetsLab: %w", err)
	}

	return SnippetsLabReport{Counts: counts, Losses: losses}, nil
}
