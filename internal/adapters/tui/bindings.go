package tui

import (
	"slices"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/move"
)

type navigationBinding struct {
	name     string
	navigate navigation
}

type movementBinding struct {
	name string
	move move.Direction
}

type bindings struct {
	forcedQuit  string
	global      binding.Set
	folders     binding.Set
	tags        binding.Set
	snippetList binding.Set
	snippetPane binding.Set
	editor      editorBindings
	search      binding.Set
	confirm     binding.Set
	emptyHints  []key.Binding
}

func newBindings(settings Settings) bindings {
	keys := settings.Keys

	return bindings{
		forcedQuit:  settings.ForcedQuitKey,
		global:      keys.For(binding.ScopeGlobal),
		folders:     keys.For(binding.ScopeFolders),
		tags:        keys.For(binding.ScopeTags),
		snippetList: keys.For(binding.ScopeSnippetList),
		snippetPane: keys.For(binding.ScopeSnippetPane),
		editor:      editorBindings{fields: keys.For(binding.ScopeEditor), content: keys.For(binding.ScopeContent)},
		search:      keys.For(binding.ScopeSearch),
		confirm:     keys.For(binding.ScopeConfirm),
		emptyHints:  keys.EmptyListHints(),
	}
}

func (b bindings) forPane(focus pane) binding.Set {
	switch focus {
	case paneFolders:
		return b.folders
	case paneTags:
		return b.tags
	case paneList:
		return b.snippetList
	case paneSnippet:
		return b.snippetPane
	}

	return b.folders
}

func (b bindings) navigationFor(msg tea.KeyPressMsg) (navigation, bool) {
	for _, candidate := range navigationBindings() {
		if b.global.Matches(msg, candidate.name) {
			return candidate.navigate, true
		}
	}

	return nil, false
}

func (b bindings) movementFor(msg tea.KeyPressMsg) (move.Direction, bool) {
	for _, candidate := range movementBindings() {
		if b.global.Matches(msg, candidate.name) {
			return candidate.move, true
		}
	}

	return move.None, false
}

func boundOnly(firstKeys ...string) []string {
	return slices.DeleteFunc(firstKeys, func(firstKey string) bool { return firstKey == "" })
}

func navigationBindings() []navigationBinding {
	return []navigationBinding{
		{name: binding.FocusNext, navigate: pane.next},
		{name: binding.FocusPrev, navigate: pane.prev},
		{name: binding.FocusRight, navigate: pane.right},
		{name: binding.FocusLeft, navigate: pane.left},
		{name: binding.FocusFolders, navigate: focusOn(paneFolders)},
		{name: binding.FocusTags, navigate: focusOn(paneTags)},
		{name: binding.FocusList, navigate: focusOn(paneList)},
		{name: binding.FocusSnippet, navigate: focusOn(paneSnippet)},
		{name: binding.Open, navigate: pane.drillIn},
		{name: binding.Back, navigate: pane.backOut},
	}
}

func movementBindings() []movementBinding {
	return []movementBinding{
		{name: binding.Down, move: move.Down},
		{name: binding.Up, move: move.Up},
		{name: binding.Top, move: move.Top},
		{name: binding.Bottom, move: move.Bottom},
		{name: binding.PageDown, move: move.PageDown},
		{name: binding.PageUp, move: move.PageUp},
	}
}
