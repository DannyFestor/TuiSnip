package main

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/DannyFestor/TuiSnip/internal/app/library"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

func importLibrary(ctx context.Context, app *bootstrap.App, operands []string, stdout io.Writer) error {
	counts, err := app.Import.Run(ctx, library.ImportInput{Path: operands[0]})
	if err != nil {
		return fmt.Errorf("import: %w", withEmptyLibraryHint(err))
	}

	_, err = io.WriteString(stdout, "Imported "+countsPhrase(counts)+"\n")
	if err != nil {
		return fmt.Errorf("print import report: %w", err)
	}

	return nil
}

func withEmptyLibraryHint(err error) error {
	if !errors.Is(err, domain.ErrLibraryNotEmpty) {
		return err
	}

	return fmt.Errorf("%w. Move the database aside first (%s --paths shows it), then import again", err, programName)
}
