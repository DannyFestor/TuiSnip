package library

import "github.com/DannyFestor/TuiSnip/internal/domain"

type TrashCounts struct {
	Folders  int
	Snippets int
}

func trashCountsOf(library domain.Library) TrashCounts {
	inTrash := library.InTrash()

	return TrashCounts{Folders: len(inTrash.Folders), Snippets: len(inTrash.Snippets)}
}

func (c TrashCounts) IsEmpty() bool {
	return c.Folders == 0 && c.Snippets == 0
}
