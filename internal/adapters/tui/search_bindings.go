package tui

import "charm.land/bubbles/v2/key"

const (
	labelMove   = "move"
	labelReveal = "reveal"
	labelClose  = "close"
)

type searchBindings struct {
	down   key.Binding
	up     key.Binding
	accept key.Binding
	copy   key.Binding
	cancel key.Binding
}

func newSearchBindings(search SearchKeyMap) searchBindings {
	return searchBindings{
		down:   labelled(search.Down, labelMove),
		up:     unlabelled(search.Up),
		accept: labelled(search.Accept, labelReveal),
		copy:   labelled(search.Copy, labelCopy),
		cancel: labelled(search.Cancel, labelClose),
	}
}

func (b searchBindings) hints() []key.Binding {
	return []key.Binding{b.down, b.accept, b.copy, b.cancel}
}
