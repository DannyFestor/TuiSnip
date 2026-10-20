package mainscreen

import (
	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpath"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/snippetpane"
	"github.com/DannyFestor/TuiSnip/internal/app/trash"
)

const (
	trashTitle           = "Trash"
	searchSkipsTrashText = "Search doesn't cover the Trash"
)

func (p panes) trashUpdated(focus pane, msg tea.Msg) (panes, []outcome.Outcome, tea.Cmd) {
	var (
		outcomes []outcome.Outcome
		cmd      tea.Cmd
	)

	switch {
	case focus == paneList:
		p.trash, outcomes, cmd = p.trash.Update(msg)

		return p.previewTrashSelected(), outcomes, cmd
	case p.selectedTrashFolder():
		p.outline, outcomes, cmd = p.outline.Update(msg)
	default:
		p.preview, outcomes, cmd = p.preview.Update(msg)
	}

	return p, outcomes, cmd
}

func (p panes) trashBody(framed pane, frame look.FrameStyle) string {
	_, ok := p.trash.Selected()

	switch {
	case framed == paneList:
		return p.trash.View(frame)
	case !ok:
		return look.EmptyTrash(p.styles)
	case p.selectedTrashFolder():
		return p.outline.View()
	}

	return p.preview.View()
}

func (p panes) trashKeyMaps() map[pane]help.KeyMap {
	if p.selectedTrashFolder() {
		return p.keyMaps(p.trash, p.outline)
	}

	return p.keyMaps(p.trash, p.preview)
}

func (p panes) withTrash(items []trash.Item) panes {
	next := p
	next.trash = p.trash.WithItems(items)

	return next.previewTrashSelected()
}

func (p panes) previewTrashSelected() panes {
	item, ok := p.trash.Selected()
	if !ok {
		p.preview = p.preview.Cleared()

		return p
	}

	if item.Kind.IsFolder() {
		p.outline = p.outline.Showing(item, folderpath.FullOf(item.Path))

		return p
	}

	p.preview = p.preview.ShowingInTrash(item.Snippet, snippetpane.Trashed{
		Place:     folderpath.FullOf(item.Path),
		DeletedAt: item.DeletedAt,
	})

	return p
}

func (p panes) selectedTrashFolder() bool {
	item, ok := p.trash.Selected()

	return ok && item.Kind.IsFolder()
}
