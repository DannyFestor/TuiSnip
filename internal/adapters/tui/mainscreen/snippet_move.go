package mainscreen

import (
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type snippetMove struct {
	moving    domain.Snippet
	successor domain.Snippet
	selection browseselection.Selection
}

func (m snippetMove) requested(picked domain.FolderID) outcome.SnippetMoveRequested {
	return outcome.SnippetMoveRequested{
		Input:     snippet.MoveInput{SnippetID: m.moving.ID(), FolderID: picked},
		Selection: m.selection,
		Selecting: m.selectedAfter(picked).ID(),
	}
}

func (m snippetMove) selectedAfter(picked domain.FolderID) domain.Snippet {
	if m.leavesList(picked) {
		return m.successor
	}

	return m.moving
}

func (m snippetMove) leavesList(picked domain.FolderID) bool {
	listed, inFolder := m.selection.Folder()

	return inFolder && listed != picked
}
