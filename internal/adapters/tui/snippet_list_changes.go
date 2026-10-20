package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
)

func (m Model) runSnippetListChange(reported outcome.Outcome) tea.Cmd {
	switch reported := reported.(type) {
	case outcome.SnippetDuplicateRequested:
		return m.duplicateSnippet(reported)
	case outcome.SnippetDeleteRequested:
		return m.deleteSnippet(reported)
	case outcome.SnippetMoveRequested:
		return m.moveSnippet(reported)
	case outcome.TrashRestoreRequested:
		return m.restoreFromTrash(reported)
	default:
		return nil
	}
}

func (m Model) duplicateSnippet(requested outcome.SnippetDuplicateRequested) tea.Cmd {
	return func() tea.Msg {
		duplicate, err := m.duplicator.Run(m.ctx, requested.Input)
		if err != nil {
			return itemActionFailedMsg{
				operation: operationDuplicate, err: err, gone: goneItem(snippetNoun, requested.Title),
			}
		}

		return snippetsChangedMsg{selection: requested.Selection, selecting: duplicate.ID(), status: ""}
	}
}

func (m Model) deleteSnippet(requested outcome.SnippetDeleteRequested) tea.Cmd {
	return func() tea.Msg {
		err := m.deleter.Run(m.ctx, requested.Input)
		if err != nil {
			return itemActionFailedMsg{
				operation: operationDeleteSnippet, err: err, gone: goneItem(snippetNoun, requested.Title),
			}
		}

		return snippetsChangedMsg{
			selection: requested.Selection,
			selecting: requested.Selecting,
			status:    movedToTrashText(snippetNoun, requested.Title),
		}
	}
}

func (m Model) moveSnippet(requested outcome.SnippetMoveRequested) tea.Cmd {
	return func() tea.Msg {
		_, err := m.mover.Run(m.ctx, requested.Input)
		if err != nil {
			return itemActionFailedMsg{
				operation: operationMoveSnippet,
				err:       err,
				gone:      goneItem(snippetNoun, requested.Title).placingInto(requested.TargetName),
			}
		}

		return snippetsChangedMsg{selection: requested.Selection, selecting: requested.Selecting, status: ""}
	}
}
