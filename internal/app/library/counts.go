package library

import "github.com/DannyFestor/TuiSnip/internal/domain"

type Counts struct {
	Folders  int
	Tags     int
	Snippets int
}

func countsOf(library domain.Library) Counts {
	return Counts{
		Folders:  len(library.Folders),
		Tags:     len(library.Tags),
		Snippets: len(library.Snippets),
	}
}
