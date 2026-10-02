package overlay

import tea "charm.land/bubbletea/v2"

type Step struct {
	next    Overlay
	closes  bool
	child   Overlay
	outcome Outcome
	cmd     tea.Cmd
}

func Stay(next Overlay) Step {
	return Step{next: next, closes: false, child: nil, outcome: nil, cmd: nil}
}

func Close() Step {
	return Step{next: nil, closes: true, child: nil, outcome: nil, cmd: nil}
}

func (s Step) Opening(child Overlay) Step {
	s.child = child

	return s
}

func (s Step) Passing(outcome Outcome) Step {
	s.outcome = outcome

	return s
}

func (s Step) Running(cmd tea.Cmd) Step {
	s.cmd = cmd

	return s
}
