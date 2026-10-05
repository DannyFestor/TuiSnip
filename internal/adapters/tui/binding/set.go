package binding

import (
	"slices"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

type Set struct {
	bound  map[string]key.Binding
	hints  []key.Binding
	listed []key.Binding
}

func (s Set) Matches(msg tea.KeyPressMsg, name string) bool {
	return key.Matches(msg, s.bound[name])
}

func (s Set) FirstKey(name string) string {
	keys := s.bound[name].Keys()
	if len(keys) == 0 {
		return ""
	}

	return keys[0]
}

func (s Set) ShortHelp() []key.Binding {
	return slices.Clone(s.hints)
}

func (s Set) FullHelp() [][]key.Binding {
	return [][]key.Binding{slices.Clone(s.listed)}
}
