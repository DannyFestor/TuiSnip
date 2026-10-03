package tui

import (
	"slices"

	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
)

type navigationBinding struct {
	name     string
	navigate navigation
}

type bindings struct {
	table      binding.Keys
	forcedQuit string
	global     binding.Set
	folders    binding.Set
	tags       binding.Set
	editor     editorBindings
	confirm    binding.Set
}

func newBindings(settings Settings) bindings {
	keys := settings.Keys

	return bindings{
		table:      keys,
		forcedQuit: settings.ForcedQuitKey,
		global:     keys.For(binding.ScopeGlobal),
		folders:    keys.For(binding.ScopeFolders),
		tags:       keys.For(binding.ScopeTags),
		editor:     editorBindings{fields: keys.For(binding.ScopeEditor), content: keys.For(binding.ScopeContent)},
		confirm:    keys.For(binding.ScopeConfirm),
	}
}

func (b bindings) navigationFor(msg tea.KeyPressMsg) (navigation, bool) {
	for _, candidate := range navigationBindings() {
		if b.global.Matches(msg, candidate.name) {
			return candidate.navigate, true
		}
	}

	return nil, false
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
