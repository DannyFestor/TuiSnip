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

	_, err = io.WriteString(stdout, exportReport(path, counts))
	if err != nil {
		return fmt.Errorf("print export report: %w", err)
	}

	return nil
}

func exportReport(path string, counts library.ExportCounts) string {
	return fmt.Sprintf(
		"Exported %s, %s, %s to %s\n",
		pluralised(counts.Folders, "Folder"),
		pluralised(counts.Tags, "Tag"),
		pluralised(counts.Snippets, "Snippet"),
		path,
	)
}

func pluralised(count int, noun string) string {
	if count == 1 {
		return "1 " + noun
	}

	return fmt.Sprintf("%d %ss", count, noun)
}
