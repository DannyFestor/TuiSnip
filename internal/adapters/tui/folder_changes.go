package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/DannyFestor/TuiSnip/internal/adapters/tui/outcome"
	"github.com/DannyFestor/TuiSnip/internal/domain"
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
	created := reporting(
		actionFailure{operation: operationCreateFolder, whenGone: goneElsewhere(folderNoun, requested.ParentName)},
		func(folder domain.Folder) tea.Msg { return folderTreeChangedMsg{selecting: folder.ID(), status: ""} },
	)

	return func() tea.Msg { return created(m.folderCreator.Run(m.ctx, requested.Input)) }
}

func (m Model) deleteFolder(requested outcome.FolderDeleteRequested) tea.Cmd {
	deleted := reporting(
		actionFailure{operation: operationDeleteFolder, whenGone: goneElsewhere(folderNoun, requested.Name)},
		func(struct{}) tea.Msg {
			return folderTreeChangedMsg{
				selecting: requested.ParentID,
				status:    movedToTrashText(folderNoun, requested.Name),
			}
		},
	)

	return func() tea.Msg { return deleted(withoutResult(m.folderDeleter.Run(m.ctx, requested.Input))) }
}

func (m Model) moveFolder(requested outcome.FolderMoveRequested) tea.Cmd {
	moved := reporting(
		actionFailure{
			operation: operationMoveFolder,
			whenGone:  goneElsewhere(folderNoun, requested.Name).placingInto(requested.TargetName),
		},
		func(folder domain.Folder) tea.Msg { return folderTreeChangedMsg{selecting: folder.ID(), status: ""} },
	)

	return func() tea.Msg { return moved(m.folderMover.Run(m.ctx, requested.Input)) }
}

func (m Model) renameFolder(requested outcome.FolderRenameRequested) tea.Cmd {
	renamed := reporting(
		actionFailure{operation: operationRenameFolder, whenGone: goneElsewhere(folderNoun, requested.PreviousName)},
		func(domain.Folder) tea.Msg { return folderEditedMsg{} },
	)

	return func() tea.Msg { return renamed(m.folderRenamer.Run(m.ctx, requested.Input)) }
}

func (m Model) setFolderDefaultLanguage(requested outcome.DefaultLanguageRequested) tea.Cmd {
	set := reporting(
		actionFailure{operation: operationSetDefaultLanguage, whenGone: goneElsewhere(folderNoun, requested.Name)},
		func(domain.Folder) tea.Msg { return folderEditedMsg{} },
	)

	return func() tea.Msg { return set(m.defaultLanguageSetter.Run(m.ctx, requested.Input)) }
}
