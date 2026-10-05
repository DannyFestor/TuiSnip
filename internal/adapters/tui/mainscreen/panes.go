package mainscreen

import (
	"time"

	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpane"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpath"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/move"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/searchpopup"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/snippetlist"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/snippetpane"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tagpane"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type panes struct {
	folders    folderpane.Pane
	tags       tagpane.Pane
	list       snippetlist.List
	preview    snippetpane.Pane
	paths      folderpath.Paths
	orderLabel string
}

func newPanes(keys binding.Keys, styles look.Styles, location *time.Location, collapsed []domain.FolderID) panes {
	return panes{
		folders:    folderpane.New(keys, collapsed),
		tags:       tagpane.New(keys, styles),
		list:       snippetlist.New(keys, styles, snippetlist.Language),
		preview:    snippetpane.New(keys, styles, location),
		paths:      folderpath.Paths{},
		orderLabel: "",
	}
}

func (p panes) withOrderLabel(label string) panes {
	p.orderLabel = label

	return p
}

func (p panes) naming() bool {
	return p.folders.Naming()
}

func (p panes) updated(focus pane, msg tea.Msg) (panes, []outcome.Outcome, tea.Cmd) {
	switch focus {
	case paneList:
		list, outcomes, cmd := p.list.Update(msg)
		p.list = list

		return p.previewSelected(), outcomes, cmd
	case paneSnippet:
		preview, outcomes, cmd := p.preview.Update(msg)
		p.preview = preview

		return p, outcomes, cmd
	case paneFolders:
		folders, outcomes, cmd := p.folders.Update(msg)
		p.folders = folders

		return p, outcomes, cmd
	case paneTags:
	}

	return p, nil, nil
}

func (p panes) resized(to layout) panes {
	p.folders, _, _ = p.folders.Update(look.Resized{Box: to.folders.Inner()})
	p.tags, _, _ = p.tags.Update(look.Resized{Box: to.tags.Inner()})
	p.list, _, _ = p.list.Update(look.Resized{Box: to.list.Inner()})
	p.preview, _, _ = p.preview.Update(look.Resized{Box: to.snippet.Inner()})

	return p
}

func (p panes) withBackground(msg tea.BackgroundColorMsg) panes {
	p.preview, _, _ = p.preview.Update(msg)

	return p
}

func (p panes) withTree(tree browse.Tree) (panes, []outcome.Outcome) {
	var outcomes []outcome.Outcome

	p.folders, outcomes = p.folders.WithTree(tree)
	p.paths = folderpath.New(tree)
	p.preview = p.preview.WithPaths(p.paths)

	return p, outcomes
}

func (p panes) withSnippetsIfStillSelected(loaded SnippetsLoaded) (panes, error) {
	if loaded.FolderID != p.folders.Selected() {
		return p, nil
	}

	p.list = p.list.WithSnippets(loaded.Snippets).Moved(move.Top).WithCursorOn(loaded.Selecting)

	label, err := orderLabel(loaded.Order)
	if err == nil {
		p = p.withOrderLabel(label)
	}

	return p.previewSelected(), err
}

func (p panes) sortCycleAsked() outcome.SortCycleAsked {
	selected, _ := p.list.Selected()

	return outcome.SortCycleAsked{FolderID: p.folders.Selected(), Selecting: selected.ID()}
}

func (p panes) selectingFolder(id domain.FolderID) (panes, []outcome.Outcome) {
	var outcomes []outcome.Outcome

	p.folders, outcomes = p.folders.WithCursorOn(id)

	return p, outcomes
}

func (p panes) selectionAndOrder() string {
	return p.paths.Full(p.folders.Selected()) + listTitleSeparator + p.orderLabel
}

func (p panes) listing() searchpopup.Listing {
	return searchpopup.Listing{Snippets: p.list.Snippets(), Paths: p.paths}
}

func (p panes) previewSelected() panes {
	selected, ok := p.list.Selected()
	if ok {
		p.preview = p.preview.Showing(selected)
	} else {
		p.preview = p.preview.Cleared()
	}

	return p
}

func (p panes) body(of pane, frame look.FrameStyle) string {
	switch of {
	case paneFolders:
		return p.folders.View(frame)
	case paneTags:
		return p.tags.View()
	case paneList:
		return p.list.View(frame)
	case paneSnippet:
		return p.preview.View()
	}

	return ""
}

func (p panes) keyMaps() map[pane]help.KeyMap {
	return map[pane]help.KeyMap{
		paneFolders: p.folders,
		paneTags:    p.tags,
		paneList:    p.list,
		paneSnippet: p.preview,
	}
}
