package mainscreen

import (
	"time"

	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpane"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/snippetlist"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/snippetpane"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tagpane"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type panes struct {
	folders folderpane.Pane
	tags    tagpane.Pane
	list    snippetlist.List
	preview snippetpane.Pane
}

func newPanes(keys binding.Keys, styles look.Styles, location *time.Location) panes {
	return panes{
		folders: folderpane.New(keys),
		tags:    tagpane.New(keys, styles),
		list:    snippetlist.New(keys, styles, snippetlist.Language),
		preview: snippetpane.New(keys, styles, location),
	}
}

func (p panes) pressed(focus pane, msg tea.KeyPressMsg) (panes, []outcome.Outcome, tea.Cmd) {
	switch focus {
	case paneList:
		list, outcomes, cmd := p.list.Update(msg)
		p.list = list

		return p.previewSelected(), outcomes, cmd
	case paneSnippet:
		preview, outcomes, cmd := p.preview.Update(msg)
		p.preview = preview

		return p, outcomes, cmd
	case paneFolders, paneTags:
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

func (p panes) withSnippets(snippets []domain.Snippet, selecting domain.SnippetID) panes {
	next := p
	next.list = p.list.WithSnippets(snippets).WithCursorOn(selecting)
	next.folders = p.folders.WithRootSnippetCount(len(snippets))

	return next.previewSelected()
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
