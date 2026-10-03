package tui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/move"
)

const (
	labelCopy       = "Copy"
	labelNewSnippet = "new Snippet"
	labelCapture    = "Capture"
	labelSearch     = "Search"
	labelNewFolder  = "new Folder"
	labelHelp       = "help"
	labelNew        = "new"
	labelSearchHint = "search"
)

type navigationBinding struct {
	binding  key.Binding
	navigate navigation
}

type movementBinding struct {
	binding key.Binding
	move    move.Direction
}

type bindings struct {
	quit        key.Binding
	newSnippet  key.Binding
	openSearch  key.Binding
	navigations []navigationBinding
	movements   []movementBinding
	listCopy    key.Binding
	paneCopy    key.Binding
	emptyHints  []key.Binding
	editor      editorBindings
	search      searchBindings
	confirm     confirmBindings
}

func newBindings(settings Settings) bindings {
	return bindings{
		quit:        unlabelled(settings.Global.Quit),
		newSnippet:  labelled(settings.Global.NewSnippet, labelNew),
		openSearch:  labelled(settings.Global.Search, labelSearchHint),
		navigations: navigationBindings(settings.Global),
		movements:   movementBindings(settings.Global),
		listCopy:    labelled(settings.SnippetList.Copy, labelCopy),
		paneCopy:    labelled(settings.SnippetPane.Copy, labelCopy),
		emptyHints: []key.Binding{
			labelled(settings.Global.NewSnippet, labelNewSnippet),
			labelled(settings.Global.Capture, labelCapture),
			labelled(settings.Global.Search, labelSearch),
			labelled(settings.Folders.NewFolder, labelNewFolder),
			labelled(settings.Global.Help, labelHelp),
		},
		editor:  newEditorBindings(settings.Editor, settings.Content),
		search:  newSearchBindings(settings.Search),
		confirm: newConfirmBindings(settings.Confirm),
	}
}

func (b bindings) paneHints(focus pane) []key.Binding {
	switch focus {
	case paneList:
		return []key.Binding{b.listCopy, b.newSnippet, b.openSearch}
	case paneSnippet:
		return []key.Binding{b.paneCopy, b.openSearch}
	case paneFolders, paneTags:
	}

	return []key.Binding{b.openSearch}
}

func (b bindings) navigationFor(msg tea.KeyPressMsg) (navigation, bool) {
	for _, candidate := range b.navigations {
		if key.Matches(msg, candidate.binding) {
			return candidate.navigate, true
		}
	}

	return nil, false
}

func (b bindings) movementFor(msg tea.KeyPressMsg) (move.Direction, bool) {
	for _, candidate := range b.movements {
		if key.Matches(msg, candidate.binding) {
			return candidate.move, true
		}
	}

	return move.None, false
}

func navigationBindings(global GlobalKeyMap) []navigationBinding {
	return []navigationBinding{
		{binding: unlabelled(global.FocusNext), navigate: pane.next},
		{binding: unlabelled(global.FocusPrev), navigate: pane.prev},
		{binding: unlabelled(global.FocusRight), navigate: pane.right},
		{binding: unlabelled(global.FocusLeft), navigate: pane.left},
		{binding: unlabelled(global.FocusFolders), navigate: focusOn(paneFolders)},
		{binding: unlabelled(global.FocusTags), navigate: focusOn(paneTags)},
		{binding: unlabelled(global.FocusList), navigate: focusOn(paneList)},
		{binding: unlabelled(global.FocusSnippet), navigate: focusOn(paneSnippet)},
		{binding: unlabelled(global.Open), navigate: pane.drillIn},
		{binding: unlabelled(global.Back), navigate: pane.backOut},
	}
}

func movementBindings(global GlobalKeyMap) []movementBinding {
	return []movementBinding{
		{binding: unlabelled(global.Down), move: move.Down},
		{binding: unlabelled(global.Up), move: move.Up},
		{binding: unlabelled(global.Top), move: move.Top},
		{binding: unlabelled(global.Bottom), move: move.Bottom},
		{binding: unlabelled(global.PageDown), move: move.PageDown},
		{binding: unlabelled(global.PageUp), move: move.PageUp},
	}
}

func unlabelled(keys []string) key.Binding {
	return key.NewBinding(key.WithKeys(keys...))
}

func labelled(keys []string, label string) key.Binding {
	binding := unlabelled(keys)
	if len(keys) == 0 {
		binding.SetEnabled(false)

		return binding
	}

	binding.SetHelp(keys[0], label)

	return binding
}
