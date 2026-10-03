package overlay

import (
	"slices"

	tea "charm.land/bubbletea/v2"
)

type round[O any] struct {
	overlays []Overlay[O]
	screen   tea.WindowSizeMsg
	outcomes []O
	cmds     []tea.Cmd
}

func roundOver[O any](stack Stack[O]) *round[O] {
	return &round[O]{overlays: slices.Clone(stack.overlays), screen: stack.screen, outcomes: nil, cmds: nil}
}

func (r *round[O]) routeToTop(msg tea.Msg) {
	top := len(r.overlays) - 1
	if top >= 0 {
		r.apply(top, r.overlays[top].Update(msg))
	}
}

func (r *round[O]) deliver(msg tea.Msg) {
	for index := range slices.Backward(r.overlays) {
		if r.stillOpen(index) {
			r.apply(index, r.overlays[index].Update(msg))
		}
	}
}

func (r *round[O]) apply(index int, step Step[O]) {
	r.cmds = append(r.cmds, step.cmd)
	r.replace(index, step)

	if step.passes {
		r.bubble(index-1, step.outcome)
	}
}

func (r *round[O]) replace(index int, step Step[O]) {
	if step.closes {
		r.overlays = r.overlays[:index]

		return
	}

	r.overlays[index] = step.next
	if step.child != nil {
		r.open(index, step.child)
	}
}

func (r *round[O]) open(parent int, child Overlay[O]) {
	r.overlays = append(r.overlays[:parent+1], child)
	r.apply(parent+1, child.Update(r.screen))
}

func (r *round[O]) bubble(index int, outcome O) {
	if index < 0 {
		r.outcomes = append(r.outcomes, outcome)

		return
	}

	parent, ok := r.overlays[index].(Parent[O])
	if !ok {
		r.bubble(index-1, outcome)

		return
	}

	r.apply(index, parent.Received(outcome))
}

func (r *round[O]) stillOpen(index int) bool {
	return index < len(r.overlays)
}

func (r *round[O]) finish() (Stack[O], []O, tea.Cmd) {
	return Stack[O]{overlays: r.overlays, screen: r.screen}, r.outcomes, tea.Batch(r.cmds...)
}
