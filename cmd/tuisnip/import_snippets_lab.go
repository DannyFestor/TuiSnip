package main

import (
	"context"
	"fmt"
	"io"

	"github.com/DannyFestor/TuiSnip/internal/app/library"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
)

func importSnippetsLab(ctx context.Context, app *bootstrap.App, operands []string, stdout io.Writer) error {
	report, err := app.ImportSnippetsLab.Run(ctx, library.ImportSnippetsLabInput{Path: operands[0]})
	if err != nil {
		return fmt.Errorf("import: %w", err)
	}

	_, err = io.WriteString(stdout, snippetsLabReportText(report))
	if err != nil {
		return fmt.Errorf("print import report: %w", err)
	}

	return nil
}
