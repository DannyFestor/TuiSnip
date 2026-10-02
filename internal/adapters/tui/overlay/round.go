package overlay

import (
	"slices"

	tea "charm.land/bubbletea/v2"
)

type round struct {
	overlays []Overlay
	screen   tea.WindowSizeMsg
	outcomes []Outcome
	cmds     []tea.Cmd
}

func roundOver(stack Stack) *round {
	return &round{overlays: slices.Clone(stack.overlays), screen: stack.screen, outcomes: nil, cmds: nil}
}

func (r *round) routeToTop(msg tea.Msg) {
	top := len(r.overlays) - 1
	if top >= 0 {
		r.apply(top, r.overlays[top].Update(msg))
	}
}

func (r *round) deliver(msg tea.Msg) {
	for index := range slices.Backward(r.overlays) {
		if r.stillOpen(index) {
			r.apply(index, r.overlays[index].Update(msg))
		}
	}
}

func (r *round) apply(index int, step Step) {
	r.cmds = append(r.cmds, step.cmd)
	r.replace(index, step)

	if step.outcome != nil {
		r.bubble(index-1, step.outcome)
	}
}

func (r *round) replace(index int, step Step) {
	if step.closes {
		r.overlays = r.overlays[:index]

		return
	}

	r.overlays[index] = step.next
	if step.child != nil {
		r.open(index, step.child)
	}
}

func (r *round) open(parent int, child Overlay) {
	r.overlays = append(r.overlays[:parent+1], child)
	r.apply(parent+1, child.Update(r.screen))
}

func (r *round) bubble(index int, outcome Outcome) {
	if index < 0 {
		r.outcomes = append(r.outcomes, outcome)

		return
	}

	parent, ok := r.overlays[index].(Parent)
	if !ok {
		r.bubble(index-1, outcome)

		return
	}

	r.apply(index, parent.Received(outcome))
}

func (r *round) stillOpen(index int) bool {
	return index < len(r.overlays)
}

func (r *round) finish() (Stack, []Outcome, tea.Cmd) {
	return Stack{overlays: r.overlays, screen: r.screen}, r.outcomes, tea.Batch(r.cmds...)
}
