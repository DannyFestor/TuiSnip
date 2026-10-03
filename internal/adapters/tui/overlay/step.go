package overlay

import (
	"slices"

	tea "charm.land/bubbletea/v2"
)

type Step[O any] struct {
	next     Overlay[O]
	closes   bool
	child    Overlay[O]
	outcomes []O
	cmd      tea.Cmd
}

func Stay[O any](next Overlay[O]) Step[O] {
	return Step[O]{next: next, closes: false, child: nil, outcomes: nil, cmd: nil}
}

func Close[O any]() Step[O] {
	return Step[O]{next: nil, closes: true, child: nil, outcomes: nil, cmd: nil}
}

func (s Step[O]) Opening(child Overlay[O]) Step[O] {
	s.child = child

	return s
}

func (s Step[O]) Passing(outcomes ...O) Step[O] {
	s.outcomes = slices.Concat(s.outcomes, outcomes)

	return s
}

func (s Step[O]) Running(cmd tea.Cmd) Step[O] {
	s.cmd = cmd

	return s
}
