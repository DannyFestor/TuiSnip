package library

import "github.com/DannyFestor/TuiSnip/internal/domain"

type Counts struct {
	Folders   int
	Tags      int
	Snippets  int
	Fragments int
	InTrash   TrashCounts
}

func countsOf(library domain.Library) Counts {
	fragments := 0
	for i := range library.Snippets {
		fragments += len(library.Snippets[i].Fragments())
	}

	return Counts{
		Folders:   len(library.Folders),
		Tags:      len(library.Tags),
		Snippets:  len(library.Snippets),
		Fragments: fragments,
		InTrash:   trashCountsOf(library),
	}
}
