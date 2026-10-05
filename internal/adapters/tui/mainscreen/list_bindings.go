package mainscreen

import (
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/confirm"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
)

func (s Screen) listPressed(msg tea.KeyPressMsg) (outcome.Step, bool) {
	if s.focus != paneList {
		return outcome.Stay(s), false
	}

	switch {
	case s.listKeys.Matches(msg, binding.CycleSort):
		return outcome.Stay(s).Passing(s.panes.sortCycleAsked(s.selection())), true
	case s.listKeys.Matches(msg, binding.Duplicate):
		return s.duplicateAsked(), true
	case s.listKeys.Matches(msg, binding.Delete):
		return s.deleteAsked(), true
	}

	return outcome.Stay(s), false
}

func (s Screen) duplicateAsked() outcome.Step {
	selected, ok := s.panes.list.Selected()
	if !ok {
		return outcome.Stay(s)
	}

	return outcome.Stay(s).Passing(outcome.SnippetDuplicateRequested{
		Input:     snippet.DuplicateInput{SnippetID: selected.ID()},
		Selection: s.selection(),
	})
}

func (s Screen) deleteAsked() outcome.Step {
	selected, ok := s.panes.list.Selected()
	if !ok {
		return outcome.Stay(s)
	}

	successor, _ := s.panes.list.Successor()
	onYes := outcome.SnippetDeleteRequested{
		Input:     snippet.DeleteInput{SnippetID: selected.ID()},
		Selection: s.selection(),
		Selecting: successor.ID(),
	}

	return outcome.Stay(s).Opening(confirm.New(s.keys, s.styles, snippetDeleteQuestion(selected), onYes))
}
