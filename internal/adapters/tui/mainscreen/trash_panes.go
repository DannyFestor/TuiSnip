package mainscreen

import (
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

func (p pane) showsTrash() bool {
	return p == paneList || p == paneSnippet
}

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

func (p panes) showingTrash(shown bool) panes {
	next := p
	next.trashShown = shown

	return next
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

	if item.Kind == trash.ItemKindFolder {
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

	return ok && item.Kind == trash.ItemKindFolder
}

func (p panes) trashBody(of pane, frame look.FrameStyle) string {
	_, ok := p.trash.Selected()

	switch {
	case of == paneList:
		return p.trash.View(frame)
	case !ok:
		return look.EmptyTrash(p.styles)
	case p.selectedTrashFolder():
		return p.outline.View()
	}

	return p.preview.View()
}
