package main

import (
	"fmt"

	"github.com/DannyFestor/TuiSnip/internal/app/library"
)

func countsPhrase(counts library.Counts) string {
	return fmt.Sprintf(
		"%s, %s, %s",
		pluralised(counts.Folders, "Folder"),
		pluralised(counts.Tags, "Tag"),
		pluralised(counts.Snippets, "Snippet"),
	)
}

func pluralised(count int, noun string) string {
	if count == 1 {
		return "1 " + noun
	}

	return fmt.Sprintf("%d %ss", count, noun)
}
