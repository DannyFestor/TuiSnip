package mainscreen

import (
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/snippet"
)

func (s Screen) listPressed(msg tea.KeyPressMsg) (outcome.Step, bool) {
	if s.focus != paneList {
		return outcome.Stay(s), false
	}

	if s.inTrash() {
		return s.trashPressed(msg)
	}

	switch {
	case s.listKeys.Matches(msg, binding.Copy):
		return s.copyAsked(), true
	case s.listKeys.Matches(msg, binding.CycleSort):
		return outcome.Stay(s).Passing(s.panes.sortCycleAsked(s.selection())), true
	case s.listKeys.Matches(msg, binding.Duplicate):
		return s.duplicateAsked(), true
	case s.listKeys.Matches(msg, binding.Delete):
		return s.deleteAsked(), true
	}

	return outcome.Stay(s), false
}

func (s Screen) copyAsked() outcome.Step {
	request, ok := s.panes.preview.Copying()
	if !ok {
		return outcome.Stay(s)
	}

	return outcome.Stay(s).Passing(request)
}

func (s Screen) duplicateAsked() outcome.Step {
	selected, ok := s.panes.list.Selected()
	if !ok {
		return outcome.Stay(s)
	}

	return outcome.Stay(s).Passing(outcome.SnippetDuplicateRequested{
		Input:     snippet.DuplicateInput{SnippetID: selected.ID()},
		Title:     selected.Title().String(),
		Selection: s.selection(),
	})
}

func (s Screen) deleteAsked() outcome.Step {
	selected, ok := s.panes.list.Selected()
	if !ok {
		return outcome.Stay(s)
	}

	successor, _ := s.panes.list.Successor()

	return outcome.Stay(s).Passing(outcome.SnippetDeleteRequested{
		Input:     snippet.DeleteInput{SnippetID: selected.ID()},
		Title:     selected.Title().String(),
		Selection: s.selection(),
		Selecting: successor.ID(),
	})
}
