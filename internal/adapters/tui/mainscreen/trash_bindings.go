package mainscreen

import (
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/binding"
	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/app/trash"
	"github.com/DannyFestor/TuiSnip/internal/domain"
)

const (
	onlySnippetsCopyText = "Only a Snippet can be copied"
	restoreToEditText    = "Restore the Snippet to edit it"
)

func (s Screen) inTrash() bool {
	return s.selection().Trash()
}

func (s Screen) shown() panes {
	return s.panes.showingTrash(s.inTrash())
}

func (s Screen) trashPressed(msg tea.KeyPressMsg) (outcome.Step, bool) {
	item, ok := s.panes.trash.Selected()

	switch {
	case !ok:
		return outcome.Stay(s), false
	case s.trashKeys.Matches(msg, binding.Copy):
		return s.trashCopyAsked(item), true
	case s.trashKeys.Matches(msg, binding.Restore):
		return outcome.Stay(s).Passing(outcome.TrashRestoreRequested{
			Input: trash.RestoreInput{Item: item.ID()},
			Name:  item.Name(),
			Path:  item.Path,
		}), true
	}

	return outcome.Stay(s), false
}

func (s Screen) trashCopyAsked(item trash.Item) outcome.Step {
	if item.IsFolder() {
		return outcome.Stay(s.withStatus(onlySnippetsCopyText))
	}

	return s.copyAsked()
}

func (s Screen) trashEditPressed(msg tea.KeyPressMsg) (outcome.Step, bool) {
	editing := s.paneKeys.Matches(msg, binding.Edit) || s.paneKeys.Matches(msg, binding.OpenInEditor)
	if !s.inTrash() || s.focus != paneSnippet || !editing || s.panes.selectedTrashFolder() {
		return outcome.Stay(s), false
	}

	if _, ok := s.panes.trash.Selected(); !ok {
		return outcome.Stay(s), false
	}

	return outcome.Stay(s.withStatus(restoreToEditText)), true
}

func (s Screen) trashLoaded(items []trash.Item) Screen {
	if !s.inTrash() {
		return s
	}

	return s.withPanes(s.panes.withTrash(items))
}

func (s Screen) saved(saved outcome.SnippetSaved) outcome.Step {
	if !s.inTrash() {
		return outcome.Stay(s).Passing(saved)
	}

	next, expanded := s.selectingFolder(domain.FolderID{})

	return outcome.Stay(next).Passing(expanded...).Passing(saved)
}
