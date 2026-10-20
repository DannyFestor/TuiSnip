package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
)

func (m Model) runFolderAction(reported outcome.Outcome) tea.Cmd {
	switch reported := reported.(type) {
	case outcome.FolderCreateRequested:
		return m.createFolder(reported)
	case outcome.FolderRenameRequested:
		return m.renameFolder(reported)
	case outcome.FolderDeleteRequested:
		return m.deleteFolder(reported)
	case outcome.DefaultLanguageRequested:
		return m.setFolderDefaultLanguage(reported)
	case outcome.FolderMoveRequested:
		return m.moveFolder(reported)
	default:
		return nil
	}
}

func (m Model) createFolder(requested outcome.FolderCreateRequested) tea.Cmd {
	return func() tea.Msg {
		created, err := m.folderCreator.Run(m.ctx, requested.Input)
		if err != nil {
			return itemActionFailedMsg{
				operation: operationCreateFolder, err: err, gone: goneItem(folderNoun, requested.ParentName),
			}
		}

		return folderTreeChangedMsg{selecting: created.ID(), status: ""}
	}
}

func (m Model) deleteFolder(requested outcome.FolderDeleteRequested) tea.Cmd {
	return func() tea.Msg {
		err := m.folderDeleter.Run(m.ctx, requested.Input)
		if err != nil {
			return itemActionFailedMsg{
				operation: operationDeleteFolder, err: err, gone: goneItem(folderNoun, requested.Name),
			}
		}

		return folderTreeChangedMsg{
			selecting: requested.ParentID,
			status:    movedToTrashText(folderNoun, requested.Name),
		}
	}
}

func (m Model) moveFolder(requested outcome.FolderMoveRequested) tea.Cmd {
	return func() tea.Msg {
		moved, err := m.folderMover.Run(m.ctx, requested.Input)
		if err != nil {
			return itemActionFailedMsg{
				operation: operationMoveFolder,
				err:       err,
				gone:      goneItem(folderNoun, requested.Name).placingInto(requested.TargetName),
			}
		}

		return folderTreeChangedMsg{selecting: moved.ID(), status: ""}
	}
}

func (m Model) renameFolder(requested outcome.FolderRenameRequested) tea.Cmd {
	return func() tea.Msg {
		_, err := m.folderRenamer.Run(m.ctx, requested.Input)
		if err != nil {
			return itemActionFailedMsg{
				operation: operationRenameFolder, err: err, gone: goneItem(folderNoun, requested.PreviousName),
			}
		}

		return folderEditedMsg{}
	}
}

func (m Model) setFolderDefaultLanguage(requested outcome.DefaultLanguageRequested) tea.Cmd {
	return func() tea.Msg {
		_, err := m.defaultLanguageSetter.Run(m.ctx, requested.Input)
		if err != nil {
			return itemActionFailedMsg{
				operation: operationSetDefaultLanguage, err: err, gone: goneItem(folderNoun, requested.Name),
			}
		}

		return folderEditedMsg{}
	}
}
