package overlay

import tea "charm.land/bubbletea/v2"

type Step[O any] struct {
	next    Overlay[O]
	closes  bool
	child   Overlay[O]
	outcome O
	passes  bool
	cmd     tea.Cmd
}

func Stay[O any](next Overlay[O]) Step[O] {
	var none O

	return Step[O]{next: next, closes: false, child: nil, outcome: none, passes: false, cmd: nil}
}

func Close[O any]() Step[O] {
	var none O

	return Step[O]{next: nil, closes: true, child: nil, outcome: none, passes: false, cmd: nil}
}

func (s Step[O]) Opening(child Overlay[O]) Step[O] {
	s.child = child

	return s
}

func (s Step[O]) Passing(outcome O) Step[O] {
	s.outcome = outcome
	s.passes = true

	return s
}

func (s Step[O]) Running(cmd tea.Cmd) Step[O] {
	s.cmd = cmd

	return s
}
