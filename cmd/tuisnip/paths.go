package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/DannyFestor/TuiSnip/internal/adapters/xdg"
	"github.com/DannyFestor/TuiSnip/internal/bootstrap"
)

const pathLabelWidth = len("config") + 3

func printPaths(stdout io.Writer, paths xdg.Paths) error {
	_, err := io.WriteString(stdout, pathTable(paths))
	if err != nil {
		return fmt.Errorf("print paths: %w", err)
	}

	return nil
}

func pathTable(paths xdg.Paths) string {
	rows := [][2]string{
		{"config", paths.ConfigFile},
		{"data", paths.DatabaseFile},
		{"backup", bootstrap.BackupDir(paths)},
		{"state", paths.StateFile},
		{"log", paths.LogFile},
	}

	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, fmt.Sprintf("%-*s%s\n", pathLabelWidth, row[0], row[1]))
	}

	return strings.Join(lines, "")
}
