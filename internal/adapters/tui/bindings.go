package tui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

const (
	labelCopy       = "Copy"
	labelNewSnippet = "new Snippet"
	labelCapture    = "Capture"
	labelSearch     = "Search"
	labelNewFolder  = "new Folder"
	labelHelp       = "help"
)

type navigationBinding struct {
	binding  key.Binding
	navigate navigation
}

type movementBinding struct {
	binding key.Binding
	move    movement
}

type bindings struct {
	quit        key.Binding
	navigations []navigationBinding
	movements   []movementBinding
	listCopy    key.Binding
	paneCopy    key.Binding
	emptyHints  []key.Binding
}

func newBindings(settings Settings) bindings {
	return bindings{
		quit:        unlabelled(settings.Global.Quit),
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
	}
}

func (b bindings) navigationFor(msg tea.KeyPressMsg) (navigation, bool) {
	for _, candidate := range b.navigations {
		if key.Matches(msg, candidate.binding) {
			return candidate.navigate, true
		}
	}

	return nil, false
}

func (b bindings) movementFor(msg tea.KeyPressMsg) (movement, bool) {
	for _, candidate := range b.movements {
		if key.Matches(msg, candidate.binding) {
			return candidate.move, true
		}
	}

	return moveNone, false
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
		{binding: unlabelled(global.Down), move: moveDown},
		{binding: unlabelled(global.Up), move: moveUp},
		{binding: unlabelled(global.Top), move: moveTop},
		{binding: unlabelled(global.Bottom), move: moveBottom},
		{binding: unlabelled(global.PageDown), move: movePageDown},
		{binding: unlabelled(global.PageUp), move: movePageUp},
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
