package mainscreen

import (
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
)

type navigationBinding struct {
	name     string
	navigate navigation
}

func navigationFor(global binding.Set, msg tea.KeyPressMsg) (navigation, bool) {
	for _, candidate := range navigationBindings() {
		if global.Matches(msg, candidate.name) {
			return candidate.navigate, true
		}
	}

	return nil, false
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
