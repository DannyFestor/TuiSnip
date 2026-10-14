package exportfile

import (
	"context"
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/adapters/atomicfile"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type Writer struct{}

func NewWriter() *Writer {
	return &Writer{}
}

func (*Writer) Write(ctx context.Context, path string, library domain.Library) error {
	err := replaceWithDocument(ctx, path, library)
	if err != nil {
		return fmt.Errorf("exportfile.Writer.Write: %w", err)
	}

	return nil
}

// Snippets often hold secrets, so the Export is readable by its owner only.
func replaceWithDocument(ctx context.Context, path string, library domain.Library) error {
	encoded, err := documentOf(library).encode()
	if err != nil {
		return err
	}

	err = ctx.Err()
	if err != nil {
		return fmt.Errorf("before writing: %w", err)
	}

	err = atomicfile.ReplaceOwnerOnly(path, encoded)
	if err != nil {
		return fmt.Errorf("replace export: %w", err)
	}

	return nil
}
