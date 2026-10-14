package main

import (
	"context"
	"fmt"
	"io"

	"github.com/DannyFestor/TuiSnip/internal/app/library"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
)

func exportLibrary(ctx context.Context, app *bootstrap.App, operands []string, stdout io.Writer) error {
	path := operands[0]

	counts, err := app.Export.Run(ctx, library.ExportInput{Path: path})
	if err != nil {
		return fmt.Errorf("export: %w", err)
	}

	_, err = io.WriteString(stdout, "Exported "+countsPhrase(counts)+" to "+path+"\n")
	if err != nil {
		return fmt.Errorf("print export report: %w", err)
	}

	return nil
}
