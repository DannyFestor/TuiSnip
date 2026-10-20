package mainscreen

import (
	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/searchpopup"
	"github.com/DannyFestor/TuiSnip/internal/app/trash"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type browsed int

const (
	browsedSnippets browsed = iota
	browsedTrash
)

func browsedBy(selection browseselection.Selection) browsed {
	if selection.Trash() {
		return browsedTrash
	}

	return browsedSnippets
}

func (b browsed) pressed(s Screen, msg tea.KeyPressMsg) (outcome.Step, bool) {
	if b == browsedTrash {
		return s.trashPressed(msg)
	}

	return s.snippetsPressed(msg)
}

func (b browsed) saved(s Screen, saved outcome.SnippetSaved) outcome.Step {
	if b == browsedTrash {
		return s.savedFromTrash(saved)
	}

	return outcome.Stay(s).Passing(saved)
}

func (b browsed) updated(p panes, focus pane, msg tea.Msg) (panes, []outcome.Outcome, tea.Cmd) {
	if b == browsedTrash {
		return p.trashUpdated(focus, msg)
	}

	return p.snippetsUpdated(focus, msg)
}

func (b browsed) listHasRowAt(p panes, cell pointer.Point) bool {
	if b == browsedTrash {
		return p.trash.HasRowAt(cell)
	}

	return p.list.HasRowAt(cell)
}

func (b browsed) body(p panes, framed pane, frame look.FrameStyle) string {
	if b == browsedTrash {
		return p.trashBody(framed, frame)
	}

	return p.snippetsBody(framed, frame)
}

func (b browsed) keyMaps(shown panes) map[pane]help.KeyMap {
	if b == browsedTrash {
		return shown.trashKeyMaps()
	}

	return shown.keyMaps(shown.list, shown.preview)
}

func (b browsed) listTitle(p panes, selection browseselection.Selection) string {
	if b == browsedTrash {
		return trashTitle
	}

	return p.selectionTitle(selection) + listTitleSeparator + p.orderLabel
}

func (b browsed) listing(p panes) searchpopup.Listing {
	if b == browsedTrash {
		return searchpopup.Listing{Snippets: nil, Paths: p.paths, EmptyQueryNote: searchSkipsTrashText}
	}

	return searchpopup.Listing{Snippets: p.list.Snippets(), Paths: p.paths, EmptyQueryNote: ""}
}

func (b browsed) filedIn(selection browseselection.Selection) browseselection.Selection {
	if b == browsedTrash {
		return browseselection.InFolder(domain.FolderID{})
	}

	return selection
}

func (b browsed) listedSnippetID(p panes) domain.SnippetID {
	if b == browsedTrash {
		return domain.SnippetID{}
	}

	listed, _ := p.list.Selected()

	return listed.ID()
}

func (b browsed) trashLoaded(p panes, items []trash.Item) panes {
	if b == browsedTrash {
		return p.withTrash(items)
	}

	return p
}
