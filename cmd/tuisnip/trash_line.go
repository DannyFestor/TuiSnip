package main

import (
	"strings"

	"github.com/DannyFestor/TuiSnip/internal/app/library"
)

func trashLine(counts library.TrashCounts) string {
	if counts.IsEmpty() {
		return ""
	}

	return "In the Trash: " + strings.Join(heldCounts(counts), ", ") + "\n"
}

func heldCounts(counts library.TrashCounts) []string {
	var held []string

	if counts.Folders > 0 {
		held = append(held, pluralised(counts.Folders, "Folder"))
	}

	if counts.Snippets > 0 {
		held = append(held, pluralised(counts.Snippets, "Snippet"))
	}

	return held
}
