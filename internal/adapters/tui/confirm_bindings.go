package tui

import "charm.land/bubbles/v2/key"

const (
	labelYes = "yes"
	labelNo  = "no"
)

type confirmBindings struct {
	yes key.Binding
	no  key.Binding
}

func newConfirmBindings(confirm ConfirmKeyMap) confirmBindings {
	return confirmBindings{
		yes: labelled(confirm.Yes, labelYes),
		no:  labelled(confirm.No, labelNo),
	}
}

func (b confirmBindings) hints() []key.Binding {
	return []key.Binding{b.yes, b.no}
}
