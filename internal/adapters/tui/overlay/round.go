package overlay

import (
	"slices"

	tea "charm.land/bubbletea/v2"
)

type round struct {
	layers   []Overlay
	screen   tea.WindowSizeMsg
	outcomes []Outcome
	cmds     []tea.Cmd
}

func roundOver(stack Stack) *round {
	return &round{layers: slices.Clone(stack.layers), screen: stack.screen, outcomes: nil, cmds: nil}
}

func (r *round) routedToTop(msg tea.Msg) {
	top := len(r.layers) - 1
	if top >= 0 {
		r.applied(top, r.layers[top].Update(msg))
	}
}

func (r *round) delivered(msg tea.Msg) {
	for index := range slices.Backward(r.layers) {
		if index < len(r.layers) {
			r.applied(index, r.layers[index].Update(msg))
		}
	}
}

func (r *round) applied(index int, step Step) {
	r.cmds = append(r.cmds, step.cmd)
	r.replaced(index, step)

	if step.outcome != nil {
		r.bubbled(index-1, step.outcome)
	}
}

func (r *round) replaced(index int, step Step) {
	if step.closes {
		r.layers = r.layers[:index]

		return
	}

	r.layers[index] = step.next
	if step.child != nil {
		r.opened(index, step.child)
	}
}

func (r *round) opened(parent int, child Overlay) {
	r.layers = append(r.layers[:parent+1], child)
	r.applied(parent+1, child.Update(r.screen))
}

func (r *round) bubbled(index int, outcome Outcome) {
	if index < 0 {
		r.outcomes = append(r.outcomes, outcome)

		return
	}

	r.applied(index, r.layers[index].Received(outcome))
}

func (r *round) finished() (Stack, []Outcome, tea.Cmd) {
	return Stack{layers: r.layers, screen: r.screen}, r.outcomes, tea.Batch(r.cmds...)
}
