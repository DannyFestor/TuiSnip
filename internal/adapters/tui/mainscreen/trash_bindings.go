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

func (s Screen) trashPressed(msg tea.KeyPressMsg) (outcome.Step, bool) {
	trashItem, ok := s.panes.trash.Selected()

	switch {
	case !ok:
		return outcome.Stay(s), false
	case s.focus == paneList:
		return s.trashListPressed(trashItem, msg)
	case s.focus == paneSnippet:
		return s.trashPanePressed(trashItem, msg)
	}

	return outcome.Stay(s), false
}

func (s Screen) savedFromTrash(saved outcome.SnippetSaved) outcome.Step {
	next, expanded := s.selectingFolder(domain.FolderID{})

	return outcome.Stay(next).Passing(expanded...).Passing(saved)
}

func (s Screen) trashListPressed(trashItem trash.Item, msg tea.KeyPressMsg) (outcome.Step, bool) {
	switch {
	case s.trashKeys.Matches(msg, binding.Copy):
		return s.trashCopyAsked(trashItem), true
	case s.trashKeys.Matches(msg, binding.Restore):
		return outcome.Stay(s).Passing(outcome.TrashRestoreRequested{
			Input: trash.RestoreInput{Item: trashItem.ID()},
			Name:  trashItem.Name(),
			Path:  trashItem.Path,
		}), true
	case s.trashKeys.Matches(msg, binding.Delete):
		return s.asking(deletePermanentlyQuestion(trashItem), outcome.TrashDeleteRequested{
			Input: trash.DeletePermanentlyInput{Item: trashItem.ID()},
			Name:  trashItem.Name(),
		}), true
	case s.trashKeys.Matches(msg, binding.Empty):
		return s.emptyTrashAsked(), true
	}

	return outcome.Stay(s), false
}

func (s Screen) emptyTrashAsked() outcome.Step {
	return s.asking(
		emptyTrashQuestion(s.panes.trash.TrashItems()),
		outcome.TrashEmptyRequested{Input: trash.EmptyInput{Items: s.panes.trash.TrashItemIDs()}},
	)
}

func (s Screen) trashPanePressed(trashItem trash.Item, msg tea.KeyPressMsg) (outcome.Step, bool) {
	editing := s.paneKeys.Matches(msg, binding.Edit) || s.paneKeys.Matches(msg, binding.OpenInEditor)

	switch {
	case s.paneKeys.Matches(msg, binding.Copy):
		return s.trashCopyAsked(trashItem), true
	case trashItem.Kind == trash.ItemKindSnippet && editing:
		return outcome.Stay(s.withStatus(restoreToEditText)), true
	}

	return outcome.Stay(s), false
}

func (s Screen) trashCopyAsked(trashItem trash.Item) outcome.Step {
	if trashItem.Kind.IsFolder() {
		return outcome.Stay(s.withStatus(onlySnippetsCopyText))
	}

	request, ok := s.panes.preview.Copying()
	if !ok {
		return outcome.Stay(s)
	}

	return outcome.Stay(s).Passing(outcome.TrashCopyRequested(request))
}
