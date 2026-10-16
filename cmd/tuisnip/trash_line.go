package main

import "github.com/DannyFestor/TuiSnip/internal/app/library"

func trashLine(counts library.TrashCounts) string {
	if counts.IsEmpty() {
		return ""
	}

	return "In the Trash: " + pluralised(
		counts.Folders,
		"Folder",
	) + ", " + pluralised(
		counts.Snippets,
		"Snippet",
	) + "\n"
}
