package mainscreen

import (
	"time"

	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/browseselection"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/editoverlay"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderoutline"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpane"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/folderpath"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/look"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/move"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/pointer"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/searchpopup"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/snippetlist"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/snippetpane"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/tagpane"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/trashlist"
	"github.com/DannyFestor/TuiSnip/internal/app/browse"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

type panes struct {
	folders    folderpane.Pane
	tags       tagpane.Pane
	list       snippetlist.List
	trash      trashlist.List
	preview    snippetpane.Pane
	outline    folderoutline.Pane
	trashShown bool
	styles     look.Styles
	paths      folderpath.Paths
	orderLabel string
}

func newPanes(keys binding.Keys, styles look.Styles, location *time.Location, collapsed []domain.FolderID) panes {
	return panes{
		folders:    folderpane.New(keys, styles, collapsed),
		tags:       tagpane.New(keys, styles),
		list:       snippetlist.New(keys, styles, snippetlist.Language),
		trash:      trashlist.New(keys, styles, snippetlist.Language),
		preview:    snippetpane.New(keys, styles, location),
		outline:    folderoutline.New(keys, styles, location),
		trashShown: false,
		styles:     styles,
		paths:      folderpath.Paths{},
		orderLabel: "",
	}
}

func (p panes) withOrderLabel(label string) panes {
	p.orderLabel = label

	return p
}

func (p panes) naming() bool {
	return p.folders.Naming() || p.tags.Naming()
}

func (p panes) updated(focus pane, msg tea.Msg) (panes, []outcome.Outcome, tea.Cmd) {
	if p.trashShown && focus.showsTrash() {
		return p.trashUpdated(focus, msg)
	}

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
		tags, outcomes, cmd := p.tags.Update(msg)
		p.tags = tags

		return p, outcomes, cmd
	}

	return p, nil, nil
}

func (p panes) hasRowAt(of pane, cell pointer.Point) bool {
	switch of {
	case paneFolders:
		return p.folders.HasRowAt(cell)
	case paneTags:
		return p.tags.HasRowAt(cell)
	case paneList:
		if p.trashShown {
			return p.trash.HasRowAt(cell)
		}

		return p.list.HasRowAt(cell)
	case paneSnippet:
	}

	return false
}

func (p panes) resized(arranged layout) panes {
	p.folders, _, _ = p.folders.Update(look.Resized{Box: arranged.folders.Inner()})
	p.tags, _, _ = p.tags.Update(look.Resized{Box: arranged.tags.Inner()})
	p.list, _, _ = p.list.Update(look.Resized{Box: arranged.list.Inner()})
	p.trash, _, _ = p.trash.Update(look.Resized{Box: arranged.list.Inner()})
	p.preview, _, _ = p.preview.Update(look.Resized{Box: arranged.snippet.Inner()})
	p.outline, _, _ = p.outline.Update(look.Resized{Box: arranged.snippet.Inner()})

	return p
}

func (p panes) restyled(msg look.Restyled) panes {
	p.folders, _, _ = p.folders.Update(msg)
	p.tags, _, _ = p.tags.Update(msg)
	p.list, _, _ = p.list.Update(msg)
	p.trash, _, _ = p.trash.Update(msg)
	p.preview, _, _ = p.preview.Update(msg)
	p.outline, _, _ = p.outline.Update(msg)
	p.styles = msg.Styles

	return p
}

func (p panes) withTree(tree browse.Tree) (panes, []outcome.Outcome) {
	var outcomes []outcome.Outcome

	p.folders, outcomes = p.folders.WithTree(tree)
	p.paths = folderpath.New(tree)
	p.preview = p.preview.WithPaths(p.paths)

	return p, outcomes
}

func (p panes) withTags(tags []browse.TagCount) panes {
	p.tags = p.tags.WithTags(tags)

	return p
}

func (p panes) withTagCursorOn(id domain.TagID) panes {
	p.tags = p.tags.WithCursorOn(id)

	return p
}

func (p panes) selectionIn(holder pane) browseselection.Selection {
	tagID, ok := p.tags.Selected()
	if holder == paneTags && ok {
		return browseselection.WithTag(tagID)
	}

	return p.folders.Selection()
}

func (p panes) destination(selection browseselection.Selection) editoverlay.Destination {
	filedIn := selection
	if selection.Trash() {
		filedIn = browseselection.InFolder(domain.FolderID{})
	}

	folderID, _ := filedIn.Folder()

	return editoverlay.Destination{
		Selection: filedIn,
		Tags:      p.tagsFor(filedIn),
		Language:  p.folders.DefaultLanguageOf(folderID),
	}
}

func (p panes) tagsFor(selection browseselection.Selection) []domain.Tag {
	tag, ok := p.tags.SelectedTag()
	if _, withTag := selection.Tag(); withTag && ok {
		return []domain.Tag{tag}
	}

	return nil
}

func (p panes) withSnippetsIfStillSelected(loaded SnippetsLoaded, selection browseselection.Selection) (panes, error) {
	if loaded.Selection != selection {
		return p, nil
	}

	p.list = p.list.WithSnippets(loaded.Snippets).Moved(move.Top).WithCursorOn(loaded.Selecting)

	label, err := orderLabel(loaded.Order)
	if err == nil {
		p = p.withOrderLabel(label)
	}

	return p.previewSelected().previewShowing(loaded.Showing), err
}

func (p panes) previewShowing(id domain.FragmentID) panes {
	p.preview = p.preview.ShowingFragment(id)

	return p
}

func (p panes) sortCycleAsked(selection browseselection.Selection) outcome.SortCycleAsked {
	selected, _ := p.list.Selected()

	return outcome.SortCycleAsked{Selection: selection, Selecting: selected.ID()}
}

func (p panes) selectingFolder(id domain.FolderID) (panes, []outcome.Outcome) {
	var outcomes []outcome.Outcome

	p.folders, outcomes = p.folders.WithCursorOn(id)

	return p, outcomes
}

func (p panes) selectionAndOrder(selection browseselection.Selection) string {
	if selection.Trash() {
		return trashTitle
	}

	return p.selectionTitle(selection) + listTitleSeparator + p.orderLabel
}

func (p panes) selectionTitle(selection browseselection.Selection) string {
	if _, ok := selection.Tag(); ok {
		return tagpane.TagPrefix + p.tags.SelectedName()
	}

	folderID, _ := selection.Folder()

	return p.paths.Full(folderID)
}

func (p panes) listing(selection browseselection.Selection) searchpopup.Listing {
	if selection.Trash() {
		return searchpopup.Listing{Snippets: nil, Paths: p.paths, EmptyQueryNote: searchSkipsTrashText}
	}

	return searchpopup.Listing{Snippets: p.list.Snippets(), Paths: p.paths, EmptyQueryNote: ""}
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
	if p.trashShown && of.showsTrash() {
		return p.trashBody(of, frame)
	}

	switch of {
	case paneFolders:
		return p.folders.View(frame)
	case paneTags:
		return p.tags.View(frame)
	case paneList:
		return p.list.View(frame)
	case paneSnippet:
		return p.preview.View()
	}

	return ""
}

func (p panes) keyMaps() map[pane]help.KeyMap {
	list := help.KeyMap(p.list)
	if p.trashShown {
		list = p.trash
	}

	return map[pane]help.KeyMap{
		paneFolders: p.folders,
		paneTags:    p.tags,
		paneList:    list,
		paneSnippet: p.preview,
	}
}
